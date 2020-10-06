package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	token       string
	resticpath  string
	gocryptpath string
	configpath  string
	gitpath     string
	secret      string
	config      *vault.Config
	Duration    string
}

func readConfig(t *testing.T) TestConfig {

	require.FileExists(t, "./test/secret_mock.yml", "Token file not found")
	viper.SetConfigName("secret_mock.yml")
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
		gitpath:     viper.GetString("gitpath"),
		configpath:  viper.GetString("configpath"),
		secret:      viper.GetString("secret"),
		Duration:    viper.GetString("duration"),
	}

	conf.config = &vault.Config{
		Address: "http://localhost:2200",
	}
	StartServer(t, conf.config.Address)
	return conf
}

func TestConfigGetCreateConfigFullFromVault(t *testing.T) {
	fmt.Println("running: TestConfigGetCreateConfigFullFromVault")
	testconfig := readConfig(t)
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

	go fun()
	time.Sleep(1 * time.Millisecond)

	config, err := CreateConfigFullFromVault(testconfig.token, testconfig.configpath, testconfig.config)
	require.NoError(t, err)
	assert.NotNil(t, config.Agent)
	assert.NotNil(t, config.Restic)
	assert.NotEmpty(t, config.Gocrypt)

	testconfig.configpath = "notExist"
	config, err = CreateConfigFullFromVault(testconfig.token, testconfig.configpath, testconfig.config)
	assert.Error(t, err)
	assert.EqualError(t, err, ERROR_VAULT_NO_SECRET)
	assert.Nil(t, config)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestConfigGetGocryptConfig(t *testing.T) {
	fmt.Println("running: TestConfigGetGocryptConfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	conf, err := GetGocryptConfig(testconfig.config, testconfig.token, testconfig.gocryptpath)
	assert.NoError(t, err)
	assert.NotNil(t, conf.Path)
	assert.NotNil(t, conf.MountPoint)
	assert.NotNil(t, conf.Password)
	assert.True(t, conf.NotEmpty)
	assert.False(t, conf.AllowOther)
}

func TestConfigGetResticConfig(t *testing.T) {
	fmt.Println("running: TestConfigGetResticConfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	conf, err := GetResticConfig(testconfig.config, testconfig.token, testconfig.resticpath)
	assert.NoError(t, err)
	assert.NotNil(t, conf.Path)
	assert.NotNil(t, conf.Password)
}

func TestConfigGetAgentConfig(t *testing.T) {
	fmt.Println("running: TestConfigGetAgentConfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	conf, err := GetAgentConfig(testconfig.config, testconfig.token, testconfig.configpath)
	require.NoError(t, err)
	assert.NotNil(t, conf.Gocryptfs)
	assert.NotEmpty(t, conf.Gocryptfs)

	testconfig.configpath = "notExist"
	conf, err = GetAgentConfig(testconfig.config, testconfig.token, testconfig.configpath)
	assert.Error(t, err)
	assert.Nil(t, conf)
}

func TestConfigGetGitConfig(t *testing.T) {
	fmt.Println("running: TestConfigGetGitConfig")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	conf, err := GetGitConfig(testconfig.config, testconfig.token, testconfig.gitpath)
	require.NoError(t, err)
	require.NotNil(t, conf.Rep)
	require.NotNil(t, conf.PersonalToken)
	require.NotNil(t, conf.Directory)
	assert.Equal(t, "https://github.com/azak-azkaran/reverse-link", conf.Rep)
	assert.Equal(t, "", conf.PersonalToken)
	assert.Equal(t, "~/test/reverse", conf.Directory)
}
