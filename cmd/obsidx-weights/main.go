package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sethfair/obsidx/internal/config"
)

func main() {
	var (
		configPath   = flag.String("config", ".obsidian-index/weights.json", "Path to weight configuration file")
		showDefaults = flag.Bool("defaults", false, "Show default weight configuration")
		init         = flag.Bool("init", false, "Initialize weight config file with defaults")
	)

	flag.Parse()

	if *showDefaults {
		cfg := config.DefaultWeightConfig()
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	if *init {
		// Check if file exists
		if _, err := os.Stat(*configPath); err == nil {
			fmt.Fprintf(os.Stderr, "Config file already exists: %s\n", *configPath)
			fmt.Fprintf(os.Stderr, "Remove it first if you want to reinitialize\n")
			os.Exit(1)
		}

		cfg := config.DefaultWeightConfig()
		if err := config.SaveWeightConfig(cfg, *configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Created weight configuration: %s\n", *configPath)
		fmt.Println("\nDefault weights configured for:")
		fmt.Println("  • Zettelkasten tags (#permanent-note, #literature-note, #fleeting-notes)")
		fmt.Println("  • PARA tags (#writerflow, #mvp-1, #archive, etc.)")
		fmt.Println("  • Writerflow tags (#customer-research, #vision, #icp, etc.)")
		fmt.Println("\nEdit this file to customize weights for your workflow!")
		return
	}

	// Default: show current config
	cfg, err := config.LoadWeightConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		fmt.Println("# Using default weights (no config file found)")
		fmt.Printf("# Run with --init to create: %s\n\n", *configPath)
	} else {
		fmt.Printf("# Weight configuration from: %s\n\n", *configPath)
	}

	fmt.Println(string(data))
}
