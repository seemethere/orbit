package main

import (
	"github.com/BurntSushi/toml"
	api "github.com/stellarproject/orbit/api/v1"
	v1 "github.com/stellarproject/orbit/config/v1"
	"github.com/urfave/cli"
)

var updateCommand = cli.Command{
	Name:  "update",
	Usage: "update an existing container's configuration",
	Action: func(clix *cli.Context) error {
		var (
			path = clix.Args().First()
			ctx  = Context()
		)
		var newConfig v1.Container
		if _, err := toml.DecodeFile(path, &newConfig); err != nil {
			return err
		}
		agent, err := Agent(clix)
		if err != nil {
			return err
		}
		defer agent.Close()
		_, err = agent.Update(ctx, &api.UpdateRequest{
			Container: newConfig.Proto(),
		})
		return err
	},
}
