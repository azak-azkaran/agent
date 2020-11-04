package main

import (
	"errors"
	"log"

	vault "github.com/hashicorp/vault/api"
)

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
	log.Println("Getting Data from: ", path)
	secret, err := GetSecret(config, token, path)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
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

func Login(config *vault.Config, role_id string, secret_id string) (string, error) {
	client, err := vault.NewClient(config)
	if err != nil {
		return "", err
	}

	// to pass the password
	options := map[string]interface{}{
		"secret_id": secret_id,
		"role_id":   role_id,
	}

	// PUT call to get a token
	secret, err := client.Logical().Write("auth/approle/login", options)
	if err != nil {
		return "", err
	}

	token := secret.Auth.ClientToken
	return token, nil
}
