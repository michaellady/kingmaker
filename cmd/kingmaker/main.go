package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mikelady/kingmaker/internal/analyzer"
	"github.com/mikelady/kingmaker/internal/cli"
	"github.com/mikelady/kingmaker/internal/config"
	"github.com/mikelady/kingmaker/internal/fetcher"
	"github.com/mikelady/kingmaker/internal/httpclient"
	"github.com/mikelady/kingmaker/internal/prompt"
	"github.com/mikelady/kingmaker/internal/shorts"
	"github.com/mikelady/kingmaker/internal/youtube"
)

func main() {
	// Parse flags
	query := flag.String("query", "", "Search query for YouTube Shorts (required)")
	maxResults := flag.Int("max", 25, "Maximum number of videos to fetch")
	maxPrompts := flag.Int("prompts", 5, "Maximum number of prompts to generate")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	verbose := flag.Bool("verbose", false, "Show detailed progress")
	flag.Parse()

	// Also accept query as positional argument
	if *query == "" && flag.NArg() > 0 {
		*query = flag.Arg(0)
	}

	if *query == "" {
		fmt.Fprintln(os.Stderr, "Usage: kingmaker -query \"your search query\"")
		fmt.Fprintln(os.Stderr, "   or: kingmaker \"your search query\"")
		fmt.Fprintln(os.Stderr, "\nRequired: YOUTUBE_API_KEY environment variable")
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Set YOUTUBE_API_KEY environment variable with your YouTube Data API key")
		os.Exit(1)
	}

	// CLI options
	cliOpts := cli.Options{
		JSON:        *jsonOutput,
		ShowSummary: true,
		Verbose:     *verbose,
	}

	// Initialize clients
	ctx := context.Background()

	cli.DisplayProgress(os.Stderr, "Initializing YouTube client...", cliOpts)
	ytClient, err := youtube.NewClient(cfg.YouTubeAPIKey)
	if err != nil {
		cli.DisplayError(os.Stderr, fmt.Errorf("failed to create YouTube client: %w", err), cliOpts)
		os.Exit(1)
	}

	httpClient := httpclient.NewNoRedirectClient(time.Duration(cfg.HTTPTimeout) * time.Second)
	shortsChecker := shorts.NewChecker(httpClient)
	shortsFetcher := fetcher.New(ytClient, shortsChecker)

	// Fetch shorts
	cli.DisplayProgress(os.Stderr, fmt.Sprintf("Searching for Shorts: %q...", *query), cliOpts)
	videos, err := shortsFetcher.FetchShorts(ctx, *query, int64(*maxResults))
	if err != nil {
		cli.DisplayError(os.Stderr, fmt.Errorf("failed to fetch Shorts: %w", err), cliOpts)
		os.Exit(1)
	}

	cli.DisplayProgress(os.Stderr, fmt.Sprintf("Found %d verified Shorts", len(videos)), cliOpts)

	// Analyze patterns
	cli.DisplayProgress(os.Stderr, "Analyzing patterns...", cliOpts)
	patterns := analyzer.AnalyzeVideos(videos)

	// Generate prompts
	cli.DisplayProgress(os.Stderr, "Generating OpusClip prompts...", cliOpts)
	promptOpts := prompt.Options{
		MaxPrompts: *maxPrompts,
		Query:      *query,
	}
	prompts := prompt.Generate(patterns, promptOpts)

	// Display results
	fmt.Fprintln(os.Stderr) // Blank line before results
	cli.DisplayResults(os.Stdout, patterns, prompts, cliOpts)

	// Show quota usage in verbose mode
	if *verbose && !*jsonOutput {
		fmt.Fprintf(os.Stderr, "\nAPI quota used: %d units\n", ytClient.QuotaUsed())
	}
}
