package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd/namespaces"
	raven "github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	"github.com/stellarproject/orbit/api"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/cmd"
	"github.com/stellarproject/orbit/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "ob"
	app.Version = version.Version
	app.Usage = "taking containers to space"
	app.Description = cmd.Banner
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output in the logs",
		},
		cli.StringFlag{
			Name:   "agent",
			Usage:  "agent address",
			Value:  "0.0.0.0:1337",
			EnvVar: "AGENT_ADDR",
		},
		cli.StringFlag{
			Name:   "sentry-dsn",
			Usage:  "sentry DSN",
			EnvVar: "SENTRY_DSN",
		},
	}
	app.Before = func(clix *cli.Context) error {
		if clix.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if dsn := clix.GlobalString("sentry-dsn"); dsn != "" {
			raven.SetDSN(dsn)
			raven.DefaultClient.SetRelease(version.Version)
		}
		return nil
	}
	app.Commands = []cli.Command{
		checkpointCommand,
		createCommand,
		deleteCommand,
		getCommand,
		killCommand,
		listCommand,
		migrateCommand,
		pushCommand,
		restoreCommand,
		rollbackCommand,
		startCommand,
		stopCommand,
		updateCommand,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		raven.CaptureErrorAndWait(err, nil)
		os.Exit(1)
	}
}

func Context() context.Context {
	return namespaces.WithNamespace(context.Background(), v1.DefaultNamespace)
}

func Agent(clix *cli.Context) (*api.LocalAgent, error) {
	return api.Agent(clix.GlobalString("agent"))
}
