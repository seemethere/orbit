package cni

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/containerd/containerd"
	gocni "github.com/containerd/go-cni"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stellarproject/orbit/opts"
	"github.com/stellarproject/orbit/route"
	"golang.org/x/sys/unix"
)

type Config struct {
	Type       string
	State      string
	Iface      string
	BridgeAddr string
}

func New(c Config, n gocni.CNI) (*cni, error) {
	return &cni{
		network: n,
		config:  c,
	}, nil
}

type cni struct {
	network gocni.CNI
	config  Config
}

func (n *cni) Create(ctx context.Context, task containerd.Container) (string, error) {
	path := filepath.Join(n.config.State, task.ID(), "net")
	if _, err := os.Lstat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", errors.Wrap(err, "lstat network namespace")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return "", errors.Wrap(err, "mkdir of network path")
		}
		if err := createNetns(path); err != nil {
			return "", errors.Wrap(err, "create netns")
		}
		result, err := n.network.Setup(task.ID(), path)
		if err != nil {
			return "", errors.Wrap(err, "setup cni network")
		}
		var ip net.IP
		for _, ipc := range result.Interfaces["eth0"].IPConfigs {
			if f := ipc.IP.To4(); f != nil {
				ip = f
				break
			}
		}
		if err := task.Update(ctx, opts.WithIP(ip.String())); err != nil {
			return "", errors.Wrap(err, "update with ip")
		}
		if n.config.Type == "macvlan" {
			route.Remove(ip.String())
			if err := route.Add(ip.String()); err != nil {
				return "", errors.Wrap(err, "add route")
			}
		}
		return ip.String(), nil
	}
	l, err := task.Labels(ctx)
	if err != nil {
		return "", errors.Wrap(err, "get container labels for ip")
	}
	return l[opts.IPLabel], nil
}

func (n *cni) Remove(ctx context.Context, c containerd.Container) error {
	path := filepath.Join(n.config.State, c.ID(), "net")
	if err := n.network.Remove(c.ID(), path); err != nil {
		logrus.WithError(err).Error("remove cni gocni")
	}
	if err := unix.Unmount(path, 0); err != nil {
		logrus.WithError(err).Error("unmount netns")
	}
	if n.config.Type == "macvlan" {
		info, err := c.Info(ctx)
		if err != nil {
			return err
		}
		ip := info.Labels[opts.IPLabel]
		if ip != "" {
			if err := route.Remove(ip); err != nil {
				logrus.WithError(err).Error("remove routes")
			}
		}
	}
	// FIXME this could cause issues later but whatever...
	return os.RemoveAll(filepath.Dir(path))
}

func createNetns(path string) error {
	cmd := exec.Command("orbit-network", "create", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWNET,
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}
