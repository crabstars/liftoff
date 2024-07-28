package hetzner

import (
	"context"
	"errors"
	"log"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

const (
	CountryUSA     = "us"
	CountryGermany = "germany"
)

// use interface types for country ???
func GetDatacenter(client *hcloud.Client, country string) (*hcloud.Datacenter, error) {
	var datacenterName string

	switch country {
	case CountryUSA:
		datacenterName = "ash-dc1" // this is us-east other one is hil-dc1 (which is us west)
	case CountryGermany:
		datacenterName = "nbg1-dc3" // other one is fsn1-dc14
		// helsinki is hel1-dc2
	default:
		return nil, errors.New("invalid country specified, must be 'us' or 'germany'")
	}

	datacenter, _, err := client.Datacenter.GetByName(context.Background(), datacenterName)

	if err != nil {
		log.Println("error while getting datacenter")
		return nil, err
	}

	if datacenter == nil {
		return nil, errors.New("error and datacenter are nil")
	}

	return datacenter, nil
}
