package hetzner

import (
	"context"
	"fmt"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

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
