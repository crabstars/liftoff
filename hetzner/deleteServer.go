package hetzner

import (
	"context"
	"log"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func deleteServer(hetzner_cloud_api_key string, serverID int) {
	client := hcloud.NewClient(hcloud.WithToken(hetzner_cloud_api_key))
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
