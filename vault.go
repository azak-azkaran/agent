package main

import (
	"errors"
	vault "github.com/hashicorp/vault/api"
)

type GocryptConfig struct {
	mountPoint string
	path       string
	password   string
}

type ResticConfig struct {
	password string
	path     string
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
		return false, err
	}

	sys := client.Sys()
	respones, err := sys.SealStatus()
	if err != nil {
		return false, err
	}
	return respones.Sealed, nil
}

func getDataFromSecret(config *vault.Config, token string, path string) (map[string]interface{}, error) {
	secret, err := GetSecret(config, token, path)
	if err != nil {
		return nil, err
	}
	data := secret.Data["data"].(map[string]interface{})
	if len(data) == 0 {
		return nil, errors.New("Data of secret with path: " + path + " is empty")
	}
	return data, nil
}

func GetResticConfig(config *vault.Config, token string, path string) (*ResticConfig, error) {
	data, err := getDataFromSecret(config, token, "restic/data/"+path)
	if err != nil {
		return nil, err
	}

	conf := ResticConfig{
		path:     data["path"].(string),
		password: data["pw"].(string),
	}
	return &conf, nil
}

func GetGocryptConfig(config *vault.Config, token string, path string) (*GocryptConfig, error) {
	data, err := getDataFromSecret(config, token, "gocrypt/data/"+path)
	if err != nil {
		return nil, err
	}

	conf := GocryptConfig{
		mountPoint: data["mount-path"].(string),
		path:       data["path"].(string),
		password:   data["pw"].(string),
	}
	return &conf, nil
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
