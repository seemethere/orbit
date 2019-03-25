package agent

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/vishvananda/netlink"
)

func newDHCPClient(link netlink.Link, clientID string) (*dhcp4client.Client, error) {
	pktsock, err := dhcp4client.NewPacketSock(link.Attrs().Index)
	if err != nil {
		return nil, err
	}

	return dhcp4client.New(
		dhcp4client.HardwareAddr(link.Attrs().HardwareAddr),
		dhcp4client.Timeout(5*time.Second),
		dhcp4client.Broadcast(false),
		dhcp4client.Connection(pktsock),
	)
}

const (
	MaxDHCPLen = 576
)

//Send the Discovery Packet to the Broadcast Channel
func DhcpSendDiscoverPacket(c *dhcp4client.Client, options dhcp4.Options) (dhcp4.Packet, error) {
	discoveryPacket := c.DiscoverPacket()

	for opt, data := range options {
		discoveryPacket.AddOption(opt, data)
	}

	discoveryPacket.PadToMinSize()
	return discoveryPacket, c.SendPacket(discoveryPacket)
}

//Send Request Based On the offer Received.
func DhcpSendRequest(c *dhcp4client.Client, options dhcp4.Options, offerPacket *dhcp4.Packet) (dhcp4.Packet, error) {
	requestPacket := c.RequestPacket(offerPacket)

	for opt, data := range options {
		requestPacket.AddOption(opt, data)
	}

	requestPacket.PadToMinSize()

	return requestPacket, c.SendPacket(requestPacket)
}

//Send Decline to the received acknowledgement.
func DhcpSendDecline(c *dhcp4client.Client, acknowledgementPacket *dhcp4.Packet, options dhcp4.Options) (dhcp4.Packet, error) {
	declinePacket := c.DeclinePacket(acknowledgementPacket)

	for opt, data := range options {
		declinePacket.AddOption(opt, data)
	}

	declinePacket.PadToMinSize()

	return declinePacket, c.SendPacket(declinePacket)
}

//Lets do a Full DHCP Request.
func DhcpRequest(c *dhcp4client.Client, options dhcp4.Options) (bool, dhcp4.Packet, error) {
	discoveryPacket, err := DhcpSendDiscoverPacket(c, options)
	if err != nil {
		return false, discoveryPacket, err
	}

	offerPacket, err := c.GetOffer(&discoveryPacket)
	if err != nil {
		return false, offerPacket, err
	}

	requestPacket, err := DhcpSendRequest(c, options, &offerPacket)
	if err != nil {
		return false, requestPacket, err
	}

	acknowledgement, err := c.GetAcknowledgement(&requestPacket)
	if err != nil {
		return false, acknowledgement, err
	}

	acknowledgementOptions := acknowledgement.ParseOptions()
	if dhcp4.MessageType(acknowledgementOptions[dhcp4.OptionDHCPMessageType][0]) != dhcp4.ACK {
		return false, acknowledgement, nil
	}

	return true, acknowledgement, nil
}

//Renew a lease backed on the Acknowledgement Packet.
//Returns Successful, The AcknoledgementPacket, Any Errors
func DhcpRenew(c *dhcp4client.Client, acknowledgement dhcp4.Packet, options dhcp4.Options) (bool, dhcp4.Packet, error) {
	renewRequest := c.RenewalRequestPacket(&acknowledgement)

	for opt, data := range options {
		renewRequest.AddOption(opt, data)
	}

	renewRequest.PadToMinSize()

	err := c.SendPacket(renewRequest)
	if err != nil {
		return false, renewRequest, err
	}

	newAcknowledgement, err := c.GetAcknowledgement(&renewRequest)
	if err != nil {
		return false, newAcknowledgement, err
	}

	newAcknowledgementOptions := newAcknowledgement.ParseOptions()
	if dhcp4.MessageType(newAcknowledgementOptions[dhcp4.OptionDHCPMessageType][0]) != dhcp4.ACK {
		return false, newAcknowledgement, nil
	}

	return true, newAcknowledgement, nil
}

