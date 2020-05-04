package main

import (
	vault "github.com/hashicorp/vault/api"
	//badger "github.com/dgraph-io/badger/v2"
	"errors"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	ERROR_VAULT_SEALED = "Vault is sealed."
)

type Configuration struct {
	Agent   AgentConfig
	Restic  ResticConfig
	Gocrypt []GocryptConfig
}

func RunJob(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func GetConfigFromVault(token string, hostname string, vaultConfig *vault.Config) (*Configuration, error) {
	resp, err := IsSealed(vaultConfig)
	if err != nil {
		return nil, err
	}
	if resp {
		return nil, errors.New(ERROR_VAULT_SEALED)
	}

	var config Configuration
	agent, err := GetAgentConfig(vaultConfig, token, hostname)
	if err != nil {
		return nil, err
	}
	config.Agent = *agent

	restic, err := GetResticConfig(vaultConfig, token, config.Agent.Restic)
	if err != nil {
		return nil, err
	}
	config.Restic = *restic

	for _, name := range config.Agent.Gocryptfs {
		gocrypt, err := GetGocryptConfig(vaultConfig, token, name)
		if err != nil {
			return nil, err
		}
		gocrypt.AllowOther = true
		gocrypt.Duration, err = time.ParseDuration("0s")
		if err != nil {
			return nil, err
		}
		config.Gocrypt = append(config.Gocrypt, *gocrypt)
	}
	return &config, nil
}

func main() {
	name, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Agent starting on: ", name)
	log.Print("Please enter Token: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Error: ", err)
	}
	token := strings.TrimSpace(string(password))

	vaultConfig := vault.DefaultConfig()
	log.Println("Getting Vault configuration for agent: ", name, " from: ", vaultConfig.Address)
	config, err := GetConfigFromVault(token, name, vaultConfig)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	cmds, empties, err := MountFolders(config.Gocrypt)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	for i, cmd := range cmds {
		if empties[i] {
			log.Println("Running: ", cmd.String())
			RunJob(&cmd)
		}
	}
}
