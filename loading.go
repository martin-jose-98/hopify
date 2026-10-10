package main

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// fetchMsg carries the result of an async fetch back to the tea loop.
type fetchMsg struct {
	data any
	err  error
}

type loadingModel struct {
	spinner   spinner.Model
	title     string
	cancelled bool
	result    fetchMsg
	fetch     func() tea.Msg
	width     int
	height    int
}

// Init starts the spinner ticker and launches the fetch in the background.
func (m loadingModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetch)
}

func (m loadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true // background fetch still finishes; its Msg is dropped
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case fetchMsg:
		m.result = msg
		return m, tea.Quit
	}
	return m, nil
}

func (m loadingModel) View() tea.View {
	line := m.spinner.View() + " " + m.title
	if m.width > 0 && m.height > 0 {
		return tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, line))
	}
	return tea.NewView(line)
}

// runLoading shows a spinner while fn runs, then returns its result.
// A generic wrapper, so any fetch signature works. Cancels with errCancelled.
func runLoading[T any](title string, fn func() (T, error)) (T, error) {
	var zero T

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
	)

	m := loadingModel{
		spinner: s,
		title:   title,
		fetch: func() tea.Msg {
			data, err := fn()
			return fetchMsg{data: data, err: err}
		},
	}

	model, err := tea.NewProgram(m).Run()
	if err != nil {
		return zero, err
	}
	lm := model.(loadingModel)
	switch {
	case lm.cancelled:
		return zero, errCancelled
	case lm.result.err != nil:
		return zero, lm.result.err
	}
	return lm.result.data.(T), nil
}
