package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/crabstars/liftoff/hetzner"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/joho/godotenv"
)

const (
	SERVER_CREATED_SUCCESS = "server-created-success"
	SERVER_CREATED_Failed  = "server-created-failed"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type CreateServerState struct {
	waitingForServerNameInput bool
	serverNameInput           textinput.Model
	creatingServer            bool
}

type TableState struct {
	showTable             bool
	serverTable           table.Model
	tabelUpdateChannel    chan TableUpdateMsg
	tabelReloadingChannel chan bool
	tableReloadRunning    bool
	rowCursor             int
}

type ActionSelectionState struct {
	choices []string // create or delete server
	cursor  int      // which list item our cursor is pointing at
}

type TableUpdateMsg []table.Row

type model struct {
	spinner              spinner.Model
	createServerState    CreateServerState
	actionSelectionState ActionSelectionState
	tableState           TableState
	program              *tea.Program
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	ti := textinput.New()
	ti.Placeholder = "Server Name"
	ti.CharLimit = 156
	ti.Width = 20
	return model{
		createServerState:    CreateServerState{serverNameInput: ti},
		actionSelectionState: ActionSelectionState{choices: []string{"Show server", "Create server", "Delete server"}},
		tableState:           TableState{tabelUpdateChannel: make(chan TableUpdateMsg), tabelReloadingChannel: make(chan bool)},
		spinner:              s,
	}
}

type TickMsg time.Time

func (m model) Init() tea.Cmd {
	return tea.Cmd(tickEvery(time.Second * 2))
}

func tickEvery(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case TickMsg:

		select {
		case m.tableState.tableReloadRunning = <-m.tableState.tabelReloadingChannel:
		default:
		}

		if m.tableState.showTable {
			// go m.fetchTableRows()
			if !m.tableState.tableReloadRunning {
				m.tableState.tableReloadRunning = true
				go m.fetchTableRows()
			}
			select {
			case rows := <-m.tableState.tabelUpdateChannel:
				m.loadTableWithoutFetch(rows)
			default:
			}
		}
		return m, tickEvery(time.Second * 2)

	case tea.KeyMsg:
		keyStroke := msg.String()

		if keyStroke == "q" || keyStroke == "ctrl+c" {
			return m, tea.Quit
		}

		if m.createServerState.creatingServer {
			return m, nil
		}

		if m.createServerState.waitingForServerNameInput {
			if !m.createServerState.serverNameInput.Focused() {
				m.createServerState.serverNameInput.Focus()
			}
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.createServerState.waitingForServerNameInput = false
				return m, nil
			case tea.KeyEnter:
				m.createServerState.waitingForServerNameInput = false
				m.createServerState.creatingServer = true
				return m, tea.Batch(m.spinner.Tick, createServer(m.createServerState.serverNameInput.Value(), &m))
			}
			m.createServerState.serverNameInput, cmd = m.createServerState.serverNameInput.Update(msg)

			return m, cmd

		}

		if m.tableState.showTable {
			switch keyStroke {
			case "esc":
				m.tableState.showTable = false
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				return m, tea.Batch(
					tea.Printf("Let's go to %s!", m.tableState.serverTable.SelectedRow()[1]),
				)
			}
			m.tableState.serverTable, cmd = m.tableState.serverTable.Update(msg)
			m.tableState.rowCursor = m.tableState.serverTable.Cursor()
			return m, cmd
		}

		switch keyStroke {
		case "up", "k":
			if m.actionSelectionState.cursor > 0 {
				m.actionSelectionState.cursor--
			}

		case "down", "j":
			if m.actionSelectionState.cursor < len(m.actionSelectionState.choices)-1 {
				m.actionSelectionState.cursor++
			}

		case "enter", " ":
			switch m.actionSelectionState.cursor {
			case 0:
				log.Printf("Showing server")
				m.loadTable()
				// m.fetchTableRows()
				// rows := <-m.tableState.tabelUpdateChannel
				// m.loadTableWithoutFetch(rows)
				m.tableState.showTable = true
			case 1:
				m.createServerState.waitingForServerNameInput = true
				log.Printf("waiting for name input")
				return m, nil
			case 2:
				log.Printf("Delete server")
			default:

				log.Printf("Choice not found")
			}

		}
	case string:
		if msg == SERVER_CREATED_SUCCESS {
			m.createServerState.creatingServer = false
			log.Printf("Server created successfully")
		} else if msg == SERVER_CREATED_Failed {
			m.createServerState.creatingServer = false
			log.Printf("Server creation failed")
		}

	case spinner.TickMsg:
		if m.createServerState.creatingServer {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) ViewState() string {
	var builder strings.Builder

	builder.WriteString("Model:\n")

	builder.WriteString("  Spinner:\n")
	builder.WriteString(fmt.Sprintf("    Current Frame: %s\n", m.spinner.View()))

	builder.WriteString("  CreateServerState:\n")
	builder.WriteString(fmt.Sprintf("    WaitingForServerNameInput: %v\n", m.createServerState.waitingForServerNameInput))
	builder.WriteString(fmt.Sprintf("    ServerNameInput: %s\n", m.createServerState.serverNameInput.Value()))
	builder.WriteString(fmt.Sprintf("    CreatingServer: %v\n", m.createServerState.creatingServer))

	builder.WriteString("  ActionSelectionState:\n")
	builder.WriteString(fmt.Sprintf("    Choices: %s\n", strings.Join(m.actionSelectionState.choices, ", ")))
	builder.WriteString(fmt.Sprintf("    Cursor: %d\n", m.actionSelectionState.cursor))

	builder.WriteString("  TableState:\n")
	builder.WriteString(fmt.Sprintf("    TableReloadRunning: %v\n", m.tableState.tableReloadRunning))
	builder.WriteString(fmt.Sprintf("    ShowTable: %v\n", m.tableState.showTable))
	builder.WriteString(fmt.Sprintf("    RowCursor: %d\n", m.tableState.rowCursor))

	builder.WriteString(fmt.Sprintf("\n\n", m.tableState.rowCursor))
	return builder.String()

}

func (m model) ViewHandleCreateServerState() string {

	if m.createServerState.waitingForServerNameInput {
		return fmt.Sprintf("Enter server name:\n\n%s\n\n%s", m.createServerState.serverNameInput.View(), "(esc to quit)")
	}

	if m.createServerState.creatingServer {
		return fmt.Sprintf("\n\n   %s Loading server creation...press q to quit LiftOff\n\n", m.spinner.View())
	}

	return ""
}

func (m model) View() string {
	s := m.ViewState()

	if state := m.ViewHandleCreateServerState(); state != "" {
		s += state
		return s
	}
	if m.tableState.showTable {
		log.Printf("%s", m.tableState.serverTable.View()+" "+m.tableState.serverTable.HelpView()+"\n")
		s += baseStyle.Render(m.tableState.serverTable.View()) + "\n " + m.tableState.serverTable.HelpView() + "\n"
		return s

	}

	s += "Choose hetzner action\n\n"

	for i, choice := range m.actionSelectionState.choices {
		cursor := " "
		if m.actionSelectionState.cursor == i {
			cursor = ">"
		}

		checked := " "
		if i == m.actionSelectionState.cursor {
			checked = "x"
		}
		s += fmt.Sprintf("%s [%s], %s\n", cursor, checked, choice)
	}
	s += "\nPress q to quit.\n"
	return s
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	model := initialModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	model.program = p

	if _, err := p.Run(); err != nil {
		log.Fatal("Error while starting %v", err)
	}
}

type CreateServer struct {
	ServerName    string `json:"serverName"`
	DeployCountry string `json:"deployCountry"`
	GithubLink    string `json:"githubLink"`
	SshKeyName    string `json:"sshKeyName"`
}

func (m *model) loadTable() {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Image", Width: 20},
		{Title: "Status", Width: 10},
		{Title: "Datacenter", Width: 10},
		{Title: "CPU Type", Width: 8},
		{Title: "Server Type", Width: 15},
		{Title: "Cores", Width: 10},
		{Title: "Memory", Width: 10},
		{Title: "Disk", Width: 10},
	}
	rows := listServer()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	m.tableState.serverTable = t
}

