package model

import (
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
)

type CreateServerState struct {
	WaitingForServerNameInput bool
	ServerNameInput           textinput.Model
	CreatingServer            bool
}

type TableState struct {
	ShowTable             bool
	ServerTable           table.Model
	TabelReloadingChannel chan bool
	TableReloadRunning    bool
	RowCursor             int
	// index corresponds to the row index
	ServerIdIndexRelations []int64
	ShowOverlay            bool
}

type ActionSelectionState struct {
	Choices []string // create or delete server
	Cursor  int      // which list item our cursor is pointing at
}

type TableUpdateMsg struct {
	rows []table.Row
	ids  []int64
}

type TickMsg time.Time

type EnvVariables struct {
	HetznerApiKey string
	SshKeyName    string
	Debug         bool
}
type Model struct {
	Spinner              spinner.Model
	CreateServerState    CreateServerState
	ActionSelectionState ActionSelectionState
	TableState           TableState
	Program              *tea.Program
	EnvValues            EnvVariables
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	ti := textinput.New()
	ti.Placeholder = "Server Name"
	ti.CharLimit = 156
	ti.Width = 20
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return Model{
		CreateServerState:    CreateServerState{ServerNameInput: ti},
		ActionSelectionState: ActionSelectionState{Choices: []string{"Show server", "Create server"}},
		TableState:           TableState{TabelReloadingChannel: make(chan bool)},
		Spinner:              s,
		EnvValues:            EnvVariables{HetznerApiKey: os.Getenv("HETZNER_CLOUD_API_KEY"), SshKeyName: os.Getenv("SSH_KEY_NAME"), Debug: (len(os.Getenv("DEBUG")) > 0)},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Cmd(tickEvery(time.Second * 2))
}

func tickEvery(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
