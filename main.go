package main

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"errors"
	"log"
	"os"
	"os/exec"
	"time"

	cqueue "github.com/enriquebris/goconcurrentqueue"
)

const (
	ERROR_VAULT_SEALED = "Vault is sealed."
)

var AgentConfiguration Configuration
var ConcurrentQueue *cqueue.FIFO

type Configuration struct {
	Agent       *AgentConfig
	Restic      *ResticConfig
	Gocrypt     []GocryptConfig
	VaultConfig *vault.Config
	Hostname    string
	Address     string
	Token       string
}

func RunJob(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func Init(vaultConfig *vault.Config, args []string) error {
	ConcurrentQueue = cqueue.NewFIFO()
	addressCommend := pflag.NewFlagSet("address", pflag.ContinueOnError)
	addressCommend.String("address", "localhost:8081", "the addess on which rest server of the agent is startet")
	viper.SetEnvPrefix("agent")
	err := viper.BindEnv("address")
	if err != nil {
		return err
	}
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	log.Println("Agent initalzing on: ", hostname)
	AgentConfiguration = Configuration{
		VaultConfig: vaultConfig,
		Hostname:    hostname,
	}
	err = addressCommend.Parse(args)
	if err != nil {
		return err
	}

	if viper.IsSet("address") {
		AgentConfiguration.Address = viper.GetString("address")
	} else {
		AgentConfiguration.Address = "localhost:8081"
	}
	return nil
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
	config.Agent = agent
	restic, err := GetResticConfig(vaultConfig, token, config.Agent.Restic)
	if err != nil {
		return nil, err
	}
	config.Restic = restic

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
	err := Init(vault.DefaultConfig(), os.Args)
	//log.Print("Please enter Token: ")
	//password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Error: ", err)
	}
	_, fun := RunRestServer("localhost:8081")

	//token := strings.TrimSpace(string(password))
	fun()

	//log.Println("Getting Vault configuration for agent: ", name, " from: ", vaultConfig.Address)
	//config, err := GetConfigFromVault(token, name, vaultConfig)
	//if err != nil {
	//  log.Fatal(err)
	//  panic(err)
	//}
	//out, err := MountFolders(config.Gocrypt, RunJob)
	//if err != nil {
	//  log.Fatal(err)
	//  panic(err)
	//}
	//for o := range out {
	//  log.Println(o)
	//}
}
