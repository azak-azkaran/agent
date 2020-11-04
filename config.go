package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Configuration struct {
	Agent            *AgentConfig
	Restic           *ResticConfig
	Gocrypt          []GocryptConfig
	Git              []GitConfig
	VaultConfig      *vault.Config
	DB               *badger.DB
	Hostname         string
	Address          string
	Token            string
	PathDB           string
	TimeBetweenStart time.Duration
	Timer            *time.Timer
	MountAllow       bool
	MountDuration    string
	VaultKeyFile     string
	RoleID           string
	SecretID         string
	useLogin         bool
}

type AgentConfig struct {
	Gocryptfs  string `mapstructure:"gocryptfs"`
	Restic     string `mapstructure:"restic"`
	Git        string `mapstructure:"git"`
	HomeFolder string `mapstructure:"home"`
}

type GocryptConfig struct {
	MountPoint    string `mapstructure:"mount-path"`
	Path          string `mapstructure:"path"`
	Password      string `mapstructure:"pw"`
	AllowOther    bool   `mapstructure:"allow"`
	NotEmpty      bool   `mapstructure:"notempty"`
	Duration      string `mapstructure:"duration"`
	MountDuration time.Duration
}

type ResticConfig struct {
	Password    string `mapstructure:"pw"`
	Path        string `mapstructure:"path"`
	Repo        string `mapstructure:"repo"`
	ExcludePath string `mapstructure:"exclude"`
	SecretKey   string `mapstructure:"secret_key"`
	AccessKey   string `mapstructure:"access_key"`
	Environment []string
}

type GitConfig struct {
	Rep           string `mapstructure:"repo"`
	Directory     string `mapstructure:"dir"`
	PersonalToken string `mapstructure:"personal_token"`
	Name          string
}

