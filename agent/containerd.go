package agent

import (
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
	"github.com/stellarproject/orbit/version"
	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"
)

// init func for a containerd service plugin
func init() {
	plugin.Register(&plugin.Registration{
		Type:   plugin.GRPCPlugin,
		ID:     "orbit",
		Config: &Config{},
		Requires: []plugin.Type{
			plugin.ServicePlugin,
		},
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			ic.Meta.Platforms = []is.Platform{platforms.DefaultSpec()}
			exports := make(map[string]string)
			exports["version"] = version.Version
			c := ic.Config.(*Config)
			if c.Iface == "" {
				i, err := getDefaultIface()
				if err != nil {
					return nil, err
				}
				c.Iface = i
			}
			exports["interface"] = c.Iface
			servicesOpts, err := getServicesOpts(ic)
			if err != nil {
				return nil, err
			}
			client, err := containerd.New(
				"",
				containerd.WithDefaultNamespace(config.DefaultNamespace),
				containerd.WithServices(servicesOpts...),
			)
			if err != nil {
				return nil, err
			}
			ic.Meta.Exports = exports
			return New(ic.Context, c, nil, client)
		},
	})
}

func getDefaultIface() (string, error) {
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
	return "", errors.New("no default route found")
}

func (a *Agent) Register(server *grpc.Server) error {
	v1.RegisterAgentServer(server, a)
	return nil
}

func (a *Agent) RegisterTCP(server *grpc.Server) error {
	v1.RegisterAgentServer(server, a)
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
