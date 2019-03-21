package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/containerd/cgroups"
	"github.com/containerd/containerd"
	tasks "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/diff"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/rootfs"
	"github.com/containerd/containerd/runtime/v2/runc/options"
	"github.com/containerd/containerd/snapshots"
	gocni "github.com/containerd/go-cni"
	"github.com/containerd/typeurl"
	"github.com/gogo/protobuf/types"
	ver "github.com/opencontainers/image-spec/specs-go"
	is "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/siddontang/ledisdb/server"
	"github.com/sirupsen/logrus"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/cni"
	"github.com/stellarproject/orbit/config"
	"github.com/stellarproject/orbit/flux"
	"github.com/stellarproject/orbit/opts"
	"github.com/stellarproject/orbit/util"
	"golang.org/x/sys/unix"
)

var (
	ErrNoID                  = errors.New("no id provided")
	ErrNoRef                 = errors.New("no ref provided")
	errServiceExistsOnTarget = errors.New("service exists on target")
	errMediaTypeNotFound     = errors.New("media type not found in index")
	errUnableToSignal        = errors.New("unable to signal task")

	plainRemotes = make(map[string]bool)

	empty = &types.Empty{}
)

type network interface {
	Create(context.Context, containerd.Container) (string, error)
	Remove(context.Context, containerd.Container) error
}

const (
	MediaTypeContainerInfo = "application/vnd.orbit.container.info.v1+json"
	StatusLabel            = "stellarproject.io/orbit/restart.status"
)

func New(ctx context.Context, c *Config, client *containerd.Client) (*Agent, error) {
	if err := setupApparmor(); err != nil {
		return nil, err
	}
	s, err := newStore(c.Root)
	if err != nil {
		return nil, err
	}
	for _, r := range c.PlainRemotes {
		plainRemotes[r] = true
	}
	a := &Agent{
		config: c,
		client: client,
		store:  newStoreClient(),
		server: s,
	}
	go a.startSupervisorLoop(namespaces.WithNamespace(ctx, config.DefaultNamespace), c.Interval.Duration)
	return a, nil
}

type Agent struct {
	client       *containerd.Client
	config       *Config
	store        *store
	server       *server.App
	supervisorMu sync.Mutex
}

func (a *Agent) Close() error {
	a.supervisorMu.Lock()

	a.store.Close()
	a.server.Close()
	return nil
}

func (a *Agent) Create(ctx context.Context, req *v1.CreateRequest) (*types.Empty, error) {
	ctx = relayContext(ctx)
	image, err := a.client.Pull(ctx, req.Container.Image, containerd.WithPullUnpack, withPlainRemote(req.Container.Image))
	if err != nil {
		return nil, err
	}
	if _, err := a.client.LoadContainer(ctx, req.Container.ID); err == nil {
		if !req.Update {
			return nil, errors.Errorf("container %s already exists", req.Container.ID)
		}
		_, err = a.Update(ctx, &v1.UpdateRequest{
			Container: req.Container,
		})
		return empty, err
	}
	container, err := a.client.NewContainer(ctx,
		req.Container.ID,
		flux.WithNewSnapshot(image),
		opts.WithOrbitConfig(a.config.Paths(req.Container.ID), req.Container, image),
	)
	if err != nil {
		return nil, err
	}
	/*
		if err := a.store.Write(ctx, req.Container); err != nil {
			container.Delete(ctx, containerd.WithSnapshotCleanup)
			return nil, err
		}
	*/
	if err := a.start(ctx, container); err != nil {
		return nil, err
	}
	return empty, nil
}

func (a *Agent) Delete(ctx context.Context, req *v1.DeleteRequest) (*types.Empty, error) {
	ctx = relayContext(ctx)
	id := req.ID
	if id == "" {
		return nil, ErrNoID
	}
	container, err := a.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "load container")
	}
	if err := a.stop(ctx, container); err != nil {
		return nil, errors.Wrap(err, "stop container")
	}
	config, err := opts.GetConfig(ctx, container)
	if err != nil {
		return nil, errors.Wrap(err, "load config")
	}
	network, err := a.getNetwork(config.Network)
	if err != nil {
		return nil, errors.Wrap(err, "get network")
	}
	if err := network.Remove(ctx, container); err != nil {
		return nil, err
	}
	return empty, container.Delete(ctx, flux.WithRevisionCleanup)
}

