package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli"
)

// ctx holds our github api context
var ctx context.Context
// githubClient is our authorized api client to pull data from github
var githubClient *github.Client

// organizations stores a list of orgs a user is a part of
var organizations []*github.Organization

// holds parsed commandline orgs that are sliced and filtered for empty string
var orgs = make([]string, 0)

// fetchOrgsError holds any errors while grabbing the user's list of orgs
var fetchOrgsError error
// repositories holds the list of repositories for all the orgs - we recursively append to this
// slice due to api request per page limits
var repositories = make([]*github.Repository, 0)

// globalListOptions is the setting used for all github api list functions
// keep in mind, github api returns default of 30 and max of 100 so you must use pagination for > 100
// this is used ot recursively loop through repositories at the moment
var globalListOptions = &github.ListOptions {
	PerPage: 100,
	Page: 1,
}

// initializeGithubClient takes the user's passed in personal github token and creates an oauth2
// github client in order to programmatically access the github api.
func initializeGithubClient(accessToken string) {
	ctx = context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tokenClient := oauth2.NewClient(ctx, tokenSource)
	githubClient = github.NewClient(tokenClient)
}

// fetchAllOrganizations uses the github api client to retrieve all organizations for the user
func fetchAllOrganizations() {
	organizations, _, fetchOrgsError = githubClient.Organizations.List(ctx, "", nil)
	// if we can't get the orgs, then we ultimately can't get repos for the team/user
	if fetchOrgsError != nil {
		log.Fatalf("Error retrieving organizations from Github!  Reason: %s", fetchOrgsError)
	} else {
		for _, org := range organizations {
			orgs = append(orgs, org.GetLogin())
		}
	}
}

// parseOrgsFromCommandline parses the comma separated list of organizations and sets them for the cli app
func parseOrgsFromCommandline(orgToken string) {
	orgs = strings.Split(orgToken, ",")
	orgs = filterEmpty(orgs)
}

// filterEmpty removes empty values from a slice - turns out more performant than strings.FieldsFunc.
func filterEmpty(strValue []string) []string {
	var filtered []string
	for _, str := range strValue {
		if str != "" {
			filtered = append(filtered, str)
		}
	}
	return filtered
}

// fetchAllRepositoriesByOrgName retrieves all the repositories for a given organization.  If there are more than 100 returned,
// then it will recursively call again and bump up the page number.
func fetchAllRepositoriesByOrgName(orgName string, repositoryListOptions *github.RepositoryListByOrgOptions) {
	log.Printf("Fetching repositories for org: %s on page %d", orgName, repositoryListOptions.ListOptions.Page)
	repositories, response, err := githubClient.Repositories.ListByOrg(ctx, orgName, repositoryListOptions)

	if err  != nil {
		log.Fatalf("Error fetching repositories for organization %s.  Error: %s", orgName, err)
	} else {
		log.Printf("Found %d repositories on page %d for org %s", len(repositories), repositoryListOptions.ListOptions.Page, orgName)
		for _, repo := range repositories {
			repositories = append(repositories, repo)
		}
		if response.NextPage != 0 {
			repositoryListOptions.ListOptions.Page++
			fetchAllRepositoriesByOrgName(orgName, repositoryListOptions)
		}
	}
}


// main kicks off the cli and parses incoming flags and commands
func main() {
	app := cli.NewApp()

	app.Name = "gitfullstory"
	app.Usage = "command-line utility for seeing your team's open pull requests"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "github_token, gt",
			Value: "",
			Usage: "your personal access token for github",
			EnvVar: "GITHUB_ACCESS_TOKEN,GITHUB_TOKEN",
		},
		cli.StringFlag {
			Name: "orgs",
			Value: "",
			Usage: "comma separated list of github organizations to filter pull requests on otherwise it fetches for all",
			EnvVar: "GITFULLSTORY_ORGS",
		},
		cli.StringFlag {
			Name: "projects",
			Value: "",
			Usage: "comma separated list of github projects to filter pull requests on otherwise it fetches for all",
			EnvVar: "GITFULLSTORY_PROJECTS",
		},
		cli.StringFlag {
			Name: "users",
			Value: "",
			Usage: "comma separated list of github users to filter pull requests on otherwise it fetches for all",
			EnvVar: "GITFULLSTORY_USERS",
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	// process the command line arguments and execute the program
	app.Action = func(c *cli.Context) error {
		if c.String("github_token") == "" {
			return cli.NewExitError("Error, missing github personal access token!", 86)
		} else {
			initializeGithubClient(c.String("github_token"))
		}

		// Either fetch all orgs from github or parse the orgs from the commandline
		if c.String("orgs") == "" {
			fetchAllOrganizations()
		} else {
			parseOrgsFromCommandline(c.String("orgs"))
		}

		log.Printf("Searching over github orgs: %s", orgs)

		if len(orgs) > 0 {
			for _, orgName := range orgs {
				repositoryListOptions := &github.RepositoryListByOrgOptions {
					 ListOptions: *globalListOptions,
				}
				fetchAllRepositoriesByOrgName(orgName, repositoryListOptions)
				if len(repositories) > 0 {
					for _, repository := range repositories {
						pullRequests, _, err := githubClient.PullRequests.List(ctx, orgName, repository.GetName(), nil)
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