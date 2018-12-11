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

// holds parsed commandline orgs, projects and users that are sliced and filtered for empty string
var orgs = make([]string, 0)
var projectsMap = make(map[string]string)
var usersMap = make(map[string]string)

// repos holds the list of repos for all the orgs - we recursively append to this
// slice due to api request per page limits
var repos = make([]*github.Repository, 0)

// globalListOptions is the setting used for all github api list functions
// keep in mind, github api returns default of 30 and max of 100 so you must use pagination for > 100
// this is used ot recursively loop through repos at the moment
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
	organizations, _, err := githubClient.Organizations.List(ctx, "", nil)
	// if we can't get the orgs, then we ultimately can't get repos for the team/user
	if err != nil {
		log.Fatalf("Error retrieving organizations from Github!  Reason: %s", err)
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

// parseProjectsFromCommandline parses the comma separated list of projects and set them for the cli app
func parseProjectsFromCommandline(projectsToken string) {
	projects := filterEmpty(strings.Split(projectsToken, ","))

	// put project values into map
	for _, project := range projects {
		projectsMap[project] = project
	}
}

// parseUsersFromCommandline parses the comma separated list of users and set them for the cli app
func parseUsersFromCommandline(usersToken string) {
	users := filterEmpty(strings.Split(usersToken, ","))

	// put user values into map
	for _, user := range users {
		usersMap[user] = user
	}
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

// fetchRepositoriesByOrgName retrieves all the repos for a given organization.  If there are more than 100 returned,
// then it will recursively call again and bump up the page number.
func fetchRepositoriesByOrgName(orgName string, repositoryListOptions *github.RepositoryListByOrgOptions) {
	//log.Printf("Fetching repos for org: %s on page %d", orgName, repositoryListOptions.ListOptions.Page)
	repositories, response, err := githubClient.Repositories.ListByOrg(ctx, orgName, repositoryListOptions)

	if err  != nil {
		log.Fatalf("Error fetching repos for organization %s.  Error: %s", orgName, err)
	} else {
		//log.Printf("Found %d repos on page %d for org %s", len(repositories), repositoryListOptions.ListOptions.Page, orgName)
		for _, repo := range repositories {
			if len(projectsMap) > 0 {
				if _, exist := projectsMap[repo.GetName()]; exist {
					repos = append(repos, repo)
				}
			} else {
				repos = append(repos, repo)
			}
		}
		if response.NextPage != 0 {
			repositoryListOptions.ListOptions.Page++
			fetchRepositoriesByOrgName(orgName, repositoryListOptions)
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

		log.Printf("Searching for open PR's within github orgs: %s\n", orgs)

		if c.String("projects") != "" {
			parseProjectsFromCommandline(c.String("projects"))
			log.Printf("...for only projects: %s\n", projectsMap)
		}

		if c.String("users") != "" {
			parseUsersFromCommandline(c.String("users"))
			log.Printf("...by users: %s \n", usersMap)
		}

		if len(orgs) > 0 {
			for _, orgName := range orgs {
				repositoryListOptions := &github.RepositoryListByOrgOptions {
					 ListOptions: *globalListOptions,
				}
				fetchRepositoriesByOrgName(orgName, repositoryListOptions)
				//log.Printf("We are going to get pull requests for %s repos", repos)
				if len(repos) > 0 {
					for _, repository := range repos {
						pullRequests, _, err := githubClient.PullRequests.List(ctx, orgName, repository.GetName(), nil)
						if err != nil {
							log.Printf("Error fetching pull requests for repository '%s' - %s", repository.GetName(), err)
						} else {
							for _, pull := range pullRequests {
								if _, exist := usersMap[ pull.GetUser().GetLogin()]; exist {
									fmt.Printf("Open Pull Request #%d by %s - %s\n", pull.GetNumber(), pull.GetUser().GetLogin(), pull.GetTitle())
								}
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