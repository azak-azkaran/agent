package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os/exec"
	"testing"
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
