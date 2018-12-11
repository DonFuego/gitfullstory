package main

import (
	"github.com/google/go-github/github"
	"log"
	"os"
	"testing"
)

// setup gets called before test cases are run and ensures we have an initialized github api client
func setup() {
	initializeGithubClient(os.Getenv("GITHUB_TOKEN"))
}

// TestFetchAllRepositoriesByOrgName tests that we can appropriately retrieve all repos for an organization
func TestFetchAllRepositoriesByOrgName(t *testing.T) {
	repositoryListOptions := &github.RepositoryListByOrgOptions {
		ListOptions: *globalListOptions,
	}

	fetchRepositoriesByOrgName("Hearst-Hatchery", repositoryListOptions)

	if repos == nil {
		t.Error("Repositories should not be empty")
	} else {
		t.Logf("fetchRepositoriesByOrgName -> has values: %d", len(repos))
	}
}

// TestParseOrgsFromCommandline
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

func TestParseProjectsFromCommandline(t *testing.T) {
	projectTest1 := "a,b"

	parseProjectsFromCommandline(projectTest1)

	found, ok := projectsMap["a"]
	if !ok  {
		t.Error("element not found in map, should contain 'a' element!")
	} else {
		t.Logf("TestParseProjectsFromCommandline -> has the value we were looking for: '%s' while projectsMap has %s", found, projectsMap)
	}

	_, notOk := projectsMap["c"]
	if notOk {
		t.Error("element was found in map, should not contain 'c' element!")
	} else {
		t.Logf("TestParseProjectsFromCommandline -> has no value we were looking for: 'c' while projectsMap has: %s", projectsMap)
	}
}

// shutdown should do any cleanup as needed once test(s) are complete
func shutdown() {
	log.Printf("Testing Complete!  Shutting down...")
}

// TestMain allows for us to include setup/teardown methods to kick off our tests. This is used to initialize
// the github token and thus api client
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}