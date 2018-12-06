package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli"
)

func createGithubClient(accessToken string) (ctx context.Context, githubClient *github.Client) {
	ctx = context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tokenClient := oauth2.NewClient(ctx, tokenSource)
	githubClient = github.NewClient(tokenClient)
	return
}

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
			EnvVar: "GITHUB_ACCESS_TOKEN,GITHUB_TOKEN",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a task on the list",
			Action:  func(c *cli.Context) error {
				if c.Bool("github_token") {
					fmt.Println("You have a github token!")
				}
				return nil
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a task to the list",
			Action:  func(c *cli.Context) error {
				if c.Bool("github_token") {
					fmt.Println("I'm adding stuff WITH a github token!")
					return nil
				}
				fmt.Println("I'm only adding some stuff...")
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Action = func(c *cli.Context) error {
		if c.String("github_token") == "" {
			return cli.NewExitError("Error, missing github personal access token!", 86)
		}
		ctx, githubClient := createGithubClient(c.String("github_token"))
		repos, _, err := githubClient.Repositories.List(ctx, "", nil)
		if err != nil {
			return cli.NewExitError("Error connecting to github", 87)
		}
		if len(repos) > 0 {
			fmt.Print(repos)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}