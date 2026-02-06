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

type RingBuffer struct {
	data  []string
	size  int
	index int
	full  bool
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]string, size),
		size: size,
	}
}

func (rb *RingBuffer) Add(line string) {
	rb.data[rb.index] = line
	rb.index++
	if rb.index >= rb.size {
		rb.index = 0
		rb.full = true
	}
}

func (rb *RingBuffer) GetAll() []string {
	if !rb.full {
		return rb.data[:rb.index]
	}
	result := make([]string, rb.size)
	copy(result, rb.data[rb.index:])
	copy(result[rb.size-rb.index:], rb.data[:rb.index])
	return result
}

func ProcessFile(path string, tracker *analyzer.P99Tracker) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	ring := NewRingBuffer(5)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var l LogLine

		if err := json.Unmarshal([]byte(line), &l); err != nil {
			ring.Add(line)
			continue
		}

		tracker.Process(analyzer.LogEntry{
			Latency: l.Latency,
			Data:    line,
			Context: ring.GetAll(),
		})

		ring.Add(line)
	}
	return scanner.Err()
}
