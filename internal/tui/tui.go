// 介面分成多個畫面，各自實作於獨立檔案：
//   - menu.go    ：主選單（stateMenu）
//   - migrate.go ：migrate 確認與執行（其餘狀態）
//
// 本檔案存放共用的 model、樣式，以及將工作分派給對應畫面的
// 最上層 Update／View 分派器。
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateMenu state = iota
	stateMigrateConfirm
	stateMigrateRunning
	stateMigrateResult
)

const (
	menuItemMigrate = "migrate"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	faintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// 供 bubbles/list 使用的選單項目。
type menuEntry struct {
	id, title, desc string
}

func (e menuEntry) Title() string       { return e.title }
func (e menuEntry) Description() string { return e.desc }
func (e menuEntry) FilterValue() string { return e.id }

// 驅動管理介面的 Bubble Tea model。
type Model struct {
	state   state
	menu    list.Model
	confirm list.Model
	spinner spinner.Model

	resultMsg string
	resultOK  bool
}

// 建立一個可直接交給 Bubble Tea 程式執行的 Model。
func New() Model {
	menu := newMenuList(mainMenuItems(), "主選單")
	confirm := newMenuList(confirmMenuItems(), "確認")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		state:   stateMenu,
		menu:    menu,
		confirm: confirm,
		spinner: sp,
	}
}

func newMenuList(items []list.Item, title string) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("252"))

	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.DisableQuitKeybindings()
	return l
}

func mainMenuItems() []list.Item {
	return []list.Item{
		menuEntry{
			id:    menuItemMigrate,
			title: "migrate",
			desc:  "執行資料庫 schema migration（kurohelper-service）",
		},
	}
}

func confirmMenuItems() []list.Item {
	return []list.Item{
		menuEntry{id: "yes", title: "確定", desc: "執行 migration"},
		menuEntry{id: "no", title: "取消", desc: "返回主選單"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// 先處理全域訊息，再依目前畫面分派後續處理。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.menu.SetSize(msg.Width, msg.Height-4)
		m.confirm.SetSize(msg.Width, msg.Height-4)
		return m, nil
	}

	switch m.state {
	case stateMenu:
		return m.updateMenu(msg)
	case stateMigrateConfirm:
		return m.updateMigrateConfirm(msg)
	case stateMigrateRunning:
		return m.updateMigrateRunning(msg)
	case stateMigrateResult:
		return m.updateMigrateResult(msg)
	default:
		return m, nil
	}
}

// 繪製標題、目前畫面內容，以及操作提示列。
func (m Model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render("Kuro Manager"))

	switch m.state {
	case stateMenu:
		b.WriteString(m.viewMenu())
	case stateMigrateConfirm:
		b.WriteString(m.viewMigrateConfirm())
	case stateMigrateRunning:
		b.WriteString(m.viewMigrateRunning())
	case stateMigrateResult:
		b.WriteString(m.viewMigrateResult())
	}

	fmt.Fprintf(&b, "\n%s", helpStyle.Render(m.helpText()))
	return b.String()
}

func (m Model) helpText() string {
	switch m.state {
	case stateMenu:
		return "↑↓：選擇 • enter：確認 • q：離開 • ctrl+c：離開"
	case stateMigrateConfirm:
		return "↑↓：選擇 • enter：確認 • esc：返回主選單 • ctrl+c：離開"
	case stateMigrateRunning:
		return "執行中…"
	case stateMigrateResult:
		return "enter / esc / q：返回主選單 • ctrl+c：離開"
	default:
		return "ctrl+c：離開"
	}
}
