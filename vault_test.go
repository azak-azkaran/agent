package main

import (
	"fmt"
	"os"
	"testing"

	cqueue "github.com/enriquebris/goconcurrentqueue"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var runMock bool = os.Getenv("RUN_MOCK") == "true"

type TestConfig struct {
	token       string
	resticpath  string
	gocryptpath string
	configpath  string
	secret      string
	config      *vault.Config
}

func readConfig(t *testing.T) TestConfig {
	ConcurrentQueue = cqueue.NewFIFO()
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

	name, err := os.Hostname()
	require.NoError(t, err)
	Hostname = name
	conf := TestConfig{
		token:       viper.GetString("token"),
		resticpath:  viper.GetString("resticpath"),
		gocryptpath: viper.GetString("gocryptpath"),
		configpath:  viper.GetString("configpath"),
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

func TestVaultIsSealed(t *testing.T) {
	fmt.Println("running: TestVaultGetStatus")
	testconfig := readConfig(t)

	resp, err := IsSealed(testconfig.config)
	assert.NoError(t, err)
	assert.False(t, resp, ERROR_VAULT_SEALED)
}

func TestVaultGetSecret(t *testing.T) {
	fmt.Println("running: TestVaultGetSecret")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	resp, err := GetSecret(testconfig.config, testconfig.token, "restic/data/"+testconfig.resticpath)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Data["data"])
}

func TestVaultGetGocryptConfig(t *testing.T) {
	fmt.Println("running: TestVaultGetGocryptconfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	conf, err := GetGocryptConfig(testconfig.config, testconfig.token, testconfig.gocryptpath)
	assert.NoError(t, err)
	assert.NotNil(t, conf.Path)
	assert.NotNil(t, conf.MountPoint)
	assert.NotNil(t, conf.Password)
}

func TestVaultGetResticConfig(t *testing.T) {
	fmt.Println("running: TestVaultGetResticConfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	conf, err := GetResticConfig(testconfig.config, testconfig.token, testconfig.resticpath)
	assert.NoError(t, err)
	assert.NotNil(t, conf.Path)
	assert.NotNil(t, conf.Password)
}

func TestVaultGetAgentConfig(t *testing.T) {
	fmt.Println("running: TestVaultGetAgentConfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	conf, err := GetAgentConfig(testconfig.config, testconfig.token, testconfig.configpath)
	require.NoError(t, err)
	assert.NotNil(t, conf.Gocryptfs)
	assert.NotEmpty(t, conf.Gocryptfs)
}

func TestVaultUnseal(t *testing.T) {
	fmt.Println("running: TestVaultUnseal")
	if !runMock {
		fmt.Println("Test can only be run in Mock mode")
		return
	}
	testconfig := readConfig(t)
	err := Seal(testconfig.config, VAULT_TOKEN)
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

func TestVaultCheckMap(t *testing.T) {
	fmt.Println("running: TestVaultCheckMap")

	list := []string{"test"}
	data := make(map[string]interface{})
	data["test"] = "test"
	err := CheckMap(list, data)
	assert.NoError(t, err)

	list = append(list, "test2")
	err = CheckMap(list, data)
	assert.Error(t, err)
}
