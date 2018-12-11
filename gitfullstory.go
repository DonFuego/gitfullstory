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

var ctx context.Context
var githubClient *github.Client

var organizations []*github.Organization
var fetchOrgsError error
var repos = make([]*github.Repository, 0)

// globalListOptions is the setting used for all github api list functions
// keep in mind, github api returns default of 30 and max of 100 so you must use pagination for > 100
var globalListOptions = &github.ListOptions {
	PerPage: 100,
	Page: 1,
}

func initializeGithubClient(accessToken string) {
	ctx = context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tokenClient := oauth2.NewClient(ctx, tokenSource)
	githubClient = github.NewClient(tokenClient)
}

func fetchOrganizations() {
	organizations, _, fetchOrgsError = githubClient.Organizations.List(ctx, "", nil)
	if fetchOrgsError != nil {
		log.Fatal("Error retrieving organizations from Github!")
	}
}

func fetchAllRepositoriesByOrgName(orgName string, repositoryListOptions *github.RepositoryListByOrgOptions) {
	log.Printf("Fetching repositories for org: %s on page %d", orgName, repositoryListOptions.ListOptions.Page)
	repositories, response, err := githubClient.Repositories.ListByOrg(ctx, orgName, repositoryListOptions)

	if err  != nil {
		log.Fatalf("Error fetching repositories for organization %s.  Error: %s", orgName, err)
	} else {
		log.Printf("Found %d repos on page %d for org %s", len(repositories), repositoryListOptions.ListOptions.Page, orgName)
		for _, repo := range repositories {
			repos = append(repos, repo)
		}
		if response.NextPage != 0 {
			repositoryListOptions.ListOptions.Page++
			fetchAllRepositoriesByOrgName(orgName, repositoryListOptions)
		}
	}
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
		} else {
			initializeGithubClient(c.String("github_token"))
		}
		fetchOrganizations()
		if len(organizations) > 0 {
			for _, organization := range organizations {
				repositoryListOptions := &github.RepositoryListByOrgOptions {
					 ListOptions: *globalListOptions,
				}
				fetchAllRepositoriesByOrgName(organization.GetLogin(), repositoryListOptions)
				if len(repos) > 0 {
					for _, repository := range repos {
						pullRequests, _, err := githubClient.PullRequests.List(ctx, organization.GetLogin(), repository.GetName(), nil)
						if err != nil {
							log.Printf("Error fetching pull requests for repository '%s' - %s", repository.GetName(), err)
						} else {
							for _, pull := range pullRequests {
								fmt.Printf("Open Pull Request #%d by %s - %s\n", pull.GetNumber(), pull.GetUser().GetLogin(), pull.GetTitle())
							}
						}
					}
				}
			}
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}