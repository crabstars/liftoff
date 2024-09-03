package model

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/crabstars/liftoff/hetzner"
)

func (m *Model) loadTableWithoutFetch(rows []table.Row) {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Image", Width: 20},
		{Title: "Status", Width: 15},
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

	if len(rows)-1 < m.TableState.RowCursor {
		t.SetCursor(len(rows) - 1)
	} else {
		t.SetCursor(m.TableState.RowCursor)
	}

	m.TableState.ServerTable = t
}

func (m *Model) fetchTableRows() {
	servers, err := hetzner.ListServer(m.EnvValues.HetznerApiKey)
	if err != nil {
		log.Println("Failed to load server", err.Error())
	}
	rows := make([]table.Row, len(servers))
	serverIndexIdRelations := make([]int64, len(servers))
	for i, server := range servers {
		rows[i] = table.Row{server.Name, server.Image.Name, string(server.Status), server.Datacenter.Location.City, string(server.ServerType.CPUType), server.ServerType.Name, fmt.Sprintf("%d", server.ServerType.Cores), fmt.Sprintf("%.0f GB", server.ServerType.Memory), fmt.Sprintf("%d GB", server.ServerType.Disk)}
		serverIndexIdRelations[i] = server.ID
	}

	m.Program.Send(TableUpdateMsg{rows, serverIndexIdRelations})
}