func (m *model) loadTableWithoutFetch(rows []table.Row) {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Image", Width: 20},
		{Title: "Status", Width: 10},
		{Title: "Datacenter", Width: 10},
		{Title: "CPU Type", Width: 8},
		{Title: "Server Type", Width: 15},
		{Title: "Cores", Width: 10},
		{Title: "Memory", Width: 10},
		{Title: "Disk", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	if len(rows)-1 < m.tableState.rowCursor {
		t.SetCursor(len(rows) - 1)
	} else {
		t.SetCursor(m.tableState.rowCursor)
	}

	m.tableState.serverTable = t
}
func (m *model) fetchTableRows() {
	var rows []table.Row
	hetzner_key := os.Getenv("HETZNER_CLOUD_API_KEY")
	client := hcloud.NewClient(hcloud.WithToken(hetzner_key))
	servers, err := client.Server.All(context.Background())
	if err != nil {
		log.Println("could not get fetch server", err)
		return
	}
	for _, server := range servers {
		rows = append(rows, table.Row{server.Name, server.Image.Name, string(server.Status), server.Datacenter.Location.City,
			string(server.ServerType.CPUType), server.ServerType.Name, fmt.Sprintf("%d", server.ServerType.Cores), fmt.Sprintf("%.0f GB", server.ServerType.Memory), fmt.Sprintf("%d GB", server.ServerType.Disk)})
	}
	log.Println("table fetched")
	m.tableState.tableReloadingChannel <- false
	m.tableState.tabelUpdateChannel <- rows
}
func listServer() []table.Row {

	var rows []table.Row
	hetzner_key := os.Getenv("HETZNER_CLOUD_API_KEY")
	client := hcloud.NewClient(hcloud.WithToken(hetzner_key))
	servers, err := client.Server.All(context.Background())
	if err != nil {
		log.Println("could not get all server", err)
		return rows
	}
	for _, server := range servers {
		rows = append(rows, table.Row{server.Name, server.Image.Name, string(server.Status), server.Datacenter.Location.City,
			string(server.ServerType.CPUType), server.ServerType.Name, fmt.Sprintf("%d", server.ServerType.Cores), fmt.Sprintf("%.0f GB", server.ServerType.Memory), fmt.Sprintf("%d GB", server.ServerType.Disk)})
	}
	return rows
}

func createServer(serverName string, m *model) tea.Cmd {
	m.createServerState.serverNameInput.Reset()
	return func() tea.Msg {
		return createHetznerServer(serverName, m)
	}
}

func createHetznerServer(serverName string, m *model) tea.Msg {
	hetzner_key := os.Getenv("HETZNER_CLOUD_API_KEY")
	sshKeyName := os.Getenv("SSH_KEY_NAME")
	client := hcloud.NewClient(hcloud.WithToken(hetzner_key))

	serverOption := CreateServer{ServerName: serverName, DeployCountry: "germany", GithubLink: "", SshKeyName: sshKeyName}
	serverType, err := hetzner.GetSmallestServer(client)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	image, err := hetzner.GetDockerCeImage(client)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}
	datacenter, err := hetzner.GetDatacenter(client, serverOption.DeployCountry)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	sshKey, err := hetzner.GetSshKey(client, serverOption.SshKeyName)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	automount := false
	serverCreateResult, _, err := client.Server.Create(
		context.Background(),
		hcloud.ServerCreateOpts{
			Name:       serverOption.ServerName,
			Automount:  &automount, // volumes for mounting
			Datacenter: datacenter,
			Image:      image,
			ServerType: serverType,
			SSHKeys: []*hcloud.SSHKey{
				sshKey,
			},
			// UserData: cloudConfig,
		},
	)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	go checkAction(client, serverCreateResult.Action.ID)
	m.createServerState.creatingServer = false
	return SERVER_CREATED_SUCCESS
	//
	// RestartServer(client, serverCreateResult.Server.ID)
	// err = sshconnector.RunCommandsOnServer(serverCreateResult.Server.PublicNet.IPv4.IP.String(), []sshconnector.Command{})
	// if err != nil {
	// 	return
	// }
}

