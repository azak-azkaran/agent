package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	MAIN_TEST_ADDRESS = "localhost:8081"
)

func TestRunCmd(t *testing.T) {
	fmt.Println("running: TestRunCmd")
	cmd := exec.Command("echo", "hallo")
	out, err := RunJob(cmd)
	assert.NoError(t, err)
	assert.Equal(t, "hallo\n", out)
}

func TestGetConfigFromVault(t *testing.T) {
	fmt.Println("running: TestGetConfig")
	testconfig := readConfig(t)

	config, err := GetConfigFromVault(testconfig.token, testconfig.configpath, testconfig.config)
	require.NoError(t, err)
	assert.NotNil(t, config.Agent)
	assert.NotNil(t, config.Restic)
	assert.NotEmpty(t, config.Gocrypt)
}

func TestInit(t *testing.T) {
	fmt.Println("running: TestInitWithEnvironment")
	testconfig := readConfig(t)
	hostname, err := os.Hostname()
	require.NoError(t, err)

	// Test with flags
	os.Args = append(os.Args, "--address="+MAIN_TEST_ADDRESS)
	err = Init(testconfig.config, os.Args[2:])
	require.NoError(t, err)
	assert.Equal(t, AgentConfiguration.Hostname, hostname)
	assert.Equal(t, AgentConfiguration.Address, MAIN_TEST_ADDRESS)
	length := len(os.Args)

	os.Args = os.Args[:length-1]

	// Test with Environment variables
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	err = Init(testconfig.config, os.Args[2:])
	require.NoError(t, err)
	assert.Equal(t, AgentConfiguration.Hostname, hostname)
	assert.Equal(t, AgentConfiguration.Address, MAIN_TEST_ADDRESS)
	assert.NotNil(t, ConcurrentQueue)
}
