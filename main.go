package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
)

type logWriter struct {
	io.Writer
}

func (w logWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	hetzner_key := os.Getenv("HETZNER_CLOUD_API_KEY")
	client := hcloud.NewClient(hcloud.WithToken(hetzner_key))
	datacenters, err := client.Datacenter.All(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-----Datacenters------")
	for _, datacenter := range datacenters {
		fmt.Println(
			datacenter.Name,
			datacenter.ID,
			datacenter.Location.Name,
			datacenter.Location.Description,
		)
	}

	fmt.Println("-----Images------")
	images, err := client.Image.All(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, image := range images {
		fmt.Println(image.Name)
	}

	fmt.Println("-----Server Types------")
	serverTypes, err := client.ServerType.All(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, serverType := range serverTypes {
		fmt.Println(serverType.Name, serverType.Cores, serverType.Memory)
	}

	fmt.Println("-----SSH Keys------")
	sshKeys, err := client.SSHKey.All(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, sshKey := range sshKeys {
		fmt.Println(sshKey.Name, sshKey.PublicKey)
	}

	automount := false
	serverCreateResult, _, err := client.Server.Create(
		context.Background(),
		hcloud.ServerCreateOpts{
			Name:       "LiftTest",
			Automount:  &automount,
			Datacenter: datacenters[0],
			Image:      images[1], // get image later by image Name
			// Location:   datacenters[0].Location,
			ServerType: serverTypes[0],
			SSHKeys: []*hcloud.SSHKey{
				sshKeys[0],
			},
			// UserData: "",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(serverCreateResult.RootPassword, serverCreateResult.Server.PublicNet.IPv4.IP)
	ip := serverCreateResult.Server.PublicNet.IPv4.IP

	for i := 0; i < 60; i++ {
		newServer, _, _ := client.Server.GetByID(context.Background(), serverCreateResult.Server.ID)
		fmt.Println(newServer.Status)
		time.Sleep(1 * time.Second)
		if newServer.Status == "running" {
			break
		}

	}
	fmt.Println(ip.String())
	fmt.Println("Finish Creating Server")
	createSSHConnection(ip.String())
}

func createSSHConnection(serverIP string) {
	ssh_key_path := os.Getenv("SSH_KEY_PATH")
	if len(ssh_key_path) == 0 {
		log.Fatalf("ssh path is empyt")
	}
	if ssh_key_path[0] == '~' {
		user, err := user.Current()
		if err != nil {
			log.Fatalf("ssh path is empyt")
		}

		ssh_key_path = filepath.Join(user.HomeDir, ssh_key_path[1:])
	}

	key, err := os.ReadFile(
		"/home/kami/.ssh/id_ed25519",
	) // TODO add as default and else get from env file
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO
	}

	var client *ssh.Client
	for i := 0; i < 10; i++ {

		client, err = ssh.Dial("tcp", serverIP+":22", config)

		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			defer client.Close()
			break
		}
	}

	fmt.Println("SSH connection established")

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	var outputBuffer bytes.Buffer
	// MultiWriter allows us to capture the output later again
	session.Stdout = io.MultiWriter(&outputBuffer, logWriter{})
	session.Stderr = io.MultiWriter(&outputBuffer, logWriter{})

	// Run the command
	if err := session.Run("apt update -y && apt install -y ansible"); err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}
	// TODO capture if error and log to file
	// capturedOutput := outputBuffer.String()
	// fmt.Println("Captured output:", capturedOutput)

	fmt.Println("finished installing ansible")
}