//Release a lease backed on the Acknowledgement Packet.
//Returns Any Errors
func DhcpRelease(c *dhcp4client.Client, acknowledgement dhcp4.Packet, options dhcp4.Options) error {
	release := c.ReleasePacket(&acknowledgement)

	for opt, data := range options {
		release.AddOption(opt, data)
	}

	release.PadToMinSize()

	return c.SendPacket(release)
}

func parseRouter(opts dhcp4.Options) net.IP {
	if opts, ok := opts[dhcp4.OptionRouter]; ok {
		if len(opts) == 4 {
			return net.IP(opts)
		}
	}
	return nil
}

func classfulSubnet(sn net.IP) *v1.CNIIPNet {
	return &v1.CNIIPNet{
		IP:   []byte(sn),
		Mask: []byte(sn.DefaultMask()),
	}
}

func parseRoutes(opts dhcp4.Options) []*v1.CNIRoute {
	// StaticRoutes format: pairs of:
	// Dest = 4 bytes; Classful IP subnet
	// Router = 4 bytes; IP address of router

	routes := []*v1.CNIRoute{}
	if opt, ok := opts[dhcp4.OptionStaticRoute]; ok {
		for len(opt) >= 8 {
			sn := opt[0:4]
			r := opt[4:8]
			rt := &v1.CNIRoute{
				Dst: classfulSubnet(sn),
				Gw:  []byte(r),
			}
			routes = append(routes, rt)
			opt = opt[8:]
		}
	}

	return routes
}

func parseCIDRRoutes(opts dhcp4.Options) []*v1.CNIRoute {
	// See RFC4332 for format (http://tools.ietf.org/html/rfc3442)

	routes := []*v1.CNIRoute{}
	if opt, ok := opts[dhcp4.OptionClasslessRouteFormat]; ok {
		for len(opt) >= 5 {
			width := int(opt[0])
			if width > 32 {
				// error: can't have more than /32
				return nil
			}
			// network bits are compacted to avoid zeros
			octets := 0
			if width > 0 {
				octets = (width-1)/8 + 1
			}

			if len(opt) < 1+octets+4 {
				// error: too short
				return nil
			}

			sn := make([]byte, 4)
			copy(sn, opt[1:octets+1])

			gw := net.IP(opt[octets+1 : octets+5])

			rt := &v1.CNIRoute{
				Dst: &v1.CNIIPNet{
					IP:   sn,
					Mask: []byte(net.CIDRMask(width, 32)),
				},
				Gw: []byte(gw),
			}
			routes = append(routes, rt)

			opt = opt[octets+5 : len(opt)]
		}
	}
	return routes
}

func parseSubnetMask(opts dhcp4.Options) net.IPMask {
	mask, ok := opts[dhcp4.OptionSubnetMask]
	if !ok {
		return nil
	}

	return net.IPMask(mask)
}

func parseDuration(opts dhcp4.Options, code dhcp4.OptionCode, optName string) (time.Duration, error) {
	val, ok := opts[code]
	if !ok {
		return 0, fmt.Errorf("option %v not found", optName)
	}
	if len(val) != 4 {
		return 0, fmt.Errorf("option %v is not 4 bytes", optName)
	}

	secs := binary.BigEndian.Uint32(val)
	return time.Duration(secs) * time.Second, nil
}

func parseLeaseTime(opts dhcp4.Options) (time.Duration, error) {
	return parseDuration(opts, dhcp4.OptionIPAddressLeaseTime, "LeaseTime")
}

func parseRenewalTime(opts dhcp4.Options) (time.Duration, error) {
	return parseDuration(opts, dhcp4.OptionRenewalTimeValue, "RenewalTime")
}

func parseRebindingTime(opts dhcp4.Options) (time.Duration, error) {
	return parseDuration(opts, dhcp4.OptionRebindingTimeValue, "RebindingTime")
}
