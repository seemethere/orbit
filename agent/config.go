package agent

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/cni"
	"github.com/stellarproject/orbit/util"
)

type Config struct {
	ID           string   `toml:"id"`     //TODO: remove for hostname
	Iface        string   `toml:"iface"`  // TODO: dynamic public route
	Domain       string   `toml:"domain"` // TODO: hostname and domain name
	CNI          *CNI     `toml:"cni"`    // TODO: move networking to container
	Nameservers  []string `toml:"nameservers"`
	Timezone     string   `toml:"timezone"`
	PlainRemotes []string `toml:"plain_remotes"`
	VolumeRoot   string   `toml:"volume_root"`
	Interval     duration `toml:"supervisor_interval"`
}

type IPAM struct {
	Type   string `toml:"type" json:"type"`
	Subnet string `toml:"subnet" json:"subnet"`
}

type CNI struct {
	Image         string `toml:"image" json:"-"`
	Version       string `toml:"-" json:"cniVersion,omitempty"`
	NetworkName   string `toml:"name" json:"name"`
	Type          string `toml:"type" json:"type"`
	Master        string `toml:"master" json:"master,omitempty"`
	IPAM          IPAM   `toml:"ipam" json:"ipam"`
	Bridge        string `toml:"bridge" json:"bridge,omitempty"`
	BridgeAddress string `toml:"bridge_address" json:"-"`
}

func (c *CNI) Bytes() []byte {
	data, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return data
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

func getNetwork(publicInterface, networkType string, c *CNI) (v1.Network, error) {
	ip, err := util.GetIP(publicInterface)
	if err != nil {
		return nil, err
	}
	switch networkType {
	case "", "none":
		return &none{}, nil
	case "host":
		return &host{ip: ip}, nil
	case "cni":
		if c == nil {
			return nil, errors.New("[cni] is not enabled in the system config")
		}
		if c.Type == "macvlan" && c.BridgeAddress == "" {
			return nil, errors.New("bridge_address must be specified with macvlan")
		}
		// populate cni data from main config if fields are missing
		c.Version = "0.3.1"
		if c.NetworkName == "" {
			c.NetworkName = c.Domain
		}
		if c.Master == "" {
			c.Master = c.Iface
		}
		n, err := gocni.New(
			gocni.WithPluginDir([]string{"/opt/containerd/bin"}),
			gocni.WithConf(c.Bytes()),
			gocni.WithLoNetwork,
		)
		if err != nil {
			return nil, err
		}
		return cni.New(networkType, publicInterface, c.BridgeAddress, n)
	}
	return nil, errors.Errorf("network %s does not exist", networkType)
}
