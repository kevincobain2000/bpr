package pkg

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func setupGithubHandler() *GithubHandler {
	flags := Flags{
		GithubToken:   "test-token",
		BaseURL:       "https://api.github.com/",
		Org:           "test-org",
		PRTitle:       "Test PR",
		PRBody:        "This is a test PR",
		PRBranch:      "test-branch",
		DefaultBranch: "main",
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: flags.GithubToken})
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)

	return &GithubHandler{
		flags:        flags,
		githubClient: githubClient,
		ctx:          ctx,
	}
}

func TestGetRepos(t *testing.T) {
	defer gock.Off()
	h := setupGithubHandler()

	// Mock the organization check API request
	gock.New("https://api.github.com").
		Get("/orgs/test-org").
		Reply(http.StatusOK).
		JSON(map[string]interface{}{"login": "test-org"})

	// Mock the repositories API request
	gock.New("https://api.github.com").
		Get("/orgs/test-org/repos").
		Reply(http.StatusOK).
		JSON([]map[string]interface{}{
			{"name": "repo1"},
			{"name": "repo2"},
		})

	repos, err := h.GetRepos()
	assert.NoError(t, err)
	assert.Equal(t, []string{"repo1", "repo2"}, repos)
}

func TestCreatePR(t *testing.T) {
	defer gock.Off()
	h := setupGithubHandler()

	gock.New("https://api.github.com").
		Post("/repos/test-org/test-repo/pulls").
		Reply(http.StatusCreated).
		JSON(map[string]interface{}{
			"html_url": "https://github.com/test-org/test-repo/pull/1",
		})

	err := h.createPR("test-repo")
	assert.NoError(t, err)
}

func TestIsOrganization(t *testing.T) {
	defer gock.Off()
	h := setupGithubHandler()

	gock.New("https://api.github.com").
		Get("/orgs/test-org").
		Reply(http.StatusOK).
		JSON(map[string]interface{}{"login": "test-org"})

	isOrg, err := h.isOrganization("test-org")
	assert.NoError(t, err)
	assert.True(t, isOrg)
}

func TestSetHeadBranch(t *testing.T) {
	defer gock.Off()
	h := setupGithubHandler()

	gock.New("https://api.github.com").
		Get("/repos/test-org/test-repo").
		Reply(http.StatusOK).
		JSON(map[string]interface{}{"default_branch": "main"})

	err := h.setHeadBranch("test-org", "test-repo")
	assert.NoError(t, err)
}
