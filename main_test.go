package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	cqueue "github.com/enriquebris/goconcurrentqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")

	go fun()
	time.Sleep(1 * time.Millisecond)

	config, err := GetConfigFromVault(testconfig.token, testconfig.configpath, testconfig.config)
	require.NoError(t, err)
	assert.NotNil(t, config.Agent)
	assert.NotNil(t, config.Restic)
	assert.NotEmpty(t, config.Gocrypt)

	testconfig.configpath = "notExist"
	config, err = GetConfigFromVault(testconfig.token, testconfig.configpath, testconfig.config)
	assert.Error(t, err)
	assert.EqualError(t, err, ERROR_VAULT_NO_SECRET)
	assert.Nil(t, config)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
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

func TestMainStart(t *testing.T) {
	fmt.Println("running: TestMainStart")
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")

	go fun()
	assert.NotNil(t, AgentConfiguration.DB)

	time.Sleep(1 * time.Millisecond)
	Start("5s", false)

	tokenMessage := TokenMessage{
		Token: "randomtoken",
	}
	reqBody, err := json.Marshal(tokenMessage)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post("http://localhost:8081/token",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	Start("5s", false)

	time.Sleep(1 * time.Millisecond)
	assert.Eventually(t, func() bool {
		v, ok := jobmap.Get("backup")
		require.True(t, ok)
		require.NotNil(t, v)
		j := v.(Job)
		return j.Cmd.Process != nil
	},
		time.Duration(25*time.Second), time.Duration(1*time.Second))

	if ConcurrentQueue.GetLen() > 0 {
		_, err = ConcurrentQueue.Dequeue()
		assert.NoError(t, err)
	}
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = RemoveContents(BACKUP_TEST_FOLDER)
	assert.NoError(t, err)
	assert.NoFileExists(t, BACKUP_TEST_CONF_FILE)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)
}
