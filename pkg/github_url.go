package pkg

import (
	"fmt"
	"log/slog"
	"strings"
)

type GithubURL struct {
}

func NewGithubURL() *GithubURL {
	return &GithubURL{}
}

func (e *GithubURL) Api(baseURL string) string {
	s := fmt.Sprintf("https://%s/api/v3", baseURL)
	slog.Info("API URL", "url", s)
	return s
}

func (e *GithubURL) Web(baseURL string) string {
	s := fmt.Sprintf("https://%s", baseURL)
	slog.Info("Web URL", "url", s)
	return s
}

func (e *GithubURL) CloneURL(baseURL, orgName, repoName string) string {
	s := fmt.Sprintf("%s/%s/%s.git", e.Web(baseURL), orgName, repoName)
	slog.Info("Clone URL", "url", s)
	return s
}

func (e *GithubURL) RemoteURL(baseURL, orgName, repoName, token string) string {
	host := e.Web(baseURL)
	host = strings.TrimPrefix(host, "https://")
	s := fmt.Sprintf("https://%s@%s/%s/%s.git", token, host, orgName, repoName)
	slog.Info("Remote URL", "url", s)
	return s
}
