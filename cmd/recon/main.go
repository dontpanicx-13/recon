package main

import (
	"log"

	"os"

	tea "github.com/charmbracelet/bubbletea"

	"recon/internal/ui"
)

func main() {
	program := tea.NewProgram(ui.InitalModel(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
