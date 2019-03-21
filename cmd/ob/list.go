package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	units "github.com/docker/go-units"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/urfave/cli"
)

var listCommand = cli.Command{
	Name:  "list",
	Usage: "list containers",
	Action: func(clix *cli.Context) error {
		ctx := Context()
		agent, err := Agent(clix)
		if err != nil {
			return err
		}
		defer agent.Close()
		resp, err := agent.List(ctx, &v1.ListRequest{})
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 10, 1, 3, ' ', 0)
		const tfmt = "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\n"
		fmt.Fprint(w, "ID\tIMAGE\tSTATUS\tIP\tCPU\tMEMORY\tPIDS\tSIZE\tREVISIONS\n")
		for _, c := range resp.Containers {
			ip := "none"
			if len(c.Services) > 0 {
				ip = c.Services[0].IP
			}
			fmt.Fprintf(w, tfmt,
				c.ID,
				c.Image,
				c.Status,
				ip,
				time.Duration(int64(c.Cpu)),
				fmt.Sprintf("%s/%s", units.HumanSize(c.MemoryUsage), units.HumanSize(c.MemoryLimit)),
				fmt.Sprintf("%d/%d", c.PidUsage, c.PidLimit),
				units.HumanSize(float64(c.FsSize)),
				len(c.Snapshots),
			)
		}
		return w.Flush()
	},
}
