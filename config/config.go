package config

import (
	"context"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/containerd/containerd"
	"github.com/hashicorp/consul/api"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/consulregister"
)

const (
	Path = "/etc/boss/boss.toml"
)

type ConfigStore interface {
	Write(context.Context, *v1.Container) error
	Watch(context.Context, containerd.Container, *v1.Container) (<-chan error, error)
}

// Load the system config from disk
// fix up any missing fields with runtime data
func Load() (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(Path, &c); err != nil {
		return nil, err
	}
	if c.ID == "" {
		id, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		c.ID = id
	}
	if c.Iface == "" {
		c.Iface = "eth0"
	}
	return &c, nil
}

var (
	consul     *api.Client
	consulErr  error
	consulOnce sync.Once
)

func getConsul() {
	consul, consulErr = api.NewClient(api.DefaultConfig())
}

type Config struct {
	ID           string        `toml:"id"`
	Iface        string        `toml:"iface"`
	Domain       string        `toml:"domain"`
	Buildkit     *Buildkit     `toml:"buildkit"`
	CNI          *CNI          `toml:"cni"`
	Consul       *Consul       `toml:"consul"`
	NodeExporter *NodeExporter `toml:"nodeexporter"`
	Nameservers  []string      `toml:"nameservers"`
	Timezone     string        `toml:"timezone"`
	MOTD         *MOTD         `toml:"motd"`
	SSH          *SSH          `toml:"ssh"`
	Agent        Agent         `toml:"agent"`
	Containerd   Containerd    `toml:"containerd"`
	Criu         *Criu         `toml:"criu"`
}

func (c *Config) Store() (ConfigStore, error) {
	if c.Consul != nil {
		consulOnce.Do(getConsul)
		if consulErr != nil {
			return nil, consulErr
		}
		return &configStore{
			consul: consul,
		}, nil
	}
	return &nullStore{}, nil
}

func (c *Config) GetNameservers() ([]string, error) {
	if c.Consul != nil {
		consulOnce.Do(getConsul)
		if consulErr != nil {
			return nil, consulErr
		}
		nodes, _, err := consul.Catalog().Nodes(&api.QueryOptions{})
		if err != nil {
			return nil, err
		}
		var ns []string
		for _, n := range nodes {
			ns = append(ns, n.Address)
		}
		return ns, nil
	}
	if len(c.Nameservers) == 0 {
		return []string{
			"8.8.8.8",
			"8.8.4.4",
		}, nil
	}
	return c.Nameservers, nil
}

func (c *Config) GetRegister() (v1.Register, error) {
	if c.Consul != nil {
		consulOnce.Do(getConsul)
		if consulErr != nil {
			return nil, consulErr
		}
		return consulregister.New(consul), nil
	}
	return &nullRegister{}, nil
}

func (c *Config) consul() bool {
	return c.Consul != nil
}
