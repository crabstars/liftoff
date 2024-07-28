package hetzner

import (
	"context"
	"errors"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func GetDockerCeImage(client *hcloud.Client) (*hcloud.Image, error) {

	image, _, err := client.Image.GetByID(context.Background(), 40093247) // for intel, use 105888141 for amd
	if err != nil {
		return nil, err
	}
	if image == nil {
		return nil, errors.New("image and error are nil")
	}
	return image, nil
}
