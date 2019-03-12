package v1

import (
	"context"
	"path/filepath"

	"github.com/containerd/containerd"
)

const (
	Root             = "/var/lib/orbit"
	state            = "/run/orbit"
	DefaultRuntime   = "io.containerd.runc.v2"
	DefaultNamespace = "orbit"
)

func StatePath(id string) string {
	return filepath.Join(state, id)
}

// Register is an object that registers and manages service information in its backend
type Register interface {
	Register(id, name, ip string, s *Service) error
	Deregister(id, name string) error
}

type Network interface {
	Create(context.Context, containerd.Container) (string, error)
	Remove(context.Context, containerd.Container) error
}

func NetworkPath(id string) string {
	return filepath.Join(StatePath(id), "net")
}

func ConfigPath(id, name string) string {
	return filepath.Join(StatePath(id), "configs", name)
}
