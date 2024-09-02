package hetzner

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/table"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func ListServer(hetzner_cloud_api_key string) []table.Row {
	var rows []table.Row
	client := hcloud.NewClient(hcloud.WithToken(hetzner_cloud_api_key))
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