func deleteServer(serverID int) {
	hetzner_key := os.Getenv("HETZNER_CLOUD_API_KEY")
	client := hcloud.NewClient(hcloud.WithToken(hetzner_key))
	server, _, err := client.Server.GetByID(context.Background(), int64(serverID))
	if err != nil {
		log.Println("could not get server for deleting", err)
		return
	}
	_, _, err = client.Server.DeleteWithResult(context.Background(), server)
	if err != nil {
		log.Println("could not delete server", err)
		return
	}
}

// func RestartServer(client *hcloud.Client, serverID int64) {
// 	for {
// 		server, _, err := client.Server.GetByID(context.Background(), serverID)
// 		if err != nil {
// 			fmt.Println("error while reading serverinformation", err)
// 		}
// 		fmt.Println(serverID, ":", server.Status)
// 		if server.Status == hcloud.ServerStatusRunning {
// 			_, _, err := client.Server.Reboot(context.Background(), server)
// 			if err != nil {
// 				fmt.Println("error while reboot server", err)
// 			}
// 			return
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// }

func checkAction(client *hcloud.Client, actionID int64) {
	for {
		action, _, err := client.Action.GetByID(context.Background(), actionID)
		if err != nil {
			log.Println("Checking action failed", actionID)
			return
		}
		if action.Status == hcloud.ActionStatusSuccess || action.Status == hcloud.ActionStatusError {
			log.Println("Action success")
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func createServerSimulation(zahl int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Duration(zahl) * time.Second)
		return SERVER_CREATED_SUCCESS
	}
}
