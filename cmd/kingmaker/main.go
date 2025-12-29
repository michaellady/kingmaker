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
	"github.com/mikelady/kingmaker/internal/metadataprompt"
	"github.com/mikelady/kingmaker/internal/model"
	"github.com/mikelady/kingmaker/internal/openai"
	"github.com/mikelady/kingmaker/internal/prompt"
	"github.com/mikelady/kingmaker/internal/shorts"
	"github.com/mikelady/kingmaker/internal/youtube"
)

func main() {
	// Parse flags
	query := flag.String("query", "", "Search query for YouTube videos (required)")
	maxResults := flag.Int("max", 25, "Maximum number of videos to fetch")
	maxPrompts := flag.Int("prompts", 5, "Maximum number of prompts to generate (clips mode)")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	verbose := flag.Bool("verbose", false, "Show detailed progress")
	mode := flag.String("mode", "clips", "Mode: 'clips' for OpusClip prompts, 'metadata' for create-default prompt")
	niche := flag.String("niche", "", "Content niche for metadata mode (e.g., 'AI vibe coding')")
	includeAllVideos := flag.Bool("include-all-videos", false, "Include all videos, not just Shorts")
	flag.Parse()

	// Also accept query as positional argument
	if *query == "" && flag.NArg() > 0 {
		*query = flag.Arg(0)
	}

	if *query == "" {
		fmt.Fprintln(os.Stderr, "Usage: kingmaker -query \"your search query\"")
		fmt.Fprintln(os.Stderr, "   or: kingmaker \"your search query\"")
		fmt.Fprintln(os.Stderr, "\nModes:")
		fmt.Fprintln(os.Stderr, "  -mode clips     Generate OpusClip search prompts (default)")
		fmt.Fprintln(os.Stderr, "  -mode metadata  Generate create-default prompt for titles/descriptions")
		fmt.Fprintln(os.Stderr, "\nRequired: YOUTUBE_API_KEY environment variable")
		fmt.Fprintln(os.Stderr, "For metadata mode: OPENAI_API_KEY environment variable")
		os.Exit(1)
	}

	// Validate mode
	if *mode != "clips" && *mode != "metadata" {
		fmt.Fprintf(os.Stderr, "Error: invalid mode %q (use 'clips' or 'metadata')\n", *mode)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Set YOUTUBE_API_KEY environment variable with your YouTube Data API key")
		os.Exit(1)
	}

	// Check OpenAI API key for metadata mode
	if *mode == "metadata" && cfg.OpenAIAPIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY environment variable is required for metadata mode")
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

	var videos []model.Video

	if *includeAllVideos || *mode == "metadata" {
		// Fetch all videos (no shorts filter)
		cli.DisplayProgress(os.Stderr, fmt.Sprintf("Searching for videos: %q...", *query), cliOpts)
		// Use SearchWithDuration with no filter
		videos, err = ytClient.SearchWithDuration(ctx, *query, int64(*maxResults), youtube.DurationAny)
		if err != nil {
			cli.DisplayError(os.Stderr, fmt.Errorf("failed to fetch videos: %w", err), cliOpts)
			os.Exit(1)
		}
		cli.DisplayProgress(os.Stderr, fmt.Sprintf("Found %d videos", len(videos)), cliOpts)
	} else {
		// Fetch shorts (original behavior)
		httpClient := httpclient.NewNoRedirectClient(time.Duration(cfg.HTTPTimeout) * time.Second)
		shortsChecker := shorts.NewChecker(httpClient)
		shortsFetcher := fetcher.New(ytClient, shortsChecker)

		cli.DisplayProgress(os.Stderr, fmt.Sprintf("Searching for Shorts: %q...", *query), cliOpts)
		videos, err = shortsFetcher.FetchShorts(ctx, *query, int64(*maxResults))
		if err != nil {
			cli.DisplayError(os.Stderr, fmt.Errorf("failed to fetch Shorts: %w", err), cliOpts)
			os.Exit(1)
		}
		cli.DisplayProgress(os.Stderr, fmt.Sprintf("Found %d verified Shorts", len(videos)), cliOpts)
	}

	// Analyze patterns
	cli.DisplayProgress(os.Stderr, "Analyzing patterns...", cliOpts)
	patterns := analyzer.AnalyzeVideos(videos)

	// Handle mode-specific output
	if *mode == "metadata" {
		// Generate metadata prompt using LLM
		cli.DisplayProgress(os.Stderr, "Generating create-default prompt with LLM...", cliOpts)

		openaiClient, err := openai.NewClient(cfg.OpenAIAPIKey)
		if err != nil {
			cli.DisplayError(os.Stderr, fmt.Errorf("failed to create OpenAI client: %w", err), cliOpts)
			os.Exit(1)
		}

		gen := metadataprompt.NewGenerator(openaiClient)
		nicheStr := *niche
		if nicheStr == "" {
			nicheStr = *query // Use query as niche if not specified
		}
		opts := metadataprompt.Options{
			Niche: nicheStr,
		}

		metaPrompt, err := gen.Generate(ctx, patterns, opts)
		if err != nil {
			cli.DisplayError(os.Stderr, fmt.Errorf("failed to generate metadata prompt: %w", err), cliOpts)
			os.Exit(1)
		}

		// Display results
		fmt.Fprintln(os.Stderr) // Blank line before results
		cli.DisplayMetadataPrompt(os.Stdout, metaPrompt, patterns, cliOpts)

		// Show token usage in verbose mode
		if *verbose && !*jsonOutput {
			fmt.Fprintf(os.Stderr, "\nYouTube API quota used: %d units\n", ytClient.QuotaUsed())
			fmt.Fprintf(os.Stderr, "OpenAI tokens used: %d\n", openaiClient.TokensUsed())
		}
	} else {
		// Clips mode (original behavior)
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
}
