package cloudconfig

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	Basic = "basic"
)

type Config struct {
	Users          []User   `yaml:"users"`
	Packages       []string `yaml:"packages"`
	PackageUpdate  bool     `yaml:"package_update"`
	PackageUpgrade bool     `yaml:"package_upgrade"`
	RunCmd         []string `yaml:"runcmd"`
}

type User struct {
	Name              string   `yaml:"name"`
	Groups            string   `yaml:"groups"`
	Sudo              string   `yaml:"sudo"`
	Shell             string   `yaml:"shell"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

func GetConfig(configType string, ssh_key string) (string, error) {

	var cloudConfig []byte
	var err error
	switch configType {
	case Basic:
		cloudConfig, err = os.ReadFile("cloudConfig/basic.yaml")

	default:
		return "", errors.New("no matching config type")
	}
	fmt.Println(string(cloudConfig))

	if err != nil {
		return "", err
	}

	var yamlConf Config
	err = yaml.Unmarshal(cloudConfig, &yamlConf)
	if err != nil {
		return "", err
	}

	if len(yamlConf.Users) == 0 {
		return "", errors.New("Config needs at least one user")
	}

	yamlConf.Users[0].SSHAuthorizedKeys = append(yamlConf.Users[0].SSHAuthorizedKeys, ssh_key)
	stringConf, err := yaml.Marshal(&yamlConf)
	if err != nil {
		return "", err
	}
	return string(stringConf), nil
}
