package ingestor

import (
	"bufio"
	"encoding/json"
	"os"
	"github.com/mohammadalshikh/tail-cli/internal/analyzer"
)

type LogLine struct {
	Latency float64 `json:"latency_ms"`
	Msg     string  `json:"msg"`
}

func ProcessFile(path string, tracker *analyzer.P99Tracker) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var l LogLine

		if err := json.Unmarshal([]byte(scanner.Text()), &l); err != nil {
			continue
		}

		tracker.Process(analyzer.LogEntry{
			Latency: l.Latency,
			Data:    scanner.Text(),
		})
	}
	return scanner.Err()
}