package hetzner

import (
	"context"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// TODO: use a struct msg
var (
	SERVER_CREATED_SUCCESS = "server-created-success"
	SERVER_CREATED_Failed  = "server-created-failed"
)

type CreateServerModel struct {
	ServerName    string `json:"serverName"`
	DeployCountry string `json:"deployCountry"`
	GithubLink    string `json:"githubLink"`
	SshKeyName    string `json:"sshKeyName"`
}

func CreateServer(hetzner_cloud_api_key string, ssh_key_name string, serverName string) tea.Cmd {
	return func() tea.Msg {
		return createHetznerServer(hetzner_cloud_api_key, ssh_key_name, serverName)
	}
}

func createHetznerServer(hetzner_cloud_api_key string, ssh_key_name string, serverName string) tea.Msg {
	client := hcloud.NewClient(hcloud.WithToken(hetzner_cloud_api_key))

	serverOption := CreateServerModel{ServerName: serverName, DeployCountry: "germany", GithubLink: "", SshKeyName: ssh_key_name}
	serverType, err := GetSmallestServer(client)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	image, err := GetDockerCeImage(client)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}
	datacenter, err := GetDatacenter(client, serverOption.DeployCountry)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	sshKey, err := GetSshKey(client, ssh_key_name)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	automount := false
	serverCreateResult, _, err := client.Server.Create(
		context.Background(),
		hcloud.ServerCreateOpts{
			Name:       serverOption.ServerName,
			Automount:  &automount, // volumes for mounting
			Datacenter: datacenter,
			Image:      image,
			ServerType: serverType,
			SSHKeys: []*hcloud.SSHKey{
				sshKey,
			},
			// UserData: cloudConfig,
		},
	)
	if err != nil {
		log.Println(err.Error())
		return SERVER_CREATED_Failed
	}

	go checkAction(client, serverCreateResult.Action.ID)
	return SERVER_CREATED_SUCCESS
	//
	// RestartServer(client, serverCreateResult.Server.ID)
	// err = sshconnector.RunCommandsOnServer(serverCreateResult.Server.PublicNet.IPv4.IP.String(), []sshconnector.Command{})
	// if err != nil {
	// 	return
	// }
}

func checkAction(client *hcloud.Client, actionID int64) {
	for {
		action, _, err := client.Action.GetByID(context.Background(), actionID)
		if err != nil {
			log.Println("Checking action failed", actionID)
			return
		}
		if action.Status == hcloud.ActionStatusSuccess || action.Status == hcloud.ActionStatusError {
			log.Println("Action success")
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func createServerSimulation(zahl int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Duration(zahl) * time.Second)
		return SERVER_CREATED_SUCCESS
	}
}
