package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	erogssync "kuromanager/internal/erogs_sync"
)

type erogsSyncDoneMsg struct {
	summary erogssync.Summary
	err     error
}

type erogsSyncProgressMsg struct {
	progress   erogssync.Progress
	progressCh <-chan erogssync.Progress
}

func runErogsSyncCmd(minID, maxID int, progressCh chan<- erogssync.Progress) tea.Cmd {
	return func() tea.Msg {
		defer close(progressCh)
		summary, err := erogssync.RunWithProgress(minID, maxID, func(progress erogssync.Progress) {
			progressCh <- progress
		})
		return erogsSyncDoneMsg{summary: summary, err: err}
	}
}

func waitErogsSyncProgressCmd(progressCh <-chan erogssync.Progress) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-progressCh
		if !ok {
			return nil
		}
		return erogsSyncProgressMsg{progress: progress, progressCh: progressCh}
	}
}

func (m *Model) beginErogsSyncInput() tea.Cmd {
	m.state = stateErogsSyncInput
	m.minIDInput.SetValue("")
	m.maxIDInput.SetValue("")
	m.maxIDInput.Blur()
	m.inputErr = ""
	return m.minIDInput.Focus()
}

func (m Model) updateErogsSyncInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.state = stateMenu
			return m, nil
		case "tab", "shift+tab":
			if m.minIDInput.Focused() {
				m.minIDInput.Blur()
				return m, m.maxIDInput.Focus()
			}
			m.maxIDInput.Blur()
			return m, m.minIDInput.Focus()
		case "enter":
			minID, maxID, err := parseErogsSyncRange(m.minIDInput.Value(), m.maxIDInput.Value())
			if err != nil {
				m.inputErr = err.Error()
				return m, nil
			}

			m.syncMinID = minID
			m.syncMaxID = maxID
			m.inputErr = ""
			m.syncConfirm.Select(0)
			m.state = stateErogsSyncConfirm
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.minIDInput.Focused() {
		m.minIDInput, cmd = m.minIDInput.Update(msg)
	} else {
		m.maxIDInput, cmd = m.maxIDInput.Update(msg)
	}
	return m, cmd
}

func parseErogsSyncRange(minText, maxText string) (int, int, error) {
	minID, err := strconv.Atoi(strings.TrimSpace(minText))
	if err != nil || minID <= 0 {
		return 0, 0, fmt.Errorf("Brand ID 下限必須是大於 0 的整數")
	}

	maxID, err := strconv.Atoi(strings.TrimSpace(maxText))
	if err != nil || maxID <= 0 {
		return 0, 0, fmt.Errorf("Brand ID 上限必須是大於 0 的整數")
	}
	if minID > maxID {
		return 0, 0, fmt.Errorf("Brand ID 下限不可大於上限")
	}
	return minID, maxID, nil
}

func (m Model) updateErogsSyncConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.state = stateErogsSyncInput
			return m, nil
		case "enter":
			item, ok := m.syncConfirm.SelectedItem().(menuEntry)
			if !ok {
				return m, nil
			}
			switch item.id {
			case "yes":
				m.state = stateErogsSyncRunning
				m.syncProgress = erogssync.Progress{}
				progressCh := make(chan erogssync.Progress)
				return m, tea.Batch(
					m.spinner.Tick,
					runErogsSyncCmd(m.syncMinID, m.syncMaxID, progressCh),
					waitErogsSyncProgressCmd(progressCh),
				)
			case "no":
				m.state = stateMenu
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.syncConfirm, cmd = m.syncConfirm.Update(msg)
	return m, cmd
}

func (m Model) updateErogsSyncRunning(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case erogsSyncProgressMsg:
		m.syncProgress = msg.progress
		return m, waitErogsSyncProgressCmd(msg.progressCh)
	case erogsSyncDoneMsg:
		m.state = stateErogsSyncResult
		if msg.err != nil {
			m.resultOK = false
			m.resultMsg = msg.err.Error()
		} else {
			m.resultOK = true
			m.resultMsg = fmt.Sprintf(
				"已同步 %d 個 brand、%d 個 game。",
				msg.summary.BrandCount,
				msg.summary.GameCount,
			)
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateErogsSyncResult(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m Model) viewErogsSyncInput() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", labelStyle.Render("輸入要同步的 Brand ID 範圍："))
	b.WriteString(m.minIDInput.View())
	b.WriteString("\n")
	b.WriteString(m.maxIDInput.View())

	if m.inputErr != "" {
		fmt.Fprintf(&b, "\n\n%s", errStyle.Render(m.inputErr))
	}
	return b.String()
}

func (m Model) viewErogsSyncConfirm() string {
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"%s\n\n",
		labelStyle.Render(fmt.Sprintf("確定要同步 Brand ID %d ～ %d 嗎？", m.syncMinID, m.syncMaxID)),
	)
	b.WriteString(m.syncConfirm.View())
	return b.String()
}

func (m Model) viewErogsSyncRunning() string {
	if m.syncProgress.Current == 0 {
		return fmt.Sprintf("%s  %s\n", m.spinner.View(), labelStyle.Render("正在取得 Erogs 資料…"))
	}

	return fmt.Sprintf(
		"%s  %s\n",
		m.spinner.View(),
		labelStyle.Render(
			fmt.Sprintf(
				"正在同步第 %d / %d 個 brand：%s（ID: %d）",
				m.syncProgress.Current,
				m.syncProgress.Total,
				m.syncProgress.BrandName,
				m.syncProgress.BrandID,
			),
		),
	)
}

func (m Model) viewErogsSyncResult() string {
	style := okStyle
	prefix := "完成"
	if !m.resultOK {
		style = errStyle
		prefix = "失敗"
	}
	return fmt.Sprintf("%s\n\n%s", style.Render(prefix), labelStyle.Render(m.resultMsg))
}
