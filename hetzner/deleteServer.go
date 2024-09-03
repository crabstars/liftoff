package hetzner

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type ServerDeletedSuccessMsg struct{}
type ServerDeletedErrorMsg struct{}

func DeleteServer(hetzner_cloud_api_key string, serverID int64) tea.Cmd {
	return func() tea.Msg {
		return deleteServerHetzner(hetzner_cloud_api_key, serverID)
	}
}
func deleteServerHetzner(hetzner_cloud_api_key string, serverID int64) tea.Msg {
	client := hcloud.NewClient(hcloud.WithToken(hetzner_cloud_api_key))
	server, _, err := client.Server.GetByID(context.Background(), serverID)
	if err != nil {
		log.Println("could not get server for deleting", err)
		return ServerDeletedErrorMsg{}
	}
	_, _, err = client.Server.DeleteWithResult(context.Background(), server)
	if err != nil {
		log.Println("could not delete server", err)
		return ServerDeletedErrorMsg{}
	}

	return ServerDeletedSuccessMsg{}
}
func DeleteServer2(hetzner_cloud_api_key string, serverID int64) error {
	client := hcloud.NewClient(hcloud.WithToken(hetzner_cloud_api_key))
	server, _, err := client.Server.GetByID(context.Background(), serverID)
	if err != nil {
		log.Println("could not get server for deleting", err)
		return err
	}
	_, _, err = client.Server.DeleteWithResult(context.Background(), server)
	if err != nil {
		log.Println("could not delete server", err)
		return err
	}

	return nil
}