func GetGocryptConfig(config *vault.Config, token string, path string) (*GocryptConfig, error) {
	data, err := getDataFromSecret(config, token, "gocrypt/data/"+path)
	if err != nil {
		return nil, err
	}

	var conf GocryptConfig
	decoderConfig := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &conf,
	}

	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	if conf.Duration == "" {
		conf.MountDuration, err = time.ParseDuration("0s")
	} else {

		conf.MountDuration, err = time.ParseDuration(conf.Duration)
	}
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func GetAgentConfig(config *vault.Config, token string, path string) (*AgentConfig, error) {
	data, err := getDataFromSecret(config, token, "config/"+path)
	if err != nil {
		return nil, err
	}

	var conf AgentConfig
	err = mapstructure.Decode(data, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func GetGitConfig(config *vault.Config, token string, path string) (*GitConfig, error) {
	data, err := getDataFromSecret(config, token, "git/data/"+path)
	if err != nil {
		return nil, err
	}

	var conf GitConfig
	decoderConfig := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &conf,
	}

	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	conf.Name = path
	return &conf, nil
}

func GetResticConfig(config *vault.Config, token string, path string) (*ResticConfig, error) {
	data, err := getDataFromSecret(config, token, "restic/data/"+path)
	if err != nil {
		return nil, err
	}

	var conf ResticConfig
	err = mapstructure.Decode(data, &conf)
	if err != nil {
		return nil, err

	}
	conf.Environment = []string{
		RESTIC_ACCESS_KEY + conf.AccessKey,
		RESTIC_SECRET_KEY + conf.SecretKey,
		RESTIC_REPOSITORY + conf.Repo,
		RESTIC_PASSWORD + conf.Password,
	}

	if data["exclude"] != nil {
		conf.ExcludePath = data["exclude"].(string)

	}
	return &conf, nil

}

func CreateConfigFromVault(token string, hostname string, vaultConfig *vault.Config) (*Configuration, error) {
	config := Configuration{
		VaultConfig: vaultConfig,
		Token:       token,
		Hostname:    hostname,
	}

	if err := config.VaultReady(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (config *Configuration) VaultReady() error {
	if config.VaultConfig == nil {
		return errors.New(ERROR_VAULT_CONFIG_MISSING)
	}
	resp, err := IsSealed(config.VaultConfig)
	if err != nil {
		return err
	}
	if resp {
		return errors.New(ERROR_VAULT_SEALED)
	}
	return nil
}

func (config *Configuration) GetAgentConfig() error {
	if err := config.VaultReady(); err != nil {
		return err
	}

	agent, err := GetAgentConfig(config.VaultConfig, config.Token, config.Hostname)
	if err != nil {
		return err
	}

	config.Agent = agent
	return nil
}

func (config *Configuration) GetResticConfig() error {
	if err := config.VaultReady(); err != nil {
		return err
	}
	err := config.GetAgentConfig()
	if err != nil {
		return err
	}
	restic, err := GetResticConfig(config.VaultConfig, config.Token, config.Agent.Restic)
	if err != nil {
		return err
	}
	config.Restic = restic
	return nil
}

func (config *Configuration) GetGocryptConfig() error {
	if err := config.VaultReady(); err != nil {
		return err
	}
	err := config.GetAgentConfig()
	if err != nil {
		return err
	}
	crypts := strings.Split(config.Agent.Gocryptfs, ",")
	for _, name := range crypts {
		gocrypt, err := GetGocryptConfig(config.VaultConfig, config.Token, name)
		if err != nil {
			return err
		}
		config.Gocrypt = append(config.Gocrypt, *gocrypt)
	}
	return nil
}

func (config *Configuration) GetGitConfig() error {
	if err := config.VaultReady(); err != nil {
		return err
	}
	err := config.GetAgentConfig()
	if err != nil {
		return err
	}

	gits := strings.Split(config.Agent.Git, ",")
	for _, name := range gits {
		git, err := GetGitConfig(config.VaultConfig, config.Token, name)
		if err != nil {
			return err
		}
		config.Git = append(config.Git, *git)
	}
	return nil
}

func CreateConfigFullFromVault(token string, hostname string, vaultConfig *vault.Config) (*Configuration, error) {
	config, err := CreateConfigFromVault(token, hostname, vaultConfig)
	if err != nil {
		return nil, err
	}

	err = config.GetResticConfig()
	if err != nil {
		return nil, err
	}

	err = config.GetGocryptConfig()
	if err != nil {
		return nil, err
	}

	err = config.GetGitConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func ParseConfiguration(confi *Configuration) {
	if viper.IsSet(MAIN_ADDRESS) {
		confi.Address = viper.GetString(MAIN_ADDRESS)
	} else {
		confi.Address = "localhost:8081"
	}

	if viper.IsSet(MAIN_PATHDB) {
		confi.PathDB = viper.GetString(MAIN_PATHDB)
	} else {
		confi.PathDB = "/opt/agent/db"
	}

	if viper.IsSet(MAIN_TIME_DURATION) {
		dur, err := time.ParseDuration(viper.GetString(MAIN_TIME_DURATION))
		if err != nil {
			log.Println("Error parsing duration: ", err)
			dur = 30 * time.Minute
		}
		confi.TimeBetweenStart = dur
	} else {
		confi.TimeBetweenStart = 30 * time.Minute
	}

	if viper.IsSet(MAIN_MOUNT_DURATION) {
		confi.MountDuration = viper.GetString(MAIN_MOUNT_DURATION)
	} else {
		confi.MountDuration = ""
	}

	if viper.IsSet(MAIN_MOUNT_ALLOW) {
		confi.MountAllow = viper.GetBool(MAIN_MOUNT_ALLOW)
	} else {
		confi.MountAllow = false
	}

	if viper.IsSet(MAIN_VAULT_KEY_FILE) {
		confi.VaultKeyFile = viper.GetString(MAIN_VAULT_KEY_FILE)
	} else {
		confi.VaultKeyFile = ""
	}

	if viper.IsSet(MAIN_VAULT_ADDRESS) {
		confi.VaultConfig.Address = viper.GetString(MAIN_VAULT_ADDRESS)
	}

	if viper.IsSet(MAIN_VAULT_ROLE_ID) {
		confi.RoleID = viper.GetString(MAIN_VAULT_ROLE_ID)
	}

	if viper.IsSet(MAIN_VAULT_SECRET_ID) {
		confi.SecretID = viper.GetString(MAIN_VAULT_SECRET_ID)
	}

	log.Println("Agent initalzing on: ", confi.Hostname)
	log.Println("Agent Configuration:",
		"\nAddress: ", confi.Address,
		"\nPath to DB:", confi.PathDB,
		"\nTime Between Backup Runs: ", confi.TimeBetweenStart,
		"\nVault Address: ", confi.VaultConfig.Address,
		"\nVault KeyFile path: ", confi.VaultKeyFile,
		"\nMount Duration: ", confi.MountDuration,
		"\nMount AllowOther: ", confi.MountAllow,
	)

	if (confi.RoleID == "" && confi.SecretID == "") || (confi.RoleID == "" && confi.SecretID != "") || (confi.RoleID != "" && confi.SecretID == "") {
		confi.RoleID = ""
		confi.SecretID = ""
		log.Println("Secret ID and Role ID reset")
		confi.useLogin = false
	} else {
		log.Println("RoleID: ", confi.RoleID)
		log.Println("SecretID: ", confi.SecretID)
		confi.useLogin = true
	}
}
