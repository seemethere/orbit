package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	proto "github.com/gogo/protobuf/types"
	v1 "github.com/stellarproject/orbit/api/v1"

	"github.com/d2g/dhcp4"
	"github.com/vishvananda/netlink"

	"github.com/containernetworking/plugins/pkg/ns"
)

func (a *Agent) DHCPAdd(ctx context.Context, r *v1.DHCPAddRequest) (*v1.DHCPAddResponse, error) {
	clientID := generateClientID(r.ID, r.Name, r.Iface)
	l, err := AcquireLease(clientID, r.Netns, r.Iface)
	if err != nil {
		return nil, err
	}

	ipn, err := l.IPNet()
	if err != nil {
		l.Stop()
		return nil, err
	}

	a.dhcp.setLease(clientID, l)
	result := &v1.DHCPAddResponse{
		IPs: []*v1.CNIIP{
			{
				Version: "4",
				Address: ipn,
				Gateway: []byte(l.Gateway()),
			},
		},
		Routes: l.Routes(),
	}
	return result, nil
}

func (a *Agent) DHCPDelete(ctx context.Context, r *v1.DHCPDeleteRequest) (*proto.Empty, error) {
	clientID := generateClientID(r.ID, r.Name, r.Iface)
	if l := a.dhcp.getLease(clientID); l != nil {
		l.Stop()
		a.dhcp.clearLease(clientID)
	}
	return empty, nil
}

// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// RFC 2131 suggests using exponential backoff, starting with 4sec
// and randomized to +/- 1sec
const resendDelay0 = 4 * time.Second
const resendDelayMax = 32 * time.Second

const (
	leaseStateBound = iota
	leaseStateRenewing
	leaseStateRebinding
)

// This implementation uses 1 OS thread per lease. This is because
// all the network operations have to be done in network namespace
// of the interface. This can be improved by switching to the proper
// namespace for network ops and using fewer threads. However, this
// needs to be done carefully as dhcp4client ops are blocking.

type DHCPLease struct {
	clientID      string
	ack           *dhcp4.Packet
	opts          dhcp4.Options
	link          netlink.Link
	renewalTime   time.Time
	rebindingTime time.Time
	expireTime    time.Time
	stopping      uint32
	stop          chan struct{}
	wg            sync.WaitGroup
}

// AcquireLease gets an DHCP lease and then maintains it in the background
// by periodically renewing it. The acquired lease can be released by
// calling DHCPLease.Stop()
func AcquireLease(clientID, netns, ifName string) (*DHCPLease, error) {
	errCh := make(chan error, 1)
	l := &DHCPLease{
		clientID: clientID,
		stop:     make(chan struct{}),
	}

	log.Printf("%v: acquiring lease", clientID)

	l.wg.Add(1)
	go func() {
		errCh <- ns.WithNetNSPath(netns, func(_ ns.NetNS) error {
			defer l.wg.Done()

			link, err := netlink.LinkByName(ifName)
			if err != nil {
				return fmt.Errorf("error looking up %q: %v", ifName, err)
			}

			l.link = link

			if err = l.acquire(); err != nil {
				return err
			}

			log.Printf("%v: lease acquired, expiration is %v", l.clientID, l.expireTime)

			errCh <- nil

			l.maintain()
			return nil
		})
	}()

	if err := <-errCh; err != nil {
		return nil, err
	}

	return l, nil
}

// Stop terminates the background task that maintains the lease
// and issues a DHCP Release
func (l *DHCPLease) Stop() {
	if atomic.CompareAndSwapUint32(&l.stopping, 0, 1) {
		close(l.stop)
	}
	l.wg.Wait()
}

