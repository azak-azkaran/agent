package main

import (
	vault "github.com/hashicorp/vault/api"
	//badger "github.com/dgraph-io/badger/v2"
	"errors"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os/exec"
	"strings"
	"syscall"
)

const (
	ERROR_VAULT_SEALED = "Vault is sealed."
)

type Configuration struct {
	agent   AgentConfig
	restic  ResticConfig
	gocrypt []GocryptConfig
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
	config.agent = *agent

	restic, err := GetResticConfig(vaultConfig, token, config.agent.restic)
	if err != nil {
		return nil, err
	}
	config.restic = *restic

	for _, name := range config.agent.gocryptfs {
		gocrypt, err := GetGocryptConfig(vaultConfig, token, name)
		if err != nil {
			return nil, err
		}
		config.gocrypt = append(config.gocrypt, *gocrypt)
	}
	return &config, nil
}

func main() {
	log.Print("Please enter Token: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		log.Fatal(err)
	}
	token := strings.TrimSpace(string(password))
	log.Printf("token %s", token)

	resp, err := IsSealed(vault.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Check if Vault is unsealed")
	if resp {
		log.Println("Vault is sealed")
	}
}
