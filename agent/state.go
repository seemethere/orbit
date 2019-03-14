package agent

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/contrib/apparmor"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type stateChange interface {
	apply(context.Context) error
}

type stopDiff struct {
	a         *Agent
	container containerd.Container
}

func (s *stopDiff) apply(ctx context.Context) error {
	return s.a.stop(ctx, s.container)
}

type startDiff struct {
	a         *Agent
	container containerd.Container
}

func (s *startDiff) apply(ctx context.Context) error {
	return s.a.start(ctx, s.container)
}

func sameDiff() stateChange {
	return &nullDiff{}
}

type nullDiff struct {
}

func (n *nullDiff) apply(_ context.Context) error {
	return nil
}

func setupApparmor() error {
	return apparmor.WithDefaultProfile("orbit")(nil, nil, nil, &specs.Spec{
		Process: &specs.Process{},
	})
}
