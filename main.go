package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crabstars/liftoff/model"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	hetznerKey := os.Getenv("HETZNER_CLOUD_API_KEY")
	sshKeyName := os.Getenv("SSH_KEY_NAME")
	if hetznerKey == "" || sshKeyName == "" {
		panic("Add HETZNER_CLOUD_API_KEY and SSH_KEY_NAME to .env file, see .env.example")
	}

}

func main() {

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	model := model.InitialModel()
	p := tea.NewProgram(&model, tea.WithAltScreen())
	model.Program = p

	if _, err := p.Run(); err != nil {
		log.Fatal("Error while starting %v", err)
	}
}
