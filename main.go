package main

import (
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/kevincobain2000/bpr/pkg"
	"github.com/manifoldco/promptui"
)

var f pkg.Flags
var version = "dev"

func main() {
	pkg.ParseFlags(&f)
	pkg.SetupLoggingStdout(f)
	if err := pkg.ValidateFlags(&f); err != nil {
		slog.Error("Failed to validate flags:", "error", err)
		os.Exit(1)
	}
	wantsVersion()

	gh, err := pkg.NewGithubHandler(f)
	if err != nil {
		slog.Error("Failed to create GitHub Enterprise client:", "error", err)
		return
	}
	repos, err := gh.GetRepos()

	if err != nil {
		slog.Error("Failed to fetch repositories:", "error", err)
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, f.Parallel)
	selectedRepos := filterSliceByCSV(repos, f.Repos)
	for _, repo := range selectedRepos {
		slog.Info("Selected", "repo", repo)
	}
	//promptui are you sure?
	prompt := promptui.Prompt{
		Label:     "Are you sure you want to continue?",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil {
		slog.Warn("Exiting without processing")
		return
	}
	if result != "y" {
		slog.Info("Exiting")
		return
	}

	for _, repo := range selectedRepos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			ghh, err := pkg.NewGithubHandler(f)
			if err != nil {
				slog.Error("Failed to create GitHub Enterprise client:", "repo", repo, "error", err)
				slog.Info("Run --log-level=-4 for more details")
				return
			}
			err = ghh.Handle(repo)
			if err != nil {
				slog.Error("Failed to handle repository:", "repo", repo, "error", err)
				slog.Info("Run --log-level=-4 for more details")
				return
			}
		}(repo)
	}

	wg.Wait()
}

func wantsVersion() {
	if f.Version {
		slog.Info("Version", "version", version)
		os.Exit(0)
	}
}

func filterSliceByCSV(slice []string, csv string) []string {
	if csv == "" {
		return slice
	}

	csvMap := make(map[string]struct{})
	for _, item := range strings.Split(csv, ",") {
		csvMap[strings.TrimSpace(item)] = struct{}{}
	}

	// Filter the slice
	var result []string
	for _, item := range slice {
		if _, exists := csvMap[item]; exists {
			result = append(result, item)
		}
	}

	return result
}
