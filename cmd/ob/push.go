package main

import (
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/urfave/cli"
)

var pushCommand = cli.Command{
	Name:  "push",
	Usage: "push a image to a registry",
	Action: func(clix *cli.Context) error {
		agent, err := Agent(clix)
		if err != nil {
			return err
		}
		defer agent.Close()
		_, err = agent.Push(Context(), &v1.PushRequest{
			Ref: clix.Args().First(),
		})
		return err
	},
}
