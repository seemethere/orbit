package agent

import (
	"os"
	"strconv"
	"time"

	"github.com/containerd/containerd"
	containers "github.com/containerd/containerd/api/services/containers/v1"
	diff "github.com/containerd/containerd/api/services/diff/v1"
	images "github.com/containerd/containerd/api/services/images/v1"
	namespaces "github.com/containerd/containerd/api/services/namespaces/v1"
	tasks "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/services"
	"github.com/containerd/containerd/snapshots"
	is "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/config"
	"github.com/stellarproject/orbit/util"
	"github.com/stellarproject/orbit/version"
	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"
)

const defaultRuntime = "io.containerd.runc.v2"

// init func for a containerd service plugin
func init() {
	plugin.Register(&plugin.Registration{
		Type:   plugin.GRPCPlugin,
		ID:     "stellarproject.io/orbit",
		Config: &Config{},
		Requires: []plugin.Type{
			plugin.ServicePlugin,
		},
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			ic.Meta.Platforms = []is.Platform{platforms.DefaultSpec()}
			c := ic.Config.(*Config)
			if c.Iface == "" {
				i, err := getDefaultIface()
				if err != nil {
					return nil, err
				}
				c.Iface = i
			}
			if c.Domain == "" {
				d, err := util.GetDomainName()
				if err != nil {
					return nil, err
				}
				c.Domain = d
			}
			if c.ID == "" {
				h, err := os.Hostname()
				if err != nil {
					return nil, err
				}
				c.ID = h
			}

			// set root and state from containerd plugins
			c.Root = ic.Root
			c.State = ic.State

			server, err := newStore(c.Root, c.Master)
			if err != nil {
				return nil, err
			}

			// set orbit information
			exports := make(map[string]string)
			exports["interface"] = c.Iface
			exports["domain"] = c.Domain
			exports["store"] = server.Address()
			exports["version"] = version.Version
			exports["master"] = strconv.FormatBool(c.Master)
			ic.Meta.Exports = exports

			client, err := getClient(ic)
			if err != nil {
				return nil, err
			}
			return New(ic.Context, c, client, server)
		},
	})
}

func getClient(ic *plugin.InitContext) (*containerd.Client, error) {
	servicesOpts, err := getServicesOpts(ic)
	if err != nil {
		return nil, err
	}
	return containerd.New(
		"",
		containerd.WithDefaultNamespace(config.DefaultNamespace),
		containerd.WithServices(servicesOpts...),
		containerd.WithDefaultRuntime(defaultRuntime),
	)
}

func getDefaultIface() (string, error) {
	for i := 0; i < 20; i++ {
		routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
		if err != nil {
			return "", err
		}
		for _, r := range routes {
			if r.Gw != nil {
				link, err := netlink.LinkByIndex(r.LinkIndex)
				if err != nil {
					return "", err
				}
				return link.Attrs().Name, nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return "", errors.New("no default route found")
}

func (a *Agent) Register(server *grpc.Server) error {
	v1.RegisterAgentServer(server, a)
	v1.RegisterDHCPServer(server, a)
	return nil
}

func (a *Agent) RegisterTCP(server *grpc.Server) error {
	v1.RegisterAgentServer(server, a)
	v1.RegisterDHCPServer(server, a)
	return nil
}

// getServicesOpts get service options from plugin context.
func getServicesOpts(ic *plugin.InitContext) ([]containerd.ServicesOpt, error) {
	plugins, err := ic.GetByType(plugin.ServicePlugin)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service plugin")
	}

	opts := []containerd.ServicesOpt{
		containerd.WithEventService(ic.Events),
	}
	for s, fn := range map[string]func(interface{}) containerd.ServicesOpt{
		services.ContentService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithContentStore(s.(content.Store))
		},
		services.ImagesService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithImageService(s.(images.ImagesClient))
		},
		services.SnapshotsService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithSnapshotters(s.(map[string]snapshots.Snapshotter))
		},
		services.ContainersService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithContainerService(s.(containers.ContainersClient))
		},
		services.TasksService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithTaskService(s.(tasks.TasksClient))
		},
		services.DiffService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithDiffService(s.(diff.DiffClient))
		},
		services.NamespacesService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithNamespaceService(s.(namespaces.NamespacesClient))
		},
		services.LeasesService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithLeasesService(s.(leases.Manager))
		},
	} {
		p := plugins[s]
		if p == nil {
			return nil, errors.Errorf("service %q not found", s)
		}
		i, err := p.Instance()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get instance of service %q", s)
		}
		if i == nil {
			return nil, errors.Errorf("instance of service %q not found", s)
		}
		opts = append(opts, fn(i))
	}
	return opts, nil
}
