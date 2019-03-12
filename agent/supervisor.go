package agent

import (
	"context"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/contrib/apparmor"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type stopDiff struct {
	container containerd.Container
}

func (s *stopDiff) apply(ctx context.Context, client *containerd.Client) error {
	return killTask(ctx, s.container)
}

type startDiff struct {
	container containerd.Container
}

func (s *startDiff) apply(ctx context.Context, client *containerd.Client) error {
	log := cio.NullIO
	if s.logPath != "" {
		log = cio.LogFile(s.logPath)
	}
	killTask(ctx, s.container)
	task, err := s.container.NewTask(ctx, log)
	if err != nil {
		return err
	}
	return task.Start(ctx)
}

func killTask(ctx context.Context, container containerd.Container) error {
	task, err := container.Task(ctx, nil)
	if err == nil {
		wait, err := task.Wait(ctx)
		if err != nil {
			if _, derr := task.Delete(ctx); derr == nil {
				return nil
			}
			return err
		}
		if err := task.Kill(ctx, syscall.SIGKILL, containerd.WithKillAll); err != nil {
			if _, derr := task.Delete(ctx); derr == nil {
				return nil
			}
			return err
		}
		<-wait
		if _, err := task.Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}

func sameDiff() stateChange {
	return &nullDiff{}
}

type nullDiff struct {
}

func (n *nullDiff) apply(_ context.Context, _ *containerd.Client) error {
	return nil
}

type stateChange interface {
	apply(context.Context, *containerd.Client) error
}

func setupApparmor() error {
	return apparmor.WithDefaultProfile("boss")(nil, nil, nil, &specs.Spec{
		Process: &specs.Process{},
	})
}
