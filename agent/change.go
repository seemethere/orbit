package agent

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/flux"
	"github.com/stellarproject/orbit/opts"
)

type change interface {
	update(context.Context, containerd.Container) error
}

type imageUpdateChange struct {
	ref    string
	client *containerd.Client
}

func (c *imageUpdateChange) update(ctx context.Context, container containerd.Container) error {
	image, err := c.client.Pull(ctx, c.ref, containerd.WithPullUnpack, withPlainRemote(c.ref))
	if err != nil {
		return err
	}
	return container.Update(ctx, flux.WithUpgrade(image))
}

type deregisterChange struct {
	register v1.Register
	name     string
}

func (c *deregisterChange) update(ctx context.Context, container containerd.Container) error {
	return c.register.Deregister(container.ID(), c.name)
}

type configChange struct {
	c          *v1.Container
	client     *containerd.Client
	volumeRoot string
}

func (c *configChange) update(ctx context.Context, container containerd.Container) error {
	image, err := c.client.GetImage(ctx, c.c.Image)
	if err != nil {
		return err
	}
	return container.Update(ctx, opts.WithSetPreviousConfig, opts.WithBossConfig(c.volumeRoot, c.c, image))
}

type filesChange struct {
	c *v1.Container
}

func (c *filesChange) update(ctx context.Context, container containerd.Container) error {
	return nil
	// return c.store.Write(ctx, c.c)
}

func pauseAndRun(ctx context.Context, container containerd.Container, fn func() error) error {
	task, err := container.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return fn()
		}
		return err
	}
	if err := task.Pause(ctx); err != nil {
		return err
	}
	defer task.Resume(ctx)
	return fn()
}

func withImage(i containerd.Image) containerd.UpdateContainerOpts {
	return func(ctx context.Context, client *containerd.Client, c *containers.Container) error {
		c.Image = i.Name()
		return nil
	}
}