func (l *DHCPLease) acquire() error {
	c, err := newDHCPClient(l.link, l.clientID)
	if err != nil {
		return err
	}
	defer c.Close()

	if (l.link.Attrs().Flags & net.FlagUp) != net.FlagUp {
		log.Printf("Link %q down. Attempting to set up", l.link.Attrs().Name)
		if err = netlink.LinkSetUp(l.link); err != nil {
			return err
		}
	}

	opts := make(dhcp4.Options)
	opts[dhcp4.OptionClientIdentifier] = []byte(l.clientID)
	opts[dhcp4.OptionParameterRequestList] = []byte{byte(dhcp4.OptionRouter)}

	pkt, err := backoffRetry(func() (*dhcp4.Packet, error) {
		ok, ack, err := DhcpRequest(c, opts)
		switch {
		case err != nil:
			return nil, err
		case !ok:
			return nil, fmt.Errorf("DHCP server NACK'd own offer")
		default:
			return &ack, nil
		}
	})
	if err != nil {
		return err
	}

	return l.commit(pkt)
}

func (l *DHCPLease) commit(ack *dhcp4.Packet) error {
	opts := ack.ParseOptions()

	leaseTime, err := parseLeaseTime(opts)
	if err != nil {
		return err
	}

	rebindingTime, err := parseRebindingTime(opts)
	if err != nil || rebindingTime > leaseTime {
		// Per RFC 2131 Section 4.4.5, it should default to 85% of lease time
		rebindingTime = leaseTime * 85 / 100
	}

	renewalTime, err := parseRenewalTime(opts)
	if err != nil || renewalTime > rebindingTime {
		// Per RFC 2131 Section 4.4.5, it should default to 50% of lease time
		renewalTime = leaseTime / 2
	}

	now := time.Now()
	l.expireTime = now.Add(leaseTime)
	l.renewalTime = now.Add(renewalTime)
	l.rebindingTime = now.Add(rebindingTime)
	l.ack = ack
	l.opts = opts

	return nil
}

func (l *DHCPLease) maintain() {
	state := leaseStateBound

	for {
		var sleepDur time.Duration

		switch state {
		case leaseStateBound:
			sleepDur = l.renewalTime.Sub(time.Now())
			if sleepDur <= 0 {
				log.Printf("%v: renewing lease", l.clientID)
				state = leaseStateRenewing
				continue
			}

		case leaseStateRenewing:
			if err := l.renew(); err != nil {
				log.Printf("%v: %v", l.clientID, err)

				if time.Now().After(l.rebindingTime) {
					log.Printf("%v: renawal time expired, rebinding", l.clientID)
					state = leaseStateRebinding
				}
			} else {
				log.Printf("%v: lease renewed, expiration is %v", l.clientID, l.expireTime)
				state = leaseStateBound
			}

		case leaseStateRebinding:
			if err := l.acquire(); err != nil {
				log.Printf("%v: %v", l.clientID, err)

				if time.Now().After(l.expireTime) {
					log.Printf("%v: lease expired, bringing interface DOWN", l.clientID)
					l.downIface()
					return
				}
			} else {
				log.Printf("%v: lease rebound, expiration is %v", l.clientID, l.expireTime)
				state = leaseStateBound
			}
		}

		select {
		case <-time.After(sleepDur):

		case <-l.stop:
			if err := l.release(); err != nil {
				log.Printf("%v: failed to release DHCP lease: %v", l.clientID, err)
			}
			return
		}
	}
}

func (l *DHCPLease) downIface() {
	if err := netlink.LinkSetDown(l.link); err != nil {
		log.Printf("%v: failed to bring %v interface DOWN: %v", l.clientID, l.link.Attrs().Name, err)
	}
}

func (l *DHCPLease) renew() error {
	c, err := newDHCPClient(l.link, l.clientID)
	if err != nil {
		return err
	}
	defer c.Close()

	opts := make(dhcp4.Options)
	opts[dhcp4.OptionClientIdentifier] = []byte(l.clientID)

	pkt, err := backoffRetry(func() (*dhcp4.Packet, error) {
		ok, ack, err := DhcpRenew(c, *l.ack, opts)
		switch {
		case err != nil:
			return nil, err
		case !ok:
			return nil, fmt.Errorf("DHCP server did not renew lease")
		default:
			return &ack, nil
		}
	})
	if err != nil {
		return err
	}

	l.commit(pkt)
	return nil
}

