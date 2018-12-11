package main

import (
	"github.com/google/go-github/github"
	"log"
	"os"
	"testing"
)

//var globalListOptions = &github.ListOptions {
//	PerPage: 100,
//	Page: 1,
//}

func setup() {
	initializeGithubClient(os.Getenv("GITHUB_TOKEN"))
}

func TestFetchAllRepositoriesByOrgName(t *testing.T) {
	repositoryListOptions := &github.RepositoryListByOrgOptions {
		ListOptions: *globalListOptions,
	}

	fetchAllRepositoriesByOrgName("Hearst-Hatchery", repositoryListOptions)

	if repositories == nil {
		t.Error("Repositories should not be empty")
	} else {
		t.Logf("fetchAllRepositoriesByOrgName -> has values: %d", len(repositories))
	}
}

func TestParseOrgsFromCommandline(t *testing.T) {
	orgTest1 := "ideo, HearstAuto"
	parseOrgsFromCommandline(orgTest1)

	if orgs == nil || len(orgs) <= 0 {
		t.Error("orgs should not be empty!")
	} else {
		t.Logf("parseOrgsFromCommandline -> has values: %s", orgs)
	}

	orgs = orgs[:0]

	orgTest2 := ""
	parseOrgsFromCommandline(orgTest2)

	if len(orgs) > 1 {
		t.Error("orgs should be empty!")
	} else {
		t.Logf("parseOrgsFromCommandline -> has no values: %s", orgs)
	}

	orgs = orgs[:0]

	orgTest3 := "ideo,,,,"
	parseOrgsFromCommandline(orgTest3)

	if len(orgs) > 1 {
		t.Error("orgs should contain only 1 element!")
	} else {
		t.Logf("parseOrgsFromCommandline -> has 1 value: %s", orgs)
	}
}

func shutdown() {
	log.Printf("Testing Complete!  Shutting down...")
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}