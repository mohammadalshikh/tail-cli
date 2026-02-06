package cmd

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"

	"github.com/mohammadalshikh/tail-cli/internal/agent"
	"github.com/mohammadalshikh/tail-cli/internal/analyzer"
	"github.com/mohammadalshikh/tail-cli/internal/ingestor"
	"github.com/mohammadalshikh/tail-cli/internal/tui"
	"github.com/spf13/cobra"
)

var (
	filePath string
	noTUI    bool
	topK     int
	noAI     bool
)

var analyzeCmd = &cobra.Command{
	Use: "analyze",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && args[0] == "help" {
			cmd.Help()
			os.Exit(0)
		}
		if filePath == "" {
			return fmt.Errorf("Error: required flag <file> not set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		tracker := analyzer.NewTracker(topK)

		fmt.Printf("\nAnalyzing %s...\n", filePath)
		err := ingestor.ProcessFile(filePath, tracker)
		if err != nil {
			fmt.Printf("Error (reading file): %v\n", err)
			return
		}

		var client *agent.OpenAIClient
		if !noAI {
			var err error
			client, err = agent.NewOpenAIClient()
			if err != nil {
				fmt.Printf("Warning: AI analysis disabled (%v)\n", err)
				noAI = true
			}
		}

		if !noTUI {
			err := tui.Run(tracker, client, noAI)
			if err != nil {
				fmt.Printf("Error running TUI: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if noAI {
			fmt.Printf("\n--- Top %d P99 outliers ---\n", topK)
			count := 1
			for tracker.Heap.Len() > 0 {
				item := heap.Pop(tracker.Heap)
				entry := item.(analyzer.LogEntry)
				fmt.Printf("\n#%d [%vms] %s\n", count, entry.Latency, entry.Data)
				count++
			}
			return
		}

		fmt.Printf("\n--- Top %d P99 outliers with AI analysis ---\n", topK)
		scanner := bufio.NewScanner(os.Stdin)
		count := 1
		totalEntries := tracker.Heap.Len()

		for tracker.Heap.Len() > 0 {
			item := heap.Pop(tracker.Heap)
			entry := item.(analyzer.LogEntry)
			fmt.Printf("\n#%d [%vms] %s\n", count, entry.Latency, entry.Data)

			analysis, err := client.Analyze(entry, entry.Context)
			if err != nil {
				fmt.Printf("AI analysis failed: %v\n", err)
			} else {
				fmt.Printf("\nAI Analysis:\n%s\n", analysis)
			}

			if count < totalEntries {
				fmt.Printf("\n[Press Enter for next analysis...] ")
				scanner.Scan()
			}
			count++
		}
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the log file (required)")
	analyzeCmd.Flags().BoolVar(&noTUI, "no-tui", false, "Plain text output (default: interactive TUI)")
	analyzeCmd.Flags().IntVarP(&topK, "top", "t", 5, "Number of top outliers to analyze")
	analyzeCmd.Flags().BoolVar(&noAI, "no-ai", false, "Skip AI analysis (no API key)")

	analyzeCmd.SetUsageTemplate("\nUsage:\n  {{.UseLine}}\n\nFlags:\n{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}\n\n")

	RootCmd.AddCommand(analyzeCmd)
}
