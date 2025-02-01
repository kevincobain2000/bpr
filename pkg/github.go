package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

type GithubHandler struct {
	flags        Flags
	githubClient *github.Client
	ctx          context.Context
}

func NewGithubHandler(flags Flags) (*GithubHandler, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: flags.GithubToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient, err := github.NewEnterpriseClient(NewGithubURL().Api(flags.BaseURL), "", tc)
	if err != nil {
		return nil, err
	}
	return &GithubHandler{
		flags:        flags,
		githubClient: githubClient,
		ctx:          ctx,
	}, nil
}

func (h *GithubHandler) Handle(repo string) error {
	repoDir, err := h.cloneRepo(h.flags.Org, repo)
	if err != nil {
		return err
	}
	defer os.RemoveAll(repoDir)

	err = h.execCommand(repoDir, h.flags.Cmd)
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	err = h.createBranch(repoDir, h.flags.PRBranch)
	if err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	err = h.commitChanges(repoDir)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	err = h.setHeadBranch(h.flags.Org, repo)
	if err != nil {
	}

	if h.flags.Dry {
		slog.Info("Dry run enabled, skipping push and PR creation")
	}
	if !h.flags.Dry {
		err = h.pushBranch(repoDir, repo, h.flags.PRBranch, h.flags.GithubToken)
		if err != nil {
			return fmt.Errorf("failed to push branch: %v", err)
		}
	}

	if !h.flags.Dry {
		err = h.createPR(repo)
		if err != nil {
			return fmt.Errorf("failed to create PR: %v", err)
		}
	}
	return nil
}

func (h *GithubHandler) GetRepos() ([]string, error) {
	slog.Info("Fetching repositories", "org", h.flags.Org)
	var repoNames []string

	// Check if the provided name is an organization or a user
	isOrg, err := h.isOrganization(h.flags.Org)
	if err != nil {
		return nil, fmt.Errorf("failed to determine if %s is an organization: %v", h.flags.Org, err)
	}

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		var repos []*github.Repository
		var resp *github.Response

		if isOrg {
			repos, resp, err = h.githubClient.Repositories.ListByOrg(h.ctx, h.flags.Org, &github.RepositoryListByOrgOptions{
				ListOptions: opt.ListOptions,
			})
		} else {
			repos, resp, err = h.githubClient.Repositories.List(h.ctx, h.flags.Org, opt)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %v", err)
		}

		for _, repo := range repos {
			repoNames = append(repoNames, *repo.Name)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return repoNames, nil
}

// Helper function to check if the given name is an organization
func (h *GithubHandler) isOrganization(name string) (bool, error) {
	org, _, err := h.githubClient.Organizations.Get(h.ctx, name)
	if err != nil {
		// If the error is because the name is not an organization, assume it's a user
		if _, ok := err.(*github.ErrorResponse); ok {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if %s is an organization: %v", name, err)
	}
	return org != nil, nil
}

func (h *GithubHandler) cloneRepo(org, repo string) (string, error) {
	repoDir := os.TempDir() + org + "/" + repo
	slog.Info("Cloning repository", "repo", repo, "dir", repoDir)

	if _, err := os.Stat(repoDir); err == nil {
		if err := os.RemoveAll(repoDir); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %v", err)
		}
	}

	cmd := exec.Command("git", "clone", NewGithubURL().CloneURL(h.flags.BaseURL, h.flags.Org, repo), repoDir)
	if h.flags.LogLevel == LogLevelDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %v", err)
	}

	return repoDir, nil
}

func (h *GithubHandler) execCommand(repoDir, cmd string) error {
	slog.Info("Executing command", "cmd", cmd)
	command := exec.Command("sh", "-c", cmd)
	command.Dir = repoDir
	if h.flags.LogLevel == LogLevelDebug {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	return command.Run()
}

func (h *GithubHandler) createBranch(repoDir, branchName string) error {
	slog.Info("Creating branch", "branch", branchName)
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoDir
	if h.flags.LogLevel == LogLevelDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func (h *GithubHandler) commitChanges(repoDir string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if h.flags.LogLevel == LogLevelDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add changes: %v", err)
	}

	slog.Info("Committing changes", "msg", h.flags.PRCommitMsg)
	cmd = exec.Command("git", "commit", "-m", h.flags.PRCommitMsg)
	cmd.Dir = repoDir
	if h.flags.LogLevel == LogLevelDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func (h *GithubHandler) pushBranch(repoDir, repo, branchName, token string) error {
	slog.Info("Pushing changes", "branch", branchName, "org", h.flags.Org, "repo", repo)
	remoteURL := NewGithubURL().RemoteURL(h.flags.BaseURL, h.flags.Org, repo, token)

	cmdSetRemote := exec.Command("git", "remote", "set-url", "origin", remoteURL)
	cmdSetRemote.Dir = repoDir
	if h.flags.LogLevel == LogLevelDebug {
		cmdSetRemote.Stdout = os.Stdout
		cmdSetRemote.Stderr = os.Stderr
	}

	if err := cmdSetRemote.Run(); err != nil {
		return fmt.Errorf("failed to set remote URL: %v", err)
	}

	cmdPush := exec.Command("git", "push", "origin", branchName)
	cmdPush.Dir = repoDir
	if h.flags.LogLevel == LogLevelDebug {
		cmdPush.Stdout = os.Stdout
		cmdPush.Stderr = os.Stderr
	}

	if err := cmdPush.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %v", err)
	}

	return nil
}

func (h *GithubHandler) createPR(repo string) error {
	slog.Info("Creating pull request",
		"title", h.flags.PRTitle,
		"body", h.flags.PRBody,
		"head", h.flags.DefaultBranch,
		"base", h.flags.PRBranch,
	)

	// Ensure the head and base branches are correctly specified
	if h.flags.DefaultBranch == h.flags.PRBranch {
		return fmt.Errorf("head branch (%s) and base branch (%s) cannot be the same", h.flags.DefaultBranch, h.flags.PRBranch)
	}

	// Create the new pull request
	newPR := &github.NewPullRequest{
		Title: &h.flags.PRTitle,
		Base:  &h.flags.DefaultBranch, // The branch with changes
		Head:  &h.flags.PRBranch,      // The branch to merge into
		Body:  &h.flags.PRBody,
	}

	// Create the pull request using the GitHub API
	pr, _, err := h.githubClient.PullRequests.Create(h.ctx, h.flags.Org, repo, newPR)
	if err != nil {
		return fmt.Errorf("failed to create PR: %v", err)
	}

	slog.Info("Pull request created successfully", "url", pr.GetHTMLURL())
	return nil
}

func (h *GithubHandler) setHeadBranch(org, repo string) error {
	defaultBranch, _, err := h.githubClient.Repositories.Get(h.ctx, org, repo)
	if err != nil {
		return err
	}

	if h.flags.DefaultBranch == "" {
		h.flags.DefaultBranch = defaultBranch.GetDefaultBranch()
	}
	if h.flags.DefaultBranch != defaultBranch.GetDefaultBranch() {
		slog.Warn("Given default is not same as remote default", "default", h.flags.DefaultBranch, "default", defaultBranch.GetDefaultBranch())
	}
	slog.Info("Using default branch as", "default", h.flags.DefaultBranch)
	if h.flags.DefaultBranch == "" {
		return fmt.Errorf("default branch is empty")
	}
	return nil
}
