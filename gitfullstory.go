package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "gitfullstory"
	app.Usage = "command-line utility for seeing your team's open pull requests"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "github_token, gt",
			Value: "",
			Usage: "your personal access token for github",
			EnvVar: "GITHUB_ACCESS_TOKEN, GITHUB_TOKEN",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a task on the list",
			Action:  func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a task to the list",
			Action:  func(c *cli.Context) error {
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Action = func(c *cli.Context) error {
		fmt.Println("boom! I say!")
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}