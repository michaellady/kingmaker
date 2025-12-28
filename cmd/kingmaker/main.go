package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mikelady/kingmaker/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Kingmaker initialized with max results: %d\n", cfg.MaxResults)
}
