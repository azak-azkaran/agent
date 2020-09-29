package main

import (
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"

	vault "github.com/hashicorp/vault/api"
)

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
}

func Seal(config *vault.Config, token string) error {
	client, err := vault.NewClient(config)
	if err != nil {
		return err
	}
	client.SetToken(token)

	sys := client.Sys()
	return sys.Seal()
}

func Unseal(config *vault.Config, key string) (*vault.SealStatusResponse, error) {
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	sys := client.Sys()
	resp, err := sys.Unseal(key)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func SealStatus(config *vault.Config) (*vault.SealStatusResponse, error) {
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	sys := client.Sys()
	respones, err := sys.SealStatus()
	if err != nil {
		return nil, err
	}
	return respones, nil

}

func IsSealed(config *vault.Config) (bool, error) {
	client, err := vault.NewClient(config)
	if err != nil {
		return true, err
	}

	sys := client.Sys()
	respones, err := sys.SealStatus()
	if err != nil {
		return true, err
	}
	return respones.Sealed, nil
}
func GetSecret(config *vault.Config, token string, path string) (*vault.Secret, error) {
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	logical := client.Logical()
	secret, err := logical.Read(path)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func getDataFromSecret(config *vault.Config, token string, path string) (map[string]interface{}, error) {
	//log.Println("Getting Data from: ", path)
	secret, err := GetSecret(config, token, path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, errors.New(ERROR_VAULT_NO_SECRET)
	}

	if _, ok := secret.Data["data"]; ok {
		data := secret.Data["data"].(map[string]interface{})
		if len(data) == 0 {
			return nil, errors.New("Data of secret with path: " + path + " is empty")
		}
		return data, nil
	} else {
		return secret.Data, nil
	}
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
	return &conf, nil
}
