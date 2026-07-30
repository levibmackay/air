package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/router"
)

const maxLogLines = 8

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	valueStyle   = lipgloss.NewStyle().Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	logStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).MarginLeft(2)
	boxStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type tickMsg time.Time

// resultMsg carries a completed Router.Run's outcome into the TUI.
type resultMsg struct {
	checkpoint *checkpoint.Checkpoint
	err        error
}

type model struct {
	sessionID string
	objective string

	provider    string
	startedAt   time.Time
	checkpoints int
	lastError   string
	statusLog   []string
	outputTail  []string

	done   bool
	result *checkpoint.Checkpoint
	runErr error

	events     <-chan router.Event
	cancel     context.CancelFunc
	spinnerIdx int
	quitting   bool
}

func initialModel(sessionID, objective string, events <-chan router.Event, cancel context.CancelFunc) model {
	return model{
		sessionID: sessionID,
		objective: objective,
		startedAt: time.Now(),
		events:    events,
		cancel:    cancel,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(waitForEvent(m.events), tick())
}

func tick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func waitForEvent(events <-chan router.Event) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-events
		if !ok {
			return nil
		}
		return e
	}
}

func (m *model) appendStatus(line string) {
	m.statusLog = append(m.statusLog, line)
	if len(m.statusLog) > maxLogLines {
		m.statusLog = m.statusLog[len(m.statusLog)-maxLogLines:]
	}
}

func (m *model) setOutputTail(output string) {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) > maxLogLines {
		lines = lines[len(lines)-maxLogLines:]
	}
	m.outputTail = lines
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}

	case tickMsg:
		if m.done {
			return m, nil
		}
		m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
		return m, tick()

	case router.Event:
		switch msg.Type {
		case router.EventProviderStarted:
			m.provider = msg.Provider
			m.appendStatus(fmt.Sprintf("started %s", msg.Provider))
		case router.EventCheckpoint:
			m.checkpoints++
			if msg.Checkpoint != nil {
				m.setOutputTail(msg.Checkpoint.TerminalOutput)
			}
		case router.EventProviderFailed:
			if msg.Checkpoint != nil && len(msg.Checkpoint.Errors) > 0 {
				m.lastError = msg.Checkpoint.Errors[len(msg.Checkpoint.Errors)-1]
			}
			m.appendStatus(fmt.Sprintf("%s failed: %s", msg.Provider, m.lastError))
		case router.EventCompleted:
			m.appendStatus(fmt.Sprintf("%s completed", msg.Provider))
		}
		return m, waitForEvent(m.events)

	case resultMsg:
		m.done = true
		m.result = msg.checkpoint
		m.runErr = msg.err
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("AIR — AI Router") + "\n")
	b.WriteString(labelStyle.Render("session:   ") + valueStyle.Render(m.sessionID) + "\n")
	b.WriteString(labelStyle.Render("objective: ") + valueStyle.Render(m.objective) + "\n\n")

	status := m.provider
	if status == "" {
		status = "starting..."
	}
	switch {
	case !m.done:
		status = spinnerFrames[m.spinnerIdx] + " " + status
	case m.runErr != nil:
		status = errorStyle.Render("failed: " + m.runErr.Error())
	default:
		status = successStyle.Render("completed via " + status)
	}
	b.WriteString(labelStyle.Render("provider:    ") + status + "\n")
	b.WriteString(labelStyle.Render("elapsed:     ") + valueStyle.Render(time.Since(m.startedAt).Round(time.Second).String()) + "\n")
	b.WriteString(labelStyle.Render("checkpoints: ") + valueStyle.Render(fmt.Sprintf("%d", m.checkpoints)) + "\n")
	if m.lastError != "" {
		b.WriteString(errorStyle.Render("last issue:  "+m.lastError) + "\n")
	}

	if len(m.statusLog) > 0 {
		b.WriteString("\n" + labelStyle.Render("recent status:") + "\n")
		for _, l := range m.statusLog {
			b.WriteString(logStyle.Render(l) + "\n")
		}
	}
	if len(m.outputTail) > 0 {
		b.WriteString("\n" + labelStyle.Render("latest output:") + "\n")
		for _, l := range m.outputTail {
			b.WriteString(logStyle.Render(l) + "\n")
		}
	}

	b.WriteString("\n" + labelStyle.Render("q to quit"))

	return boxStyle.Render(b.String())
}
