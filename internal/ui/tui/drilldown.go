package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Dubjay/specter/internal/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	drillTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63"))

	drillHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	okStatusStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	badStatusStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	changedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220"))

	addedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	removedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
)

type drilldownModel struct {
	event  types.DivergenceEvent
	width  int
	height int
	scroll int
}

func newDrilldown(event types.DivergenceEvent) drilldownModel {
	return drilldownModel{event: event}
}

func (m drilldownModel) Update(msg tea.Msg) (drilldownModel, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.scroll = clamp(m.scroll, 0, m.maxScroll())
	case tea.KeyMsg:
		switch typed.String() {
		case "j", "down":
			m.scroll = clamp(m.scroll+1, 0, m.maxScroll())
		case "k", "up":
			m.scroll = clamp(m.scroll-1, 0, m.maxScroll())
		case "pgdown":
			m.scroll = clamp(m.scroll+m.visibleRows(), 0, m.maxScroll())
		case "pgup":
			m.scroll = clamp(m.scroll-m.visibleRows(), 0, m.maxScroll())
		case "g":
			m.scroll = 0
		case "G":
			m.scroll = m.maxScroll()
		}
	}

	return m, nil
}

func (m drilldownModel) View() string {
	lines := []string{
		drillTitleStyle.Render("Divergence Drill-down"),
		drillHelpStyle.Render("Esc/b back • j/k or ↑/↓ scroll • q quit"),
		"",
		fmt.Sprintf("Request: %s %s", rowMethodStyle.Render(m.event.Method), rowPathStyle.Render(m.event.RequestPath)),
		m.renderStatusComparison(),
		m.renderLatencyComparison(),
		"",
		sectionTitleStyle.Render("Body diff"),
	}

	bodyRows := m.bodyDiffRows()
	if len(bodyRows) == 0 {
		lines = append(lines, emptyStyle.Render("No body field-level differences"))
		return strings.Join(lines, "\n")
	}

	start, end := m.visibleRange(len(bodyRows))
	lines = append(lines, bodyRows[start:end]...)

	return strings.Join(lines, "\n")
}

func (m drilldownModel) renderStatusComparison() string {
	if m.event.StatusDiff == nil {
		return fmt.Sprintf("Status: %s", okStatusStyle.Render("same"))
	}

	return fmt.Sprintf(
		"Status: live %s vs shadow %s %s",
		badStatusStyle.Render(fmt.Sprintf("%d", m.event.StatusDiff.Live)),
		badStatusStyle.Render(fmt.Sprintf("%d", m.event.StatusDiff.Shadow)),
		badStatusStyle.Render("(different)"),
	)
}

func (m drilldownModel) renderLatencyComparison() string {
	latencyStyle := okStatusStyle
	label := "faster/same"
	if m.event.LatencyDiff.DeltaMs > 0 {
		latencyStyle = badStatusStyle
		label = "slower"
	}

	return fmt.Sprintf(
		"Latency: live %s ms vs shadow %s ms (%s %s ms)",
		mutedStyle.Render(fmt.Sprintf("%d", m.event.LatencyDiff.LiveMs)),
		mutedStyle.Render(fmt.Sprintf("%d", m.event.LatencyDiff.ShadowMs)),
		latencyStyle.Render(label),
		latencyStyle.Render(fmt.Sprintf("%+d", m.event.LatencyDiff.DeltaMs)),
	)
}

func (m drilldownModel) bodyDiffRows() []string {
	if len(m.event.BodyDiff) == 0 {
		return nil
	}

	availableWidth := m.width
	if availableWidth <= 0 {
		availableWidth = 120
	}

	pathWidth := 28
	opWidth := 9
	valueWidth := (availableWidth - pathWidth - opWidth - 10) / 2
	if valueWidth < 20 {
		valueWidth = 20
	}

	rows := []string{
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			mutedStyle.Render(padRight("Op", opWidth)),
			" ",
			mutedStyle.Render(padRight("Field", pathWidth)),
			" ",
			mutedStyle.Render(padRight("Live", valueWidth)),
			" ",
			mutedStyle.Render(padRight("Shadow", valueWidth)),
		),
		mutedStyle.Render(strings.Repeat("-", min(availableWidth-2, opWidth+pathWidth+valueWidth*2+3))),
	}

	for _, entry := range m.event.BodyDiff {
		op := normalizeOp(entry.Op)
		opStyle := opStyle(op)

		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Top,
			opStyle.Render(padRight(strings.ToUpper(op), opWidth)),
			" ",
			truncate(entry.Path, pathWidth),
			" ",
			truncate(stringifyValue(entry.LiveValue), valueWidth),
			" ",
			truncate(stringifyValue(entry.ShadowValue), valueWidth),
		))
	}

	return rows
}

func (m drilldownModel) visibleRows() int {
	if m.height <= 0 {
		return 10
	}

	const staticLines = 12
	rows := m.height - staticLines
	if rows < 1 {
		return 1
	}
	return rows
}

func (m drilldownModel) maxScroll() int {
	rows := len(m.bodyDiffRows())
	if rows <= m.visibleRows() {
		return 0
	}
	return rows - m.visibleRows()
}

func (m drilldownModel) visibleRange(total int) (int, int) {
	start := clamp(m.scroll, 0, m.maxScroll())
	end := start + m.visibleRows()
	if end > total {
		end = total
	}
	return start, end
}

func stringifyValue(value any) string {
	if value == nil {
		return "null"
	}

	if str, ok := value.(string); ok {
		return str
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	return string(encoded)
}

func normalizeOp(op string) string {
	if op == "" {
		return "replace"
	}
	return strings.ToLower(op)
}

func opStyle(op string) lipgloss.Style {
	switch op {
	case "add":
		return addedStyle
	case "remove":
		return removedStyle
	default:
		return changedStyle
	}
}