func (l *DHCPLease) release() error {
	log.Printf("%v: releasing lease", l.clientID)

	c, err := newDHCPClient(l.link, l.clientID)
	if err != nil {
		return err
	}
	defer c.Close()

	opts := make(dhcp4.Options)
	opts[dhcp4.OptionClientIdentifier] = []byte(l.clientID)

	if err = DhcpRelease(c, *l.ack, opts); err != nil {
		return fmt.Errorf("failed to send DHCPRELEASE")
	}

	return nil
}

func (l *DHCPLease) IPNet() (*v1.CNIIPNet, error) {
	mask := parseSubnetMask(l.opts)
	if mask == nil {
		return nil, fmt.Errorf("DHCP option Subnet Mask not found in DHCPACK")
	}

	return &v1.CNIIPNet{
		IP:   []byte(l.ack.YIAddr()),
		Mask: []byte(mask),
	}, nil
}

func (l *DHCPLease) Gateway() net.IP {
	return parseRouter(l.opts)
}

func (l *DHCPLease) Routes() []*v1.CNIRoute {
	routes := []*v1.CNIRoute{}

	// RFC 3442 states that if Classless Static Routes (option 121)
	// exist, we ignore Static Routes (option 33) and the Router/Gateway.
	opt121_routes := parseCIDRRoutes(l.opts)
	if len(opt121_routes) > 0 {
		return append(routes, opt121_routes...)
	}

	// Append Static Routes
	routes = append(routes, parseRoutes(l.opts)...)

	// The CNI spec says even if there is a gateway specified, we must
	// add a default route in the routes section.
	if gw := l.Gateway(); gw != nil {
		_, defaultRoute, _ := net.ParseCIDR("0.0.0.0/0")
		routes = append(routes, &v1.CNIRoute{
			Dst: &v1.CNIIPNet{
				IP:   []byte(defaultRoute.IP),
				Mask: []byte(defaultRoute.Mask),
			},
			Gw: []byte(gw),
		})
	}
	return routes
}

// jitter returns a random value within [-span, span) range
func jitter(span time.Duration) time.Duration {
	return time.Duration(float64(span) * (2.0*rand.Float64() - 1.0))
}

func backoffRetry(f func() (*dhcp4.Packet, error)) (*dhcp4.Packet, error) {
	var baseDelay time.Duration = resendDelay0

	for i := 0; i < resendCount; i++ {
		pkt, err := f()
		if err == nil {
			return pkt, nil
		}

		log.Print(err)

		time.Sleep(baseDelay + jitter(time.Second))

		if baseDelay < resendDelayMax {
			baseDelay *= 2
		}
	}

	return nil, errNoMoreTries
}

const (
	listenFdsStart = 3
	resendCount    = 3
)

var errNoMoreTries = errors.New("no more tries")

type DHCP struct {
	mux    sync.Mutex
	leases map[string]*DHCPLease
}

func newDHCP() *DHCP {
	return &DHCP{
		leases: make(map[string]*DHCPLease),
	}
}

func generateClientID(containerID string, netName string, ifName string) string {
	return containerID + "/" + netName + "/" + ifName
}

func (d *DHCP) getLease(clientID string) *DHCPLease {
	d.mux.Lock()
	defer d.mux.Unlock()

	// TODO(eyakubovich): hash it to avoid collisions
	l, ok := d.leases[clientID]
	if !ok {
		return nil
	}
	return l
}

func (d *DHCP) setLease(clientID string, l *DHCPLease) {
	d.mux.Lock()
	defer d.mux.Unlock()

	// TODO(eyakubovich): hash it to avoid collisions
	d.leases[clientID] = l
}

//func (d *DHCP) clearLease(contID, netName, ifName string) {
func (d *DHCP) clearLease(clientID string) {
	d.mux.Lock()
	defer d.mux.Unlock()

	// TODO(eyakubovich): hash it to avoid collisions
	delete(d.leases, clientID)
}

func getListener(socketPath string) (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, err
	}
	return net.Listen("unix", socketPath)
}
