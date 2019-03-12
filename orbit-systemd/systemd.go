package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/sirupsen/logrus"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/config"
	"github.com/stellarproject/orbit/opts"
	"github.com/stellarproject/orbit/system"
	"github.com/urfave/cli"
)

var (
	errIDRequired     = errors.New("container id is required")
	errUnableToSignal = errors.New("unable to signal task")
)

var systemdExecStartCommand = cli.Command{
	Name:  "exec-start",
	Usage: "exec-start proxy for containers",
	Action: func(clix *cli.Context) error {
		id := clix.Args().First()
		var (
			signals = make(chan os.Signal, 64)
			ctx     = system.Context()
		)
		signal.Notify(signals)
		client, err := system.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()
		container, err := client.LoadContainer(ctx, id)
		if err != nil {
			return err
		}
		desc, err := opts.GetRestoreDesc(ctx, container)
		if err != nil {
			return err
		}
		cfg, err := opts.GetConfig(ctx, container)
		if err != nil {
			return err
		}
		c, err := config.Load()
		if err != nil {
			return err
		}
		register, err := c.GetRegister()
		if err != nil {
			return err
		}
		store, err := c.Store()
		if err != nil {
			return err
		}
		templateCh, err := store.Watch(ctx, container, cfg)
		if err != nil {
			return err
		}
		ip, err := setupNetworking(ctx, container, cfg)
		if err != nil {
			return err
		}
		if err := container.Update(ctx, opts.WithIP(ip), opts.WithoutRestore); err != nil {
			return err
		}
		task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio), opts.WithTaskRestore(desc))
		if err != nil {
			return err
		}
		status, err := monitorTask(ctx, client, task, cfg, register, signals, templateCh)
		if err != nil {
			return err
		}
		os.Exit(status)
		return nil
	},
}

func monitorTask(ctx context.Context, client *containerd.Client, task containerd.Task, config *v1.Container, register v1.Register, signals chan os.Signal, templateCh <-chan error) (int, error) {
	defer task.Delete(ctx, containerd.WithProcessKill)
	started := make(chan error, 1)
	wait, err := task.Wait(ctx)
	if err != nil {
		return -1, err
	}
	go func() {
		started <- task.Start(ctx)
	}()
	for {
		select {
		case err := <-templateCh:
			logrus.WithError(err).Error("render template")
		case err := <-started:
			if err != nil {
				return -1, err
			}
			for name := range config.Services {
				if err := register.DisableMaintainance(task.ID(), name); err != nil {
					logrus.WithError(err).Error("disable service maintenance")
				}
			}
		case s := <-signals:
			if err := trySendSignal(ctx, client, task, s); err != nil {
				logrus.WithError(err).Error("signal task")
			}
		case exit := <-wait:
			if exit.Error() != nil {
				if !isUnavailable(err) {
					return -1, err
				}
				if err := reconnect(client); err != nil {
					return -1, err
				}
				if wait, err = task.Wait(ctx); err != nil {
					return -1, err
				}
				continue
			}
			return int(exit.ExitCode()), nil
		}
	}
}

func cleanupPreviousTask(id string) error {
	ctx := system.Context()
	client, err := system.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	container, err := client.LoadContainer(ctx, id)
	if err != nil {
		return err
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return err
	}
	_, err = task.Delete(ctx, containerd.WithProcessKill)
	return err
}
