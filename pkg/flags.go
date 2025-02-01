package pkg

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	g "github.com/kevincobain2000/go-human-uuid/lib"
)

type Flags struct {
	// required
	GithubToken string // set by env var GITHUB_TOKEN
	Org         string
	Cmd         string

	// optional
	Repos         string // comma-separated list of repositories
	BaseURL       string // default: github.com
	DefaultBranch string // default: taken from api
	PRTitle       string
	PRBody        string
	PRBranch      string
	PRCommitMsg   string
	Parallel      int
	Dry           bool
	LogLevel      int
	Version       bool
}

const (
	githubTokenEnv  = "GITHUB_TOKEN" // nolint: gosec
	defaultBaseURL  = "github.com"
	defaultLogLevel = LogLevelInfo
	defaultParallel = 10

	defaultPRBranch = "bpr-<random>"
	defaultPRTitle  = "BPR: bulk PR changes"
	defaultPRBody   = "BPR: bulk PR changes"
	defaultPRCommit = "BPR: bulk PR changes"
)

func ParseFlags(f *Flags) {
	flag.StringVar(&f.GithubToken, "token", "", "GITHUB_TOKEN via env or flag")
	flag.StringVar(&f.Org, "org", "", "GitHub organization name or your own username (required)")
	flag.StringVar(&f.Cmd, "cmd", "", "action command to run (required)")

	flag.StringVar(&f.Repos, "repos", "", "comma-separated list of repositories (empty for all)")
	flag.StringVar(&f.BaseURL, "base-url", defaultBaseURL, "GitHub base URL")
	flag.StringVar(&f.DefaultBranch, "default-branch", "", "head branch")
	flag.StringVar(&f.PRTitle, "pr-title", defaultPRTitle, "pull request title")
	flag.StringVar(&f.PRBody, "pr-body", defaultPRBody, "pull request body")

	flag.StringVar(&f.PRBranch, "pr-branch", defaultPRBranch, "pull request branch")
	flag.StringVar(&f.PRCommitMsg, "pr-commit-msg", defaultPRCommit, "pull request commit message")
	flag.IntVar(&f.Parallel, "parallel", defaultParallel, "number of parallel requests")
	flag.BoolVar(&f.Dry, "dry", false, "dry run")

	flag.IntVar(&f.LogLevel, "log-level", defaultLogLevel, "log level (0=info, -4=debug, 4=warn, 8=error)")
	flag.BoolVar(&f.Version, "version", false, "print version and exit")

	flag.Parse()
}

func ValidateFlags(f *Flags) error {
	if f.GithubToken == "" {
		f.GithubToken = os.Getenv(githubTokenEnv)
	}
	var err error
	f.BaseURL, err = extractHost(f.BaseURL)
	if err != nil {
		return err
	}
	if f.BaseURL == "" {
		return fmt.Errorf("invalid base URL (--base-url)")
	}
	if f.GithubToken == "" {
		return fmt.Errorf("missing GitHub token (--token)")
	}
	if f.Org == "" {
		return fmt.Errorf("missing organization name (--org)")
	}
	if f.Cmd == "" {
		return fmt.Errorf("missing exec command (--cmd)")
	}
	if f.PRTitle == "" {
		return fmt.Errorf("missing pull request title (--pr-title)")
	}
	if f.PRBody == "" {
		return fmt.Errorf("missing pull request body (--pr-body)")
	}
	if f.PRBranch == "" {
		return fmt.Errorf("missing pull request branch (--pr-branch)")
	}
	if f.PRCommitMsg == "" {
		return fmt.Errorf("missing pull request commit message (--pr-commit-msg)")
	}
	if f.Parallel < 1 {
		return fmt.Errorf("invalid parallel value (--parallel)")
	}
	if f.LogLevel != LogLevelInfo && f.LogLevel != LogLevelDebug && f.LogLevel != LogLevelWarn && f.LogLevel != LogLevelError {
		return fmt.Errorf("invalid log level (--log-level)")
	}

	if f.PRBranch == defaultPRBranch {
		gen, err := g.NewGenerator([]g.Option{
			func(opt *g.Options) error {
				opt.Length = 4
				return nil
			},
		}...)
		if err != nil {
			return err
		}
		f.PRBranch = strings.ReplaceAll(defaultPRBranch, "<random>", gen.Generate())
	}

	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "token" {
			slog.Info("CLI", "name", f.Name, "value", "********")
			return
		}
		if f.Name == "version" || f.Name == "log-level" || f.Name == "parallel" {
			return
		}
		if f.Name == "default-branch" && f.Value.String() == "" {
			slog.Info("CLI", "name", f.Name, "value", "default")
			return
		}
		slog.Info("CLI", "name", f.Name, "value", f.Value)
	})

	return nil
}

func extractHost(fullURL string) (string, error) {
	if !strings.HasPrefix(fullURL, "http") {
		fullURL = "https://" + fullURL
	}
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}

	return parsedURL.Host, nil
}