func (a *Agent) Get(ctx context.Context, req *v1.GetRequest) (*v1.GetResponse, error) {
	ctx = relayContext(ctx)
	id := req.ID
	if id == "" {
		return nil, ErrNoID
	}
	container, err := a.client.LoadContainer(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	i, err := a.info(ctx, container)
	if err != nil {
		return nil, err
	}
	return &v1.GetResponse{
		Container: i,
	}, nil
}

func (a *Agent) info(ctx context.Context, c containerd.Container) (*v1.ContainerInfo, error) {
	info, err := c.Info(ctx)
	if err != nil {
		return nil, err
	}
	d := info.Extensions[opts.CurrentConfig]
	cfg, err := opts.UnmarshalConfig(&d)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config")
	}

	service := a.client.SnapshotService(info.Snapshotter)
	usage, err := service.Usage(ctx, info.SnapshotKey)
	if err != nil {
		return nil, err
	}
	var ss []*v1.Snapshot
	if err := service.Walk(ctx, func(ctx context.Context, si snapshots.Info) error {
		if si.Labels[flux.ContainerIDLabel] != c.ID() {
			return nil
		}
		usage, err := service.Usage(ctx, si.Name)
		if err != nil {
			return err
		}
		ss = append(ss, &v1.Snapshot{
			ID:       si.Name,
			Created:  si.Created,
			Previous: si.Labels[flux.PreviousLabel],
			FsSize:   usage.Size,
		})
		return nil
	}); err != nil {
		return nil, err
	}
	bindSizes, err := getBindSizes(cfg)
	if err != nil {
		return nil, err
	}
	task, err := c.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return &v1.ContainerInfo{
				ID:        c.ID(),
				Image:     info.Image,
				Status:    string(containerd.Stopped),
				FsSize:    usage.Size + bindSizes,
				Config:    cfg,
				Snapshots: ss,
			}, nil
		}
		return nil, err
	}
	status, err := task.Status(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := task.Metrics(ctx)
	if err != nil {
		return nil, err
	}
	v, err := typeurl.UnmarshalAny(stats.Data)
	if err != nil {
		return nil, err
	}
	var (
		cg     = v.(*cgroups.Metrics)
		cpu    = cg.CPU.Usage.Total
		memory = float64(cg.Memory.Usage.Usage - cg.Memory.TotalCache)
		limit  = float64(cg.Memory.Usage.Limit)
	)
	var services []*v1.Service
	for _, s := range cfg.Services {
		services = append(services, &v1.Service{
			Name:  s.Name,
			IP:    info.Labels[opts.IPLabel],
			Port:  s.Port,
			Url:   s.Url,
			Check: s.Check,
		})
	}
	return &v1.ContainerInfo{
		ID:          c.ID(),
		Image:       info.Image,
		Status:      string(status.Status),
		Services:    services,
		Cpu:         cpu,
		MemoryUsage: memory,
		MemoryLimit: limit,
		PidUsage:    cg.Pids.Current,
		PidLimit:    cg.Pids.Limit,
		FsSize:      usage.Size + bindSizes,
		Config:      cfg,
		Snapshots:   ss,
	}, nil
}

func (a *Agent) List(ctx context.Context, req *v1.ListRequest) (*v1.ListResponse, error) {
	var resp v1.ListResponse
	ctx = relayContext(ctx)
	containers, err := a.client.Containers(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		l, err := a.info(ctx, c)
		if err != nil {
			resp.Containers = append(resp.Containers, &v1.ContainerInfo{
				ID:     c.ID(),
				Status: err.Error(),
			})
			logrus.WithError(err).Error("info container")
			continue
		}
		resp.Containers = append(resp.Containers, l)
	}
	return &resp, nil
}

func (a *Agent) Kill(ctx context.Context, req *v1.KillRequest) (*types.Empty, error) {
	ctx = relayContext(ctx)
	id := req.ID
	if id == "" {
		return nil, ErrNoID
	}
	container, err := a.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, err
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return nil, err
	}
	if err := task.Kill(ctx, unix.SIGTERM); err != nil {
		return nil, err
	}
	return empty, nil
}

