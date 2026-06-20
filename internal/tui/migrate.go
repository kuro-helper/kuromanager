package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"kuromanager/internal/migrate"
)

type migrateDoneMsg struct {
	err error
}

func runMigrateCmd() tea.Cmd {
	return func() tea.Msg {
		return migrateDoneMsg{err: migrate.Run()}
	}
}

func (m Model) updateMigrateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.state = stateMenu
			return m, nil
		case "enter":
			item, ok := m.confirm.SelectedItem().(menuEntry)
			if !ok {
				return m, nil
			}
			switch item.id {
			case "yes":
				m.state = stateMigrateRunning
				return m, tea.Batch(m.spinner.Tick, runMigrateCmd())
			case "no":
				m.state = stateMenu
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.confirm, cmd = m.confirm.Update(msg)
	return m, cmd
}

func (m Model) updateMigrateRunning(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case migrateDoneMsg:
		m.state = stateMigrateResult
		if msg.err != nil {
			m.resultOK = false
			m.resultMsg = msg.err.Error()
		} else {
			m.resultOK = true
			m.resultMsg = "資料庫 migration 已完成。"
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateMigrateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter", "esc", "q":
			m.state = stateMenu
			m.resultMsg = ""
			return m, nil
		}
	}
	return m, nil
}

func (m Model) viewMigrateConfirm() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", labelStyle.Render("確定要執行資料庫 migration 嗎？"))
	b.WriteString(m.confirm.View())
	return b.String()
}

func (m Model) viewMigrateRunning() string {
	return fmt.Sprintf("%s  %s\n", m.spinner.View(), labelStyle.Render("正在執行 migration…"))
}

func (m Model) viewMigrateResult() string {
	style := okStyle
	prefix := "完成"
	if !m.resultOK {
		style = errStyle
		prefix = "失敗"
	}
	return fmt.Sprintf("%s\n\n%s",
		style.Render(prefix),
		labelStyle.Render(m.resultMsg),
	)
}
