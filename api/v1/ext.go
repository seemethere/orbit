package v1

import "encoding/json"

const cniVersion = "0.3.1"

type cni struct {
	Version string `json:"cniVersion,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Master  string `json:"master,omitempty"`
	IPAM    ipam   `json:"ipam,omitempty"`
	Bridge  string `json:"bridge,omitempty"`
}

type ipam struct {
	Type   string `json:"type,omitempty"`
	Subnet string `json:"subnet,omitempty"`
}

func (n *CNINetwork) MarshalCNI() []byte {
	c := cni{
		Version: cniVersion,
		Name:    n.Name,
		Type:    n.Type,
		Master:  n.Master,
		Bridge:  n.Bridge,
	}
	if n.IPAM != nil {
		c.IPAM.Type = n.IPAM.Type
		c.IPAM.Subnet = n.IPAM.Subnet
	}
	data, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return data
}
