package model

import (
	"log"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crabstars/liftoff/hetzner"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case TableUpdateMsg:
		m.loadTableWithoutFetch(msg)
		m.TableState.TableReloadRunning = false

	case TickMsg:
		if m.TableState.ShowTable && !m.TableState.TableReloadRunning {
			m.TableState.TableReloadRunning = true
			go m.fetchTableRows()
		}
		return m, tickEvery(time.Second * 3)

	case tea.KeyMsg:
		keyStroke := msg.String()

		if keyStroke == "q" || keyStroke == "ctrl+c" {
			return m, tea.Quit
		}

		if m.CreateServerState.CreatingServer {
			return m, nil
		}

		if m.CreateServerState.WaitingForServerNameInput {
			if !m.CreateServerState.ServerNameInput.Focused() {
				m.CreateServerState.ServerNameInput.Focus()
			}
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.CreateServerState.WaitingForServerNameInput = false
				return m, nil
			case tea.KeyEnter:
				m.CreateServerState.WaitingForServerNameInput = false
				m.CreateServerState.CreatingServer = true

				return m, tea.Batch(m.Spinner.Tick, hetzner.CreateServer(m.EnvValues.HetznerApiKey, m.EnvValues.SshKeyName, m.CreateServerState.ServerNameInput.Value()))
			}
			m.CreateServerState.ServerNameInput, cmd = m.CreateServerState.ServerNameInput.Update(msg)

			return m, cmd

		}

		if m.TableState.ShowTable {
			switch keyStroke {
			case "esc":
				m.TableState.ShowTable = false
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				return m, tea.Batch(
					tea.Printf("Let's go to %s!", m.TableState.ServerTable.SelectedRow()[1]),
				)
			}
			m.TableState.ServerTable, cmd = m.TableState.ServerTable.Update(msg)
			m.TableState.RowCursor = m.TableState.ServerTable.Cursor()
			return m, cmd
		}

		switch keyStroke {
		case "up", "k":
			if m.ActionSelectionState.Cursor > 0 {
				m.ActionSelectionState.Cursor--
			}

		case "down", "j":
			if m.ActionSelectionState.Cursor < len(m.ActionSelectionState.Choices)-1 {
				m.ActionSelectionState.Cursor++
			}

		case "enter", " ":
			switch m.ActionSelectionState.Cursor {
			case 0:
				log.Printf("Showing Server")
				m.TableState.ShowTable = true
				go m.fetchTableRows()
			case 1:
				m.CreateServerState.WaitingForServerNameInput = true
				log.Printf("Waiting for name input")
				return m, nil
			case 2:
				log.Printf("Delete Server")
			default:

				log.Printf("Choice not found")
			}

		}
	case string:
		if msg == hetzner.SERVER_CREATED_SUCCESS {
			m.CreateServerState.ServerNameInput.Reset()
			m.CreateServerState.CreatingServer = false
			log.Printf("Server created successfully")
		} else if msg == hetzner.SERVER_CREATED_Failed {
			m.CreateServerState.ServerNameInput.Reset()
			m.CreateServerState.CreatingServer = false
			log.Printf("Server creation failed")
		}

	case spinner.TickMsg:
		if m.CreateServerState.CreatingServer {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}
