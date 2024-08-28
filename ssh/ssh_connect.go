package sshconnector

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/crabstars/liftoff/logging"
	"golang.org/x/crypto/ssh"
)

// const UserName = "liftoff"
const UserName = "root"
const sshPort = "22"
const protocol = "tcp"
const retryCount = 30

func getSshClientConfi() (*ssh.ClientConfig, error) {

	ssh_key_path := os.Getenv("SSH_KEY_PATH")
	if len(ssh_key_path) == 0 {
		log.Fatalf("ssh path is empty")
	}
	if ssh_key_path[0] == '~' {
		user, err := user.Current()
		if err != nil {
			log.Println("ssh path is empty")
			return nil, err
		}
		ssh_key_path = filepath.Join(user.HomeDir, ssh_key_path[1:])
	}

	key, err := os.ReadFile(
		ssh_key_path,
	)
	if err != nil {
		log.Println("Unable to read private key:", err)
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Println("Unable to parse private key:", err)
		return nil, err
	}

	return &ssh.ClientConfig{
		User: UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO:
		Timeout:         time.Duration(time.Second * 10),
	}, nil
}

// Caller needs to call defer client.Close()
func EstablishSshConnection(serverIP string) (*ssh.Client, error) {
	config, err := getSshClientConfi()
	if err != nil {
		return nil, err
	}
	var client *ssh.Client
	addr := serverIP + ":" + sshPort
	for i := 0; i < retryCount; i++ {

		client, err = ssh.Dial(protocol, addr, config)

		log.Println("Trying to establish ssh connection...")
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		log.Println("Dial failed to create client ", err)
		return nil, err
	}

	log.Println("SSH connection established")
	return client, nil

}

func ExecuteCommand(client *ssh.Client, command Command) error {

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	var outputBuffer bytes.Buffer
	session.Stdout = io.MultiWriter(&outputBuffer, logging.LogWriter{})
	session.Stderr = io.MultiWriter(&outputBuffer, logging.LogWriter{})
	if err := session.Run(command.cmd); err != nil {
		return err
	}
	log.Println(command.successMessage)
	return nil

}

type Command struct {
	cmd            string
	successMessage string
}

func RunCommandsOnServer(serverIP string, commands []Command) error {

	commands = []Command{
		// {"apt update && apt upgrade -y && apt install git -y", "system updated"},
		{"git clone https://github.com/crabstars/ExampleCSharpWeather.git", "git repo pulled"},
		{"cd /root/ExampleCSharpWeather && docker build -t exampledotnet -f dotnet.Dockerfile .", "build docker image done"},
		{"cd /root/ExampleCSharpWeather && docker run -p 5021:5021 -d exampledotnet", "docker container is running"},
	}
	client, err := EstablishSshConnection(serverIP)
	if err != nil {
		return err
	}
	defer client.Close()

	for _, command := range commands {

		err = ExecuteCommand(client, command)
		if err != nil {
			return err
		}
	}

	log.Println("finished starting api")
	return nil
}
