package tui

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohammadalshikh/tail-cli/internal/agent"
	"github.com/mohammadalshikh/tail-cli/internal/analyzer"
)

type state int

const (
	stateLoading state = iota
	stateViewing
	stateAnalyzing
	stateError
)

type model struct {
	entries      []analyzer.LogEntry
	currentIndex int
	analysis     map[int]string
	state        state
	spinner      spinner.Model
	client       *agent.OpenAIClient
	noAI         bool
	err          error
	width        int
	height       int
}

type analysisMsg struct {
	index    int
	analysis string
	err      error
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Padding(0, 1)

	latencyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Padding(0, 1)

	dataStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Padding(0, 1)

	aiStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0)
)

func NewModel(tracker *analyzer.P99Tracker, client *agent.OpenAIClient, noAI bool) model {
	entries := make([]analyzer.LogEntry, 0, tracker.Heap.Len())
	for tracker.Heap.Len() > 0 {
		item := heap.Pop(tracker.Heap)
		entries = append(entries, item.(analyzer.LogEntry))
	}

	for i := 0; i < len(entries)/2; i++ {
		j := len(entries) - i - 1
		entries[i], entries[j] = entries[j], entries[i]
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		entries:      entries,
		currentIndex: 0,
		analysis:     make(map[int]string),
		state:        stateViewing,
		spinner:      s,
		client:       client,
		noAI:         noAI,
		width:        80,
		height:       24,
	}
}

func (m model) Init() tea.Cmd {
	if !m.noAI && m.client != nil {
		return tea.Batch(
			m.spinner.Tick,
			m.analyzeCurrentEntry(),
		)
	}
	return nil
}

func (m model) analyzeCurrentEntry() tea.Cmd {
	if m.noAI || m.client == nil {
		return nil
	}

	idx := m.currentIndex
	entry := m.entries[idx]

	if _, exists := m.analysis[idx]; exists {
		return nil
	}

	return func() tea.Msg {
		analysis, err := m.client.Analyze(entry, entry.Context)
		return analysisMsg{
			index:    idx,
			analysis: analysis,
			err:      err,
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "left", "a":
			if m.currentIndex > 0 {
				m.currentIndex--
				if !m.noAI && m.client != nil {
					return m, m.analyzeCurrentEntry()
				}
			}

		case "right", "d":
			if m.currentIndex < len(m.entries)-1 {
				m.currentIndex++
				if !m.noAI && m.client != nil {
					return m, m.analyzeCurrentEntry()
				}
			}
		}

	case analysisMsg:
		if msg.err != nil {
			m.analysis[msg.index] = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.analysis[msg.index] = msg.analysis
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func wrapText(text string, width int) string {
	if width <= 0 {
		width = 80
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > width {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

func (m model) View() string {
	if len(m.entries) == 0 {
		return helpStyle.Render("No outliers found.\n\nPress q to quit")
	}

	maxWidth := m.width
	if maxWidth <= 0 {
		maxWidth = 80
	}

	entry := m.entries[m.currentIndex]
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("P99 Latency Outliers [%d/%d]", m.currentIndex+1, len(m.entries))))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render(fmt.Sprintf("\nEntry #%d\n", m.currentIndex+1)))
	b.WriteString("\n")
	b.WriteString(latencyStyle.Render(fmt.Sprintf("Latency: %.0fms", entry.Latency)))
	b.WriteString("\n")

	wrappedData := wrapText(entry.Data, maxWidth-10)
	b.WriteString(dataStyle.Render(fmt.Sprintf("%s", wrappedData)))
	b.WriteString("\n\n")

	if !m.noAI && m.client != nil {
		if analysis, exists := m.analysis[m.currentIndex]; exists {
			b.WriteString(headerStyle.Render("AI Analysis"))
			b.WriteString("\n")
			wrappedAnalysis := wrapText(analysis, maxWidth-6)
			b.WriteString(aiStyle.Render(wrappedAnalysis))
			b.WriteString("\n")
		} else {
			b.WriteString(headerStyle.Render("AI Analysis"))
			b.WriteString("\n")
			b.WriteString(aiStyle.Render(fmt.Sprintf("%s Analyzing...", m.spinner.View())))
			b.WriteString("\n")
		}
	}

	helpText := "←/a: prev  →/d: next  q: quit"
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

func Run(tracker *analyzer.P99Tracker, client *agent.OpenAIClient, noAI bool) error {
	m := NewModel(tracker, client, noAI)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
