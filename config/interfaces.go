package config

import (
	"github.com/stellarproject/orbit/api/v1"
)

const (
	DefaultRuntime   = "io.containerd.runc.v2"
	DefaultNamespace = "orbit"
)

// Register is an object that registers and manages service information in its backend
type Register interface {
	Register(id, name, ip string, s *v1.Service) error
	Deregister(id, name string) error
}
