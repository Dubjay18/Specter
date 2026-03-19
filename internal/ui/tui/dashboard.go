package tui

import (
	"fmt"
	"strings"

	"github.com/Dubjay/specter/internal/divergence"
	"github.com/Dubjay/specter/internal/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110"))

	valueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229"))

	rateNormalStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	rateWarnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("81"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

	rowIndexStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	rowMethodStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	rowPathStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	rowStatusStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	rowLatencyGoodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	rowLatencyBadStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203"))
)

type dashboardModel struct {
	snapshot     divergence.StatsSnapshot
	width        int
	height       int
	scrollOffset int
}

func newDashboard() dashboardModel {
	return dashboardModel{}
}

func (m dashboardModel) WithSnapshot(snapshot divergence.StatsSnapshot) dashboardModel {
	m.snapshot = snapshot
	m.scrollOffset = clamp(m.scrollOffset, 0, m.maxScroll())
	return m
}

func (m dashboardModel) Update(msg tea.Msg) (dashboardModel, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.scrollOffset = clamp(m.scrollOffset, 0, m.maxScroll())
	case tea.KeyMsg:
		switch typed.String() {
		case "j", "down":
			m.scrollOffset = clamp(m.scrollOffset+1, 0, m.maxScroll())
		case "k", "up":
			m.scrollOffset = clamp(m.scrollOffset-1, 0, m.maxScroll())
		case "pgdown":
			m.scrollOffset = clamp(m.scrollOffset+5, 0, m.maxScroll())
		case "pgup":
			m.scrollOffset = clamp(m.scrollOffset-5, 0, m.maxScroll())
		case "g":
			m.scrollOffset = 0
		case "G":
			m.scrollOffset = m.maxScroll()
		}
	}
	return m, nil
}

func (m dashboardModel) View() string {
	divergencePercent := m.snapshot.DivergenceRate * 100
	divergenceRateLabel := fmt.Sprintf("%.2f%%", divergencePercent)
	rateStyle := rateNormalStyle
	if m.snapshot.DivergenceRate > 0.05 {
		rateStyle = rateWarnStyle
	}

	totalCard := cardStyle.Render(
		labelStyle.Render("Total requests") + "\n" +
			valueStyle.Render(fmt.Sprintf("%d", m.snapshot.TotalRequests)),
	)

	divergenceCard := cardStyle.Render(
		labelStyle.Render("Divergences") + "\n" +
			valueStyle.Render(fmt.Sprintf("%d", m.snapshot.Divergences)) + " " + rateStyle.Render("("+divergenceRateLabel+")"),
	)

	latencyCard := cardStyle.Render(
		labelStyle.Render("Avg latency delta") + "\n" +
			valueStyle.Render(fmt.Sprintf("%.2fms", m.snapshot.AvgLatencyDeltaMs)),
	)

	metricsRow := lipgloss.JoinHorizontal(lipgloss.Top, totalCard, divergenceCard, latencyCard)

	lines := []string{
		titleStyle.Render("Specter Dashboard"),
		helpStyle.Render("Auto-refresh: 1s • Controls: j/k or ↑/↓ scroll • q quit"),
		"",
		metricsRow,
		"",
		sectionTitleStyle.Render("Recent divergence events"),
	}

	events := m.snapshot.RecentDivergences
	if len(events) == 0 {
		lines = append(lines, emptyStyle.Render("No divergence events yet"))
		return strings.Join(lines, "\n")
	}

	start, end := m.visibleRange(len(events))
	for index := start; index < end; index++ {
		lines = append(lines, formatEvent(index+1, events[index]))
	}

	return strings.Join(lines, "\n")
}

func (m dashboardModel) maxScroll() int {
	visibleRows := m.visibleRows()
	eventsCount := len(m.snapshot.RecentDivergences)
	if eventsCount <= visibleRows {
		return 0
	}
	return eventsCount - visibleRows
}

func (m dashboardModel) visibleRows() int {
	if m.height <= 0 {
		return 10
	}

	const staticLines = 16
	rows := m.height - staticLines
	if rows < 1 {
		return 1
	}
	return rows
}

func (m dashboardModel) visibleRange(total int) (int, int) {
	visibleRows := m.visibleRows()
	start := clamp(m.scrollOffset, 0, m.maxScroll())
	end := start + visibleRows
	if end > total {
		end = total
	}
	return start, end
}

func formatEvent(index int, event types.DivergenceEvent) string {
	statusDiff := "-"
	if event.StatusDiff != nil {
		statusDiff = fmt.Sprintf("%d→%d", event.StatusDiff.Live, event.StatusDiff.Shadow)
	}

	method := rowMethodStyle.Render(padRight(event.Method, 6))
	path := rowPathStyle.Render(padRight(truncate(event.RequestPath, 42), 42))
	status := rowStatusStyle.Render(padRight(statusDiff, 9))
	latencyRaw := fmt.Sprintf("%+dms", event.LatencyDiff.DeltaMs)
	latencyStyle := rowLatencyGoodStyle
	if event.LatencyDiff.DeltaMs > 0 {
		latencyStyle = rowLatencyBadStyle
	}
	latency := latencyStyle.Render(latencyRaw)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		rowIndexStyle.Render(fmt.Sprintf("%2d.", index)),
		" ",
		method,
		" ",
		path,
		" ",
		labelStyle.Render("status:"),
		status,
		" ",
		labelStyle.Render("latency:"),
		latency,
	)
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
