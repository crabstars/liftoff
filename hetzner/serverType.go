package hetzner

import (
	"context"
	"errors"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func GetSmallestServer(client *hcloud.Client) (*hcloud.ServerType, error) {

	serverType, _, err := client.ServerType.GetByName(context.Background(), "cx22") // shared cpu intel, if u want amd use cax11
	//serverType, _, err := client.ServerType.GetByID(context.Background(), 104)
	if err != nil {
		return nil, err
	}
	if serverType == nil {
		return nil, errors.New("ServerType and error are nil")
	}

	return serverType, nil

}
