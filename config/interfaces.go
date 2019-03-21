package config

import (
	"path/filepath"

	"github.com/stellarproject/orbit/api/v1"
)

const (
	state            = "/run/orbit"
	DefaultRuntime   = "io.containerd.runc.v2"
	DefaultNamespace = "orbit"
)

func StatePath(id string) string {
	return filepath.Join(state, id)
}

// Register is an object that registers and manages service information in its backend
type Register interface {
	Register(id, name, ip string, s *v1.Service) error
	Deregister(id, name string) error
}

func ConfigPath(id, name string) string {
	return filepath.Join(StatePath(id), "configs", name)
}
