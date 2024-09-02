package model

import (
	"fmt"
	"log"
	"strings"
)

func (m Model) ViewState() string {
	var builder strings.Builder

	builder.WriteString("Model:\n")

	builder.WriteString("  Spinner:\n")
	builder.WriteString(fmt.Sprintf("    Current Frame: %s\n", m.Spinner.View()))

	builder.WriteString("  CreateServerState:\n")
	builder.WriteString(fmt.Sprintf("    WaitingForServerNameInput: %v\n", m.CreateServerState.WaitingForServerNameInput))
	builder.WriteString(fmt.Sprintf("    ServerNameInput: %s\n", m.CreateServerState.ServerNameInput.Value()))
	builder.WriteString(fmt.Sprintf("    CreatingServer: %v\n", m.CreateServerState.CreatingServer))

	builder.WriteString("  ActionSelectionState:\n")
	builder.WriteString(fmt.Sprintf("    Choices: %s\n", strings.Join(m.ActionSelectionState.Choices, ", ")))
	builder.WriteString(fmt.Sprintf("    Cursor: %d\n", m.ActionSelectionState.Cursor))

	builder.WriteString("  TableState:\n")
	builder.WriteString(fmt.Sprintf("    TableReloadRunning: %v\n", m.TableState.TableReloadRunning))
	builder.WriteString(fmt.Sprintf("    ShowTable: %v\n", m.TableState.ShowTable))
	builder.WriteString(fmt.Sprintf("    RowCursor: %d\n", m.TableState.RowCursor))

	builder.WriteString("\n\n")
	return builder.String()

}

func (m Model) ViewHandleCreateServerState() string {

	if m.CreateServerState.WaitingForServerNameInput {
		return fmt.Sprintf("Enter Server name:\n\n%s\n\n%s", m.CreateServerState.ServerNameInput.View(), "(esc to quit)")
	}

	if m.CreateServerState.CreatingServer {
		return fmt.Sprintf("\n\n   %s Loading Server creation...press q to quit LiftOff\n\n", m.Spinner.View())
	}

	return ""
}

func (m Model) View() string {
	s := ""
	if m.EnvValues.Debug {
		s = m.ViewState()
	}

	if state := m.ViewHandleCreateServerState(); state != "" {
		s += state
		return s
	}
	if m.TableState.ShowTable {
		log.Printf("%s", m.TableState.ServerTable.View()+" "+m.TableState.ServerTable.HelpView()+"\n")
		s += baseStyle.Render(m.TableState.ServerTable.View()) + "\n " + m.TableState.ServerTable.HelpView() + "\n"
		return s

	}

	s += "Choose hetzner Action\n\n"

	for i, choice := range m.ActionSelectionState.Choices {
		cursor := " "
		if m.ActionSelectionState.Cursor == i {
			cursor = ">"
		}

		checked := " "
		if i == m.ActionSelectionState.Cursor {
			checked = "x"
		}
		s += fmt.Sprintf("%s [%s], %s\n", cursor, checked, choice)
	}
	s += "\nPress q to quit.\n"
	return s
}
