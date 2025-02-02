package main

import (
	"log/slog"
	"os"
	"sync"

	"github.com/kevincobain2000/bpr/pkg"
	"github.com/manifoldco/promptui"
)

var (
	flags   pkg.Flags
	version = "dev"
)

func main() {
	pkg.ParseFlags(&flags)
	pkg.SetupLoggingStdout(flags) // nolint: errcheck

	if err := pkg.ValidateFlags(&flags, version); err != nil {
		slog.Error("Flag validation failed", "error", err)
		os.Exit(1)
	}

	githubHandler, err := pkg.NewGithubHandler(flags)
	if err != nil {
		slog.Error("Failed to create GitHub client", "error", err)
		return
	}

	repos, err := githubHandler.GetRepos()
	if err != nil {
		slog.Error("Failed to fetch repositories", "error", err)
		return
	}

	selectedRepos := pkg.NewSlices().FilterByCSV(repos, flags.Repos)
	if selectedRepos == nil {
		slog.Warn("No repositories were selected")
		return
	}
	for _, repo := range selectedRepos {
		slog.Info("Selected repository", "repo", repo)
	}

	if !flags.NoInteractive {
		if !confirmAction("Proceed with the selected repositories?") {
			slog.Info("Operation aborted by user")
			return
		}
	}

	processParallel(selectedRepos)
}

func confirmAction(message string) bool {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil || result != "y" {
		return false
	}
	return true
}

func processParallel(repos []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, flags.Parallel)

	for _, repo := range repos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			handler, err := pkg.NewGithubHandler(flags)
			if err != nil {
				slog.Error("Failed to initialize GitHub handler", "repo", repo, "error", err)
				slog.Info("Use --log-level=-4 for more details")
				return
			}

			if err := handler.Handle(repo); err != nil {
				slog.Error("Repository processing failed", "repo", repo, "error", err)
				slog.Info("Use --log-level=-4 for more details")
			}
		}(repo)
	}

	wg.Wait()
}
