package main

import (
	"log"

	"os"

	tea "github.com/charmbracelet/bubbletea"

	"recon/internal/ui"
)

var (
	version = "dev"
	name    = "recon"
)

func main() {
	program := tea.NewProgram(ui.InitalModel(name, version), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