func (a *Agent) Start(ctx context.Context, req *v1.StartRequest) (*types.Empty, error) {
	ctx = relayContext(ctx)
	id := req.ID
	if id == "" {
		return nil, ErrNoID
	}
	ctx, done, err := a.client.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)
	container, err := a.client.LoadContainer(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return empty, a.start(ctx, container)
}

func (a *Agent) Stop(ctx context.Context, req *v1.StopRequest) (*types.Empty, error) {
	ctx = relayContext(ctx)
	id := req.ID
	if id == "" {
		return nil, ErrNoID
	}
	ctx, done, err := a.client.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)
	container, err := a.client.LoadContainer(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return empty, a.stop(ctx, container)
}

func (a *Agent) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	ctx = relayContext(ctx)
	ctx, done, err := a.client.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)
	container, err := a.client.LoadContainer(ctx, req.Container.ID)
	if err != nil {
		return nil, err
	}
	current, err := opts.GetConfig(ctx, container)
	if err != nil {
		return nil, err
	}
	var changes []change
	for _, cs := range current.Services {
	inner:
		for _, ns := range req.Container.Services {
			if cs.Name == ns.Name {
				// if the new config does not have a service, deregister the old one
				changes = append(changes, &deregisterChange{
					store: a.store,
					name:  ns.Name,
				})
				break inner
			}
		}
	}
	changes = append(changes, &imageUpdateChange{
		ref:    req.Container.Image,
		client: a.client,
	})
	changes = append(changes, &configChange{
		client: a.client,
		c:      req.Container,
		config: a.config,
	})
	changes = append(changes, &filesChange{
		c: req.Container,
		// store: a.store,
	})

	var wait <-chan containerd.ExitStatus
	// bump the task to pickup the changes
	task, err := container.Task(ctx, nil)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
	}
	if task != nil {
		if wait, err = task.Wait(ctx); err != nil {
			return nil, err
		}
	} else {
		c := make(chan containerd.ExitStatus)
		wait = c
		close(c)
	}
	err = pauseAndRun(ctx, container, func() error {
		for _, ch := range changes {
			if err := ch.update(ctx, container); err != nil {
				return err
			}
		}
		if task == nil {
			return nil
		}
		return task.Kill(ctx, unix.SIGTERM)
	})
	if err != nil {
		return nil, err
	}
	wctx, _ := context.WithTimeout(ctx, 10*time.Second)
	for {
		select {
		case <-wctx.Done():
			if task != nil {
				return &v1.UpdateResponse{}, task.Kill(ctx, unix.SIGKILL)
			}
			return nil, wctx.Err()
		case <-wait:
			return &v1.UpdateResponse{}, nil
		}
	}
	return &v1.UpdateResponse{}, nil
}

func (a *Agent) Rollback(ctx context.Context, req *v1.RollbackRequest) (*v1.RollbackResponse, error) {
	ctx = relayContext(ctx)
	if req.ID == "" {
		return nil, ErrNoID
	}
	ctx, done, err := a.client.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)
	container, err := a.client.LoadContainer(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	err = pauseAndRun(ctx, container, func() error {
		if err := container.Update(ctx, flux.WithRollback, opts.WithRollback); err != nil {
			return err
		}
		task, err := container.Task(ctx, nil)
		if err != nil {
			return err
		}
		return task.Kill(ctx, unix.SIGTERM)
	})
	if err != nil {
		return nil, err
	}
	return &v1.RollbackResponse{}, nil
}

func (a *Agent) Push(ctx context.Context, req *v1.PushRequest) (*types.Empty, error) {
	ctx = relayContext(ctx)
	if req.Ref == "" {
		return nil, ErrNoRef
	}
	image, err := a.client.GetImage(ctx, req.Ref)
	if err != nil {
		return nil, err
	}
	return empty, a.client.Push(ctx, req.Ref, image.Target(), withPlainRemote(req.Ref))
}

