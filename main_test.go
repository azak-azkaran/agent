package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	cqueue "github.com/enriquebris/goconcurrentqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	MAIN_TEST_ADDRESS = "localhost:8081"
)

func TestMainAddJob(t *testing.T) {
	fmt.Println("running: TestMainAddJob")
	cmd := exec.Command("echo", "hallo")
	AddJob(cmd, "test")
	assert.NotNil(t, jobmap)
	assert.NotEmpty(t, jobmap)
}

func TestMainRunJobBackground(t *testing.T) {
	fmt.Println("running: TestMainRunJobBackground")
	cmd := exec.Command("echo", "hallo")

	err := RunJobBackground(cmd, "test")
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		v, ok := jobmap.Get("test")
		require.True(t, ok)
		job := v.(Job)
		return job.Cmd.Process != nil
	}, time.Duration(4*time.Second), time.Duration(1*time.Second))
	v, ok := jobmap.Get("test")
	require.True(t, ok)
	job := v.(Job)

	assert.True(t, ConcurrentQueue.GetLen() > 0)
	assert.Equal(t, "hallo\n", job.Stdout.String())
	assert.Equal(t, "", job.Stderr.String())
}

func TestMainRunJob(t *testing.T) {
	fmt.Println("running: TestMainRunJob")
	cmd := exec.Command("echo", "hallo")

	err := RunJob(cmd, "test")
	assert.NoError(t, err)

	v, ok := jobmap.Get("test")
	require.True(t, ok)
	job := v.(Job)
	assert.Equal(t, "hallo\n", job.Stdout.String())
	assert.Equal(t, "", job.Stderr.String())

	cmd = exec.Command("printenv")
	cmd.Env = []string{"TEST=hallo"}

	err = RunJob(cmd, "test")
	assert.NoError(t, err)

	v, ok = jobmap.Get("test")
	require.True(t, ok)
	job = v.(Job)
	assert.Equal(t, "TEST=hallo\n", job.Stdout.String())
	assert.Equal(t, "", job.Stderr.String())

}

func TestMainGetConfigFromVault(t *testing.T) {
	fmt.Println("running: TestMainGetConfigFromVault")
	testconfig := readConfig(t)

	config, err := GetConfigFromVault(testconfig.token, testconfig.configpath, testconfig.config)
	require.NoError(t, err)
	assert.NotNil(t, config.Agent)
	assert.NotNil(t, config.Restic)
	assert.NotEmpty(t, config.Gocrypt)
}

func TestMainInit(t *testing.T) {
	fmt.Println("running: TestMainInit")
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

func TestMainQueueJobStatus(t *testing.T) {
	fmt.Println("running: TestMainQueueJobStatus")
	ConcurrentQueue = cqueue.NewFIFO()

	cmd := exec.Command("echo", "hallo")

	err := RunJob(cmd, "test1")
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	assert.NotNil(t, jobmap)
	assert.NotEmpty(t, jobmap)
	v, ok := jobmap.Get("test1")

	require.True(t, ok)
	job := v.(Job)
	assert.NotNil(t, job.Cmd.Process)
}
