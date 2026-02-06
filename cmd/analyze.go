package cmd

import (
	"container/heap"
	"fmt"

	"github.com/mohammadalshikh/tail-cli/internal/analyzer"
	"github.com/mohammadalshikh/tail-cli/internal/ingestor"
	"github.com/spf13/cobra"
)

var filePath string

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze logs to find P99 latencies",
	Run: func(cmd *cobra.Command, args []string) {
		if filePath == "" {
			fmt.Println("Error: please provide a file path using --file")
			return
		}

		tracker := analyzer.NewTracker(5)

		fmt.Printf("Analyzing %s...\n", filePath)
		err := ingestor.ProcessFile(filePath, tracker)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		fmt.Println("\n--- Top P99 outliers found ---")

		for tracker.Heap.Len() > 0 {
			item := heap.Pop(tracker.Heap)
			entry := item.(analyzer.LogEntry)
			fmt.Printf("[%vms] %s\n", entry.Latency, entry.Data)
		}
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the log file")

	RootCmd.AddCommand(analyzeCmd)
}