func (a *Agent) Checkpoint(ctx context.Context, req *v1.CheckpointRequest) (*v1.CheckpointResponse, error) {
	ctx = relayContext(ctx)
	if req.ID == "" {
		return nil, ErrNoID
	}
	ctx, done, err := a.client.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)
	container, err := a.client.LoadContainer(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	info, err := container.Info(ctx)
	if err != nil {
		return nil, err
	}
	index := is.Index{
		Versioned: ver.Versioned{
			SchemaVersion: 2,
		},
		Annotations: make(map[string]string),
	}
	data, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	desc, err := writeContent(ctx, a.client.ContentStore(), MediaTypeContainerInfo, req.ID+"-container-info", r)
	if err != nil {
		return nil, err
	}
	desc.Platform = &is.Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
	index.Manifests = append(index.Manifests, desc)

	opts := options.CheckpointOptions{
		Exit:                req.Exit,
		OpenTcp:             false,
		ExternalUnixSockets: false,
		Terminal:            false,
		FileLocks:           true,
		EmptyNamespaces:     nil,
	}
	any, err := typeurl.MarshalAny(&opts)
	if err != nil {
		return nil, err
	}
	err = pauseAndRun(ctx, container, func() error {
		// checkpoint rw layer
		opts := []diff.Opt{
			diff.WithReference(fmt.Sprintf("checkpoint-rw-%s", info.SnapshotKey)),
			diff.WithMediaType(is.MediaTypeImageLayer),
		}
		rw, err := rootfs.CreateDiff(ctx,
			info.SnapshotKey,
			a.client.SnapshotService(info.Snapshotter),
			a.client.DiffService(),
			opts...,
		)
		if err != nil {
			return err
		}
		rw.Platform = &is.Platform{
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
		}
		index.Manifests = append(index.Manifests, rw)
		if req.Live {
			task, err := a.client.TaskService().Checkpoint(ctx, &tasks.CheckpointTaskRequest{
				ContainerID: req.ID,
				Options:     any,
			})
			if err != nil {
				return err
			}
			for _, d := range task.Descriptors {
				if d.MediaType == images.MediaTypeContainerd1CheckpointConfig {
					// we will save the entire container config to the checkpoint instead
					continue
				}
				index.Manifests = append(index.Manifests, is.Descriptor{
					MediaType: d.MediaType,
					Size:      d.Size_,
					Digest:    d.Digest,
					Platform: &is.Platform{
						OS:           runtime.GOOS,
						Architecture: runtime.GOARCH,
					},
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if desc, err = a.writeIndex(ctx, &index, req.ID+"index"); err != nil {
		return nil, err
	}
	i := images.Image{
		Name:   req.Ref,
		Target: desc,
	}
	if _, err := a.client.ImageService().Create(ctx, i); err != nil {
		return nil, err
	}
	if req.Exit {
		if err := a.stop(ctx, container); err != nil {
			return nil, errors.Wrap(err, "stop service")
		}
	}
	return &v1.CheckpointResponse{}, nil
}

func (a *Agent) Restore(ctx context.Context, req *v1.RestoreRequest) (*v1.RestoreResponse, error) {
	ctx = relayContext(ctx)
	if req.Ref == "" {
		return nil, ErrNoRef
	}
	checkpoint, err := a.client.GetImage(ctx, req.Ref)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
		ck, err := a.client.Fetch(ctx, req.Ref, withPlainRemote(req.Ref))
		if err != nil {
			return nil, err
		}
		checkpoint = containerd.NewImage(a.client, ck)
	}
	store := a.client.ContentStore()
	index, err := decodeIndex(ctx, store, checkpoint.Target())
	if err != nil {
		return nil, err
	}
	configDesc, err := getByMediaType(index, MediaTypeContainerInfo)
	if err != nil {
		return nil, err
	}
	data, err := content.ReadBlob(ctx, store, *configDesc)
	if err != nil {
		return nil, err
	}
	var c containers.Container
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	config, err := opts.GetConfigFromInfo(ctx, c)
	if err != nil {
		return nil, err
	}
	image, err := a.client.Pull(ctx, config.Image, containerd.WithPullUnpack, withPlainRemote(config.Image))
	if err != nil {
		return nil, err
	}
	o := []containerd.NewContainerOpts{
		flux.WithNewSnapshot(image),
		opts.WithOrbitConfig(a.config.Paths(c.ID), config, image),
	}
	if req.Live {
		desc, err := getByMediaType(index, images.MediaTypeContainerd1Checkpoint)
		if err != nil {
			return nil, err
		}
		o = append(o, opts.WithRestore(desc))
	}
	container, err := a.client.NewContainer(ctx,
		config.ID,
		o...,
	)
	if err != nil {
		return nil, err
	}
	// apply rw layer
	info, err := container.Info(ctx)
	if err != nil {
		return nil, err
	}
	rw, err := getByMediaType(index, is.MediaTypeImageLayerGzip)
	if err != nil {
		return nil, err
	}
	mounts, err := a.client.SnapshotService(info.Snapshotter).Mounts(ctx, info.SnapshotKey)
	if err != nil {
		return nil, err
	}
	if _, err := a.client.DiffService().Apply(ctx, *rw, mounts); err != nil {
		return nil, err
	}
	/*
		if err := a.store.Write(ctx, config); err != nil {
			container.Delete(ctx, containerd.WithSnapshotCleanup)
			return nil, err
		}
	*/
	if err := a.start(ctx, container); err != nil {
		return nil, err
	}
	return &v1.RestoreResponse{}, nil
}

func (a *Agent) Migrate(ctx context.Context, req *v1.MigrateRequest) (*v1.MigrateResponse, error) {
	ctx = relayContext(ctx)
	if req.ID == "" {
		return nil, ErrNoID
	}
	to, err := util.Agent(req.To)
	if err != nil {
		return nil, err
	}
	defer to.Close()
	if _, err := to.Get(ctx, &v1.GetRequest{
		ID: req.ID,
	}); err == nil {
		return nil, errServiceExistsOnTarget
	}
	if _, err := a.Checkpoint(ctx, &v1.CheckpointRequest{
		ID:   req.ID,
		Live: req.Live,
		Ref:  req.Ref,
		Exit: req.Stop || req.Delete,
	}); err != nil {
		return nil, err
	}
	defer a.client.ImageService().Delete(ctx, req.Ref)
	if _, err := a.Push(ctx, &v1.PushRequest{
		Ref: req.Ref,
	}); err != nil {
		return nil, err
	}
	if _, err := to.Restore(ctx, &v1.RestoreRequest{
		Ref:  req.Ref,
		Live: req.Live,
	}); err != nil {
		return nil, err
	}
	if req.Delete {
		if _, err := a.Delete(ctx, &v1.DeleteRequest{
			ID: req.ID,
		}); err != nil {
			return nil, err
		}
	}
	return &v1.MigrateResponse{}, nil
}

func (a *Agent) writeIndex(ctx context.Context, index *is.Index, ref string) (d is.Descriptor, err error) {
	labels := map[string]string{}
	for i, m := range index.Manifests {
		labels[fmt.Sprintf("containerd.io/gc.ref.content.%d", i)] = m.Digest.String()
	}
	data, err := json.Marshal(index)
	if err != nil {
		return is.Descriptor{}, err
	}
	return writeContent(ctx, a.client.ContentStore(), is.MediaTypeImageIndex, ref, bytes.NewReader(data), content.WithLabels(labels))
}

func (a *Agent) startSupervisorLoop(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			if err := ctx.Err(); err != nil {
				logrus.WithError(err).Error("context done in supervisor loop")
			}
			logrus.Info("exiting supervisor loop")
			return
		case <-ticker.C:
			a.supervisorMu.Lock()
			if err := a.reconcile(ctx); err != nil {
				logrus.WithError(err).Error("reconcile loop")
			}
			a.supervisorMu.Unlock()
		}
	}
}

func (a *Agent) reconcile(ctx context.Context) error {
	containers, err := a.client.Containers(ctx, fmt.Sprintf("labels.%q", StatusLabel))
	if err != nil {
		return err
	}
	var dd []stateChange
	for _, c := range containers {
		d, err := a.getContainerDiff(ctx, c)
		if err != nil {
			logrus.WithError(err).WithField("id", c.ID()).Error("unable to generate supervisor diff")
			continue
		}
		dd = append(dd, d)
	}
	for _, d := range dd {
		if err := d.apply(ctx); err != nil {
			logrus.WithError(err).Error("unable to apply state change")
		}
	}
	return nil
}

func (a *Agent) start(ctx context.Context, container containerd.Container) error {
	logrus.WithField("id", container.ID()).Debug("starting container")
	a.supervisorMu.Lock()
	defer a.supervisorMu.Unlock()

	config, err := opts.GetConfig(ctx, container)
	if err != nil {
		return err
	}
	desc, err := opts.GetRestoreDesc(ctx, container)
	if err != nil {
		return err
	}
	if err := a.setupRuntimeFiles(ctx, container); err != nil {
		return err
	}
	network, err := a.getNetwork(config.Network)
	if err != nil {
		return err
	}
	ip, err := network.Create(ctx, container)
	if err != nil {
		return err
	}
	if ip != "" {
		logrus.WithField("id", container.ID()).WithField("ip", ip).Debug("setup network interface")
		for _, srv := range config.Services {
			service := &v1.Service{
				Name:  srv.Name,
				IP:    ip,
				Port:  srv.Port,
				Url:   srv.Url,
				Check: srv.Check,
			}
			if err := a.store.Register(container.ID(), service); err != nil {
				return err
			}
		}
	}
	if err := container.Update(ctx, opts.WithIP(ip), opts.WithoutRestore, withStatus(containerd.Running)); err != nil {
		return err
	}
	task, err := container.NewTask(ctx, cio.BinaryIO("orbit-log", nil), opts.WithTaskRestore(desc))
	if err != nil {
		return err
	}
	return task.Start(ctx)
}

func (a *Agent) stop(ctx context.Context, container containerd.Container) error {
	logrus.WithField("id", container.ID()).Debug("stopping container")
	a.supervisorMu.Lock()
	defer a.supervisorMu.Unlock()

	if err := container.Update(ctx, withStatus(containerd.Stopped)); err != nil {
		return err
	}
	task, err := container.Task(ctx, nil)
	if err == nil {
		wait, err := task.Wait(ctx)
		if err != nil {
			if _, derr := task.Delete(ctx); derr == nil {
				return nil
			}
			return err
		}
		if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
			if _, derr := task.Delete(ctx); derr == nil {
				return nil
			}
			return err
		}
		<-wait
		if _, err := task.Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) setupRuntimeFiles(ctx context.Context, container containerd.Container) error {
	return a.setupResolvConf(container.ID())
}

func (a *Agent) nameservers() []string {
	ns := a.config.Nameservers
	if len(ns) == 0 {
		ns = []string{
			"8.8.8.8",
			"8.8.4.4",
		}
	}
	return ns
}

func (a *Agent) setupResolvConf(id string) error {
	if err := os.MkdirAll(filepath.Join(a.config.Root, id), 0711); err != nil {
		return err
	}
	f, err := ioutil.TempFile("", "orbit-resolvconf")
	if err != nil {
		return err
	}
	if err := f.Chmod(0666); err != nil {
		return err
	}
	for _, ns := range a.nameservers() {
		if _, err := f.WriteString(fmt.Sprintf("nameserver %s\n", ns)); err != nil {
			f.Close()
			return err
		}
	}
	f.Close()
	return os.Rename(f.Name(), filepath.Join(a.config.Root, id, "resolv.conf"))
}

func (a *Agent) getContainerDiff(ctx context.Context, container containerd.Container) (stateChange, error) {
	labels, err := container.Labels(ctx)
	if err != nil {
		return nil, err
	}
	desiredStatus := containerd.ProcessStatus(labels[StatusLabel])
	if !isSameStatus(ctx, desiredStatus, container) {
		switch desiredStatus {
		case containerd.Running:
			return &startDiff{
				a:         a,
				container: container,
			}, nil
		case containerd.Stopped:
			return &stopDiff{
				a:         a,
				container: container,
			}, nil
		}
	}
	return sameDiff(), nil
}

func (a *Agent) getNetwork(network *types.Any) (network, error) {
	if network == nil {
		return &none{}, nil
	}
	v, err := typeurl.UnmarshalAny(network)
	if err != nil {
		return nil, err
	}

	var networkType string
	switch c := v.(type) {
	case *v1.HostNetwork:
		networkType = "host"
		ip, err := util.GetIP(a.config.Iface)
		if err != nil {
			return nil, err
		}
		return &host{
			ip: ip,
		}, nil
	case *v1.CNINetwork:
		networkType = c.Type
		if c.Type == "macvlan" && a.config.BridgeAddress == "" {
			return nil, errors.New("bridge_address must be specified with macvlan")
		}
		if c.Name == "" {
			c.Name = a.config.Domain
		}
		if c.Master == "" {
			c.Master = a.config.Iface
		}
		n, err := gocni.New(
			gocni.WithPluginDir([]string{"/usr/local/bin", "/opt/containerd/bin"}),
			gocni.WithConf(c.MarshalCNI()),
			gocni.WithLoNetwork,
		)
		if err != nil {
			return nil, err
		}
		return cni.New(cni.Config{
			Type:       networkType,
			State:      a.config.State,
			Iface:      a.config.Iface,
			BridgeAddr: a.config.BridgeAddress,
		}, n)
	default:
		return nil, errors.Errorf("unknown network type %s", network.TypeUrl)
	}
}

func withStatus(status containerd.ProcessStatus) func(context.Context, *containerd.Client, *containers.Container) error {
	return func(_ context.Context, _ *containerd.Client, c *containers.Container) error {
		ensureLabels(c)
		c.Labels[StatusLabel] = string(status)
		return nil
	}
}

func ensureLabels(c *containers.Container) {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
}

func isSameStatus(ctx context.Context, desired containerd.ProcessStatus, container containerd.Container) bool {
	task, err := container.Task(ctx, nil)
	if err != nil {
		return desired == containerd.Stopped
	}
	state, err := task.Status(ctx)
	if err != nil {
		return desired == containerd.Stopped
	}
	return desired == state.Status
}

func writeContent(ctx context.Context, store content.Ingester, mediaType, ref string, r io.Reader, opts ...content.Opt) (d is.Descriptor, err error) {
	writer, err := store.Writer(ctx, content.WithRef(ref))
	if err != nil {
		return d, err
	}
	defer writer.Close()
	size, err := io.Copy(writer, r)
	if err != nil {
		return d, err
	}
	if err := writer.Commit(ctx, size, "", opts...); err != nil {
		return d, err
	}
	return is.Descriptor{
		MediaType: mediaType,
		Digest:    writer.Digest(),
		Size:      size,
	}, nil
}

func decodeIndex(ctx context.Context, store content.Provider, desc is.Descriptor) (*is.Index, error) {
	var index is.Index
	p, err := content.ReadBlob(ctx, store, desc)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(p, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

func relayContext(ctx context.Context) context.Context {
	return namespaces.WithNamespace(ctx, config.DefaultNamespace)
}

func getByMediaType(index *is.Index, mt string) (*is.Descriptor, error) {
	for _, d := range index.Manifests {
		if d.MediaType == mt {
			return &d, nil
		}
	}
	return nil, errMediaTypeNotFound
}

func withPlainRemote(ref string) containerd.RemoteOpt {
	remote := strings.SplitN(ref, "/", 2)[0]
	return func(_ *containerd.Client, ctx *containerd.RemoteContext) error {
		ctx.Resolver = docker.NewResolver(docker.ResolverOptions{
			PlainHTTP: plainRemotes[remote],
			Client:    http.DefaultClient,
		})
		return nil
	}
}

func getBindSizes(c *v1.Container) (size int64, _ error) {
	for _, m := range c.Mounts {
		f, err := os.Open(m.Source)
		if err != nil {
			logrus.WithError(err).Warnf("unable to open bind for size %s", m.Source)
			continue
		}
		info, err := f.Stat()
		if err != nil {
			f.Close()
			return size, err
		}
		if info.IsDir() {
			f.Close()
			if err := filepath.Walk(m.Source, func(path string, wi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if wi.IsDir() {
					return nil
				}
				size += wi.Size()
				return nil
			}); err != nil {
				return size, err
			}
			continue
		}
		size += info.Size()
		f.Close()
	}
	return size, nil
}
