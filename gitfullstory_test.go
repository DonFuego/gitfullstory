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

	if repos == nil {
		t.Error("Repositories should not be empty")
	} else {
		t.Logf("fetchAllRepositoriesByOrgName -> has values: %d", len(repos))
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