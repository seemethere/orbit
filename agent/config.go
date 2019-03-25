package agent

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/stellarproject/orbit/opts"
	"github.com/stellarproject/orbit/util"
)

type Config struct {
	// UTS
	ID     string `toml:"id"`
	Domain string `toml:"domain,omitempty"`

	// containerd provided
	Root  string `toml:"-"`
	State string `toml:"-"`

	// Networking
	BridgeAddress string `toml:"bridge_address"`
	Iface         string `toml:"iface"`

	PlainRemotes []string `toml:"plain_remotes"`
	VolumeRoot   string   `toml:"volume_root"`
	Interval     duration `toml:"interval"`

	// Store
	Master bool `toml:"master"`

	ip    string
	ipErr error
	ipO   sync.Once
}

func (c *Config) IP() (string, error) {
	c.ipO.Do(func() {
		c.ip, c.ipErr = util.GetIP(c.Iface)
	})
	if c.ipErr != nil {
		return "", c.ipErr
	}
	return c.ip, nil
}

func (c *Config) Paths(id string) opts.Paths {
	return opts.Paths{
		Root:   filepath.Join(c.Root, id),
		State:  filepath.Join(c.State, id),
		Volume: c.VolumeRoot,
	}
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (d duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

type host struct {
	ip string
}

func (n *host) Create(_ context.Context, _ containerd.Container) (string, error) {
	return n.ip, nil
}

func (n *host) Remove(_ context.Context, _ containerd.Container) error {
	return nil
}

type none struct {
}

func (n *none) Create(_ context.Context, _ containerd.Container) (string, error) {
	return "", nil
}

func (n *none) Remove(_ context.Context, _ containerd.Container) error {
	return nil
}
