package main

import (
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	PASSWORD  = "superrandompassword"
	TOKEN     = "superrandompasswordtoken"
	PATH      = "a-random-path"
	MOUNTPATH = "another-random-path"
)

var runMock bool = os.Getenv("RUN_MOCK") == "true"

type TestConfig struct {
	token       string
	resticpath  string
	gocryptpath string
	secret      string
	config      *vault.Config
}

func readConfig(t *testing.T) TestConfig {
	if runMock {
		require.FileExists(t, "./test/secret_mock.yml", "Token file not found")
		viper.SetConfigName("secret_mock.yml")
	} else {
		require.FileExists(t, "./test/secret.yml", "Token file not found")
		viper.SetConfigName("secret.yml")
	}
	viper.SetConfigType("yml")
	viper.AddConfigPath("test")
	require.NoError(t, viper.ReadInConfig())

	conf := TestConfig{
		token:       viper.GetString("token"),
		resticpath:  viper.GetString("resticpath"),
		gocryptpath: viper.GetString("gocryptpath"),
		secret:      viper.GetString("secret"),
	}

	if runMock {
		conf.config = &vault.Config{
			Address: "http://localhost:2200",
		}
		StartServer(t, conf.config.Address)
	} else {
		conf.config = &vault.Config{
			Address: "http://localhost:8200",
		}
	}
	return conf
}

func TestIsSealed(t *testing.T) {
	fmt.Println("running: TestGetStatus")
	testconfig := readConfig(t)

	resp, err := IsSealed(testconfig.config)
	assert.NoError(t, err)
	assert.False(t, resp, "Vault is not unsealed")
}

func TestGetSecret(t *testing.T) {
	fmt.Println("running: TestGetSecret")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, "Vault is not unsealed")

	resp, err := GetSecret(testconfig.config, testconfig.token, "restic/data/"+testconfig.resticpath)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Data["data"])
}

func TestGetGocryptConfig(t *testing.T) {
	fmt.Println("running: TestGetGocryptconfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, "Vault is not unsealed")

	conf, err := GetGocryptConfig(testconfig.config, testconfig.token, testconfig.gocryptpath)
	assert.NoError(t, err)
	assert.NotNil(t, conf.path)
	assert.NotNil(t, conf.mountPoint)
	assert.NotNil(t, conf.password)
}

func TestGetResticConfig(t *testing.T) {
	fmt.Println("running: TestGetResticConfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, "Vault is not unsealed")

	conf, err := GetResticConfig(testconfig.config, testconfig.token, testconfig.resticpath)
	assert.NoError(t, err)
	assert.NotNil(t, conf.path)
	assert.NotNil(t, conf.password)
}

func TestUnseal(t *testing.T) {
	fmt.Println("running: TestUnseal")
	if !runMock {
		fmt.Println("Test can only be run in Mock mode")
		return
	}
	testconfig := readConfig(t)
	err := Seal(testconfig.config, TOKEN)
	require.NoError(t, err)

	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.True(t, seal)

	_, err = Unseal(testconfig.config, testconfig.secret)
	require.NoError(t, err)

	seal, err = IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal)
}
