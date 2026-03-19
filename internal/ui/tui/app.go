package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Dubjay/specter/internal/divergence"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var errorBannerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("15")).
	Background(lipgloss.Color("160")).
	Padding(0, 1)

const pollInterval = time.Second

type statsFetchedMsg struct {
	snapshot divergence.StatsSnapshot
	err      error
}

type tickMsg time.Time

type appModel struct {
	client    *http.Client
	statsURL  string
	dashboard dashboardModel
	lastError error
}

func Run(statsURL string) error {
	program := tea.NewProgram(newAppModel(statsURL), tea.WithAltScreen())
	_, err := program.Run()
	return err
}

func newAppModel(statsURL string) appModel {
	return appModel{
		client:    &http.Client{Timeout: 2 * time.Second},
		statsURL:  statsURL,
		dashboard: newDashboard(),
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(
		fetchStatsCmd(m.client, m.statsURL),
		tickCmd(),
	)
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tickMsg:
		return m, tea.Batch(
			fetchStatsCmd(m.client, m.statsURL),
			tickCmd(),
		)
	case statsFetchedMsg:
		if typed.err != nil {
			m.lastError = typed.err
			return m, nil
		}
		m.lastError = nil
		m.dashboard = m.dashboard.WithSnapshot(typed.snapshot)
		return m, nil
	}

	var cmd tea.Cmd
	m.dashboard, cmd = m.dashboard.Update(msg)
	return m, cmd
}

func (m appModel) View() string {
	if m.lastError != nil {
		banner := errorBannerStyle.Render(fmt.Sprintf("Stats fetch failed: %v", m.lastError))
		return fmt.Sprintf("%s\n\n%s", banner, m.dashboard.View())
	}
	return m.dashboard.View()
}

func tickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchStatsCmd(client *http.Client, statsURL string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 1800*time.Millisecond)
		defer cancel()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, statsURL, nil)
		if err != nil {
			return statsFetchedMsg{err: err}
		}

		response, err := client.Do(request)
		if err != nil {
			return statsFetchedMsg{err: err}
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return statsFetchedMsg{err: fmt.Errorf("unexpected status code: %d", response.StatusCode)}
		}

		var snapshot divergence.StatsSnapshot
		if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
			return statsFetchedMsg{err: err}
		}

		return statsFetchedMsg{snapshot: snapshot}
	}
}
