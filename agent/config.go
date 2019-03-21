package agent

import (
	"context"
	"time"

	"github.com/containerd/containerd"
)

type Config struct {
	ID            string   `toml:"id"` //TODO: remove for hostname
	Root          string   `toml:"-"`
	Iface         string   `toml:"iface"`                     // TODO: dynamic public route
	Domain        string   `toml:"domain,omitempty" json:"-"` // TODO: hostname and domain name
	Nameservers   []string `toml:"nameservers"`
	Timezone      string   `toml:"timezone"`
	PlainRemotes  []string `toml:"plain_remotes"`
	VolumeRoot    string   `toml:"volume_root"`
	Interval      duration `toml:"supervisor_interval"`
	BridgeAddress string   `toml:"bridge_address" json:"-"`
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
