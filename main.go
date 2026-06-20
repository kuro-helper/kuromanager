package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"kuromanager/internal/tui"
)

func main() {
	if _, err := tea.NewProgram(tui.New(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}
