package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
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

// TODO cache later some hetzner information => only get server information at the start
type model struct {
	choices     []string // create or delete server
	cursor      int      // which list item our cursor is pointing at
	spinner     spinner.Model
	loading     bool
	showTable   bool
	serverTable table.Model
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return model{
		choices: []string{"Show server", "Create server", "Delete server"},
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		keyStroke := msg.String()

		if keyStroke == "q" || keyStroke == "ctrl+c" {
			return m, tea.Quit
		}

		if m.loading {
			return m, nil
		}

		if m.showTable {
			switch keyStroke {
			case "esc":
				if m.serverTable.Focused() {
					m.serverTable.Blur()
				} else {
					m.serverTable.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				return m, tea.Batch(
					tea.Printf("Let's go to %s!", m.serverTable.SelectedRow()[1]),
				)
			}
			m.serverTable, cmd = m.serverTable.Update(msg)
			return m, cmd
		}

		switch keyStroke {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			switch m.cursor {
			case 0:
				log.Printf("Showing server")
				m.loadTable()
				m.showTable = true
			case 1:
				log.Printf("Creating server")
				m.loading = true
				// if only one is needed just use tea.Cmd
				// TODO: add flag for test
				return m, tea.Batch(m.spinner.Tick, createServer)
				// return m, tea.Batch(m.spinner.Tick, createServerSimulation(4))
			case 2:
				log.Printf("Delete server")
			default:

				log.Printf("Choice not found")
			}

		}
	case string:
		if msg == SERVER_CREATED_SUCCESS {
			m.loading = false
			log.Printf("Server created successfully")
		} else if msg == SERVER_CREATED_Failed {
			m.loading = false
			log.Printf("Server creation failed")
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {

	if m.loading {
		str := fmt.Sprintf("\n\n   %s Loading server creation...press q to quit LiftOff\n\n", m.spinner.View())
		return str
	}

	if m.showTable {
		log.Printf("%s", m.serverTable.View()+" "+m.serverTable.HelpView()+"\n")
		return baseStyle.Render(m.serverTable.View()) + "\n " + m.serverTable.HelpView() + "\n"

	}

	s := "Choose hetzner action\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if i == m.cursor {
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

	p := tea.NewProgram(initialModel())
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
		{Title: "Server Type", Width: 10},
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
	m.serverTable = t

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
		// log.Println("%s", server.Name, server.Image.Name, server.Status, server.Datacenter.Name)
	}
	return rows
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

func createServerSimulation(zahl int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Duration(zahl) * time.Second)
		return SERVER_CREATED_SUCCESS
	}
}

func createServer() tea.Msg {
	// TODO: add userId to servername + guid
	hetzner_key := os.Getenv("HETZNER_CLOUD_API_KEY")
	client := hcloud.NewClient(hcloud.WithToken(hetzner_key))

	serverOption := CreateServer{ServerName: "liftTest", DeployCountry: "germany", GithubLink: "", SshKeyName: "danielwaechtler@protonmail.com"}
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
	}

	go checkAction(client, serverCreateResult.Action.ID)
	return SERVER_CREATED_SUCCESS
	//
	// RestartServer(client, serverCreateResult.Server.ID)
	// err = sshconnector.RunCommandsOnServer(serverCreateResult.Server.PublicNet.IPv4.IP.String(), []sshconnector.Command{})
	// if err != nil {
	// 	return
	// }
}

func RestartServer(client *hcloud.Client, serverID int64) {
	for {
		server, _, err := client.Server.GetByID(context.Background(), serverID)
		if err != nil {
			fmt.Println("error while reading serverinformation", err)
		}
		fmt.Println(serverID, ":", server.Status)
		if server.Status == hcloud.ServerStatusRunning {
			_, _, err := client.Server.Reboot(context.Background(), server)
			if err != nil {
				fmt.Println("error while reboot server", err)
			}
			return
		}
		time.Sleep(1 * time.Second)
	}
}

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
