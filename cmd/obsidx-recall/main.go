package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	serverURL  = flag.String("server", "http://localhost:8765", "Recall server URL")
	topN       = flag.Int("top", 12, "Number of results to return")
	candidateK = flag.Int("candidates", 200, "Number of candidates to retrieve")
	jsonOutput = flag.Bool("json", false, "Output as JSON")
	verbose    = flag.Bool("verbose", true, "Show timing information")
)

type SearchRequest struct {
	Query      string `json:"query"`
	TopN       int    `json:"top_n"`
	CandidateK int    `json:"candidate_k"`
}

type SearchResponse struct {
	Results []ResultItem `json:"results"`
	Timing  TimingInfo   `json:"timing"`
	Error   string       `json:"error,omitempty"`
}

type ResultItem struct {
	Score          float32 `json:"score"`
	Path           string  `json:"path"`
	HeadingPath    string  `json:"heading_path"`
	Category       string  `json:"category"`
	Status         string  `json:"status"`
	Scope          string  `json:"scope"`
	StartLine      int     `json:"start_line"`
	EndLine        int     `json:"end_line"`
	Content        string  `json:"content"`
	CategoryWeight float32 `json:"category_weight"`
}

type TimingInfo struct {
	EmbedMs  int64 `json:"embed_ms"`
	SearchMs int64 `json:"search_ms"`
	FetchMs  int64 `json:"fetch_ms"`
	RerankMs int64 `json:"rerank_ms"`
	TotalMs  int64 `json:"total_ms"`
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <query>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	query := strings.Join(flag.Args(), " ")

	// Check if server is running
	if !isServerRunning() {
		fmt.Fprintf(os.Stderr, "❌ Recall server is not running\n\n")
		fmt.Fprintf(os.Stderr, "Start the server with:\n")
		fmt.Fprintf(os.Stderr, "  ./start-daemon.sh\n\n")
		fmt.Fprintf(os.Stderr, "Or start it now automatically? [Y/n] ")

		var response string
		fmt.Scanln(&response)

		if response == "" || strings.ToLower(response) == "y" {
			fmt.Fprintf(os.Stderr, "Starting server...\n")
			cmd := exec.Command("./start-daemon.sh")
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
				os.Exit(1)
			}
			// Give server time to start
			time.Sleep(2 * time.Second)

			// Check again
			if !isServerRunning() {
				fmt.Fprintf(os.Stderr, "Server failed to start. Check logs at .obsidian-index/recall-server.log\n")
				os.Exit(1)
			}
		} else {
			os.Exit(1)
		}
	}

	// Build request
	req := SearchRequest{
		Query:      query,
		TopN:       *topN,
		CandidateK: *candidateK,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Send request to server
	resp, err := http.Post(*serverURL+"/search", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	// Check for error
	if searchResp.Error != "" {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", searchResp.Error)
		os.Exit(1)
	}

	// Output results
	if *jsonOutput {
		printJSON(searchResp.Results)
	} else {
		printResults(query, searchResp.Results, searchResp.Timing)
	}
}

func isServerRunning() bool {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Get("http://localhost:8765/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func printResults(query string, results []ResultItem, timing TimingInfo) {
	if *verbose {
		fmt.Printf("⚡ Fast search: \"%s\"\n", query)
		fmt.Printf("⏱️  Total: %dms (embed:%dms, search:%dms, fetch:%dms, rerank:%dms)\n\n",
			timing.TotalMs, timing.EmbedMs, timing.SearchMs, timing.FetchMs, timing.RerankMs)
	}

	fmt.Printf("Found %d results:\n\n", len(results))

	for i, r := range results {
		fmt.Printf("─────────────────────────────────────────────────────────────\n")
		fmt.Printf("[%d] Score: %.4f", i+1, r.Score)

		// Add category badge
		badge := getCategoryBadge(r.Category)
		if badge != "" {
			fmt.Printf(" %s", badge)
		}
		fmt.Printf("\n")

		fmt.Printf("Path: %s\n", r.Path)
		if r.HeadingPath != "" {
			fmt.Printf("Section: %s\n", r.HeadingPath)
		}
		if r.Scope != "" {
			fmt.Printf("Scope: %s\n", r.Scope)
		}
		if r.Status != "" && r.Status != "active" {
			fmt.Printf("Status: %s\n", r.Status)
		}
		fmt.Printf("Lines: %d-%d\n", r.StartLine, r.EndLine)
		fmt.Printf("\n%s\n", excerpt(r.Content, 300))
	}
	fmt.Printf("─────────────────────────────────────────────────────────────\n")
}

func printJSON(results []ResultItem) {
	output, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(output))
}

func getCategoryBadge(category string) string {
	switch category {
	case "canon":
		return "[📚 CANON]"
	case "project":
		return "[🔨 PROJECT]"
	case "workbench":
		return "[🧪 WORKBENCH]"
	case "archive":
		return "[📦 ARCHIVE]"
	default:
		return ""
	}
}

func excerpt(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
