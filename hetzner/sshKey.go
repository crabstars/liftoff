package hetzner

import (
	"context"
	"errors"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func GetSshKey(client *hcloud.Client, name string) (*hcloud.SSHKey, error) {

	sshKey, _, err := client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		return nil, err
	}
	if sshKey == nil {
		return nil, errors.New("SshKey and error are nil")
	}

	return sshKey, nil

}
