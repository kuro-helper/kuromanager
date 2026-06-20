package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			item, ok := m.menu.SelectedItem().(menuEntry)
			if !ok {
				return m, nil
			}
			if item.id == menuItemMigrate {
				m.state = stateMigrateConfirm
				m.confirm.Select(0)
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m Model) viewMenu() string {
	return m.menu.View()
}
