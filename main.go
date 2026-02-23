package main

import (
	"fmt"
	"os"

	"github.com/10yearsahead/easy-git/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	app := ui.NewApp()

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),       // Use alternate screen (no scroll history clutter)
		tea.WithMouseCellMotion(), // Optional: mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting easy-git: %v\n", err)
		os.Exit(1)
	}
}
