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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"github.com/containerd/containerd/defaults"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	raven "github.com/getsentry/raven-go"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/cmd"
	"github.com/stellarproject/orbit/util"
	bv "github.com/stellarproject/orbit/version"
	"github.com/urfave/cli"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "create" {
		runtime.LockOSThread()
		app := cli.NewApp()
		app.Name = "orbit-network"
		app.Version = bv.Version
		app.Usage = "orbit network namespace creation"
		app.Description = cmd.Banner
		app.Before = func(clix *cli.Context) error {
			if dsn := clix.GlobalString("sentry-dsn"); dsn != "" {
				raven.SetDSN(dsn)
				raven.DefaultClient.SetRelease(bv.Version)
			}
			return nil
		}
		app.Commands = []cli.Command{
			networkCreateCommand,
		}
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			raven.CaptureErrorAndWait(err, nil)
			os.Exit(1)
		}
		return
	}
	skel.PluginMain(dhcpAdd, dhcpDelete, version.All)
}

func dhcpAdd(args *skel.CmdArgs) error {
	// The daemon may be running under a different working dir
	// so make sure the netns path is absolute.
	netns, err := filepath.Abs(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to make %q an absolute path: %v", args.Netns, err)
	}
	args.Netns = netns

	// Plugin must return result in same version as specified in netconf
	versionDecoder := &version.ConfigDecoder{}
	confVersion, err := versionDecoder.Decode(args.StdinData)
	if err != nil {
		return err
	}

	agent, err := util.Agent(defaults.DefaultAddress)
	if err != nil {
		return err
	}
	defer agent.Close()

	var conf types.NetConf
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return err
	}

	ctx := context.Background()
	result, err := agent.DHCPAdd(ctx, &v1.DHCPAddRequest{
		ID:    args.ContainerID,
		Netns: args.Netns,
		Iface: args.IfName,
		Name:  conf.Name,
	})
	if err != nil {
		return err
	}
	out := &current.Result{}
	for _, ip := range result.IPs {
		out.IPs = append(out.IPs, &current.IPConfig{
			Version: ip.Version,
			Address: net.IPNet{
				IP:   net.IP(ip.Address.IP),
				Mask: net.IPMask(ip.Address.Mask),
			},
			Gateway: net.IP(ip.Gateway),
		})
	}
	for _, r := range result.Routes {
		out.Routes = append(out.Routes, &types.Route{
			Dst: net.IPNet{
				IP:   net.IP(r.Dst.IP),
				Mask: net.IPMask(r.Dst.Mask),
			},
			GW: net.IP(r.Gw),
		})
	}
	return types.PrintResult(out, confVersion)
}

func dhcpDelete(args *skel.CmdArgs) error {
	// The daemon may be running under a different working dir
	// so make sure the netns path is absolute.
	netns, err := filepath.Abs(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to make %q an absolute path: %v", args.Netns, err)
	}
	args.Netns = netns

	agent, err := util.Agent(defaults.DefaultAddress)
	if err != nil {
		return err
	}
	defer agent.Close()

	ctx := context.Background()
	_, err = agent.DHCPDelete(ctx, &v1.DHCPDeleteRequest{
		ID:    args.ContainerID,
		Netns: args.Netns,
		Iface: args.IfName,
	})
	return err
}
