package main

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"
)

type AgentConfig struct {
	Gocryptfs []string
	Restic    string
}

type GocryptConfig struct {
	MountPoint string
	Path       string
	Password   string
	AllowOther bool
	Duration   time.Duration
}

type ResticConfig struct {
	Password    string
	Path        string
	Repo        string
	ExcludePath string
	SecretKey   string
	AccessKey   string
	Environment []string
}

func CheckMap(list []string, data map[string]interface{}) error {
	var message bytes.Buffer
	message.WriteString("Data in Vault malformed")
	fail := false
	for _, v := range list {
		if data[v] == nil {
			message.WriteString("\n\t" + v + ": missing")
			fail = true
		}
	}

	if fail {
		log.Println(message.String())
		return errors.New(message.String())
	}
	return nil
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

	list := []string{"repo", "path", "exclude", "pw", "access_key", "secret_key"}
	err = CheckMap(list, data)
	if err != nil {
		return nil, err
	}

	conf := ResticConfig{
		Repo:        data[list[0]].(string),
		Path:        data[list[1]].(string),
		ExcludePath: data[list[2]].(string),
		Password:    data[list[3]].(string),
		AccessKey:   data[list[4]].(string),
		SecretKey:   data[list[5]].(string),
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

	list := []string{"mount-path", "path", "pw"}
	err = CheckMap(list, data)
	if err != nil {
		return nil, err
	}

	conf := GocryptConfig{
		MountPoint: data[list[0]].(string),
		Path:       data[list[1]].(string),
		Password:   data[list[2]].(string),
	}
	return &conf, nil
}

func GetAgentConfig(config *vault.Config, token string, path string) (*AgentConfig, error) {
	data, err := getDataFromSecret(config, token, "config/"+path)
	if err != nil {
		return nil, err
	}

	conf := AgentConfig{
		Restic:    data["restic"].(string),
		Gocryptfs: strings.Split(data["gocryptfs"].(string), ","),
	}
	return &conf, nil
}
