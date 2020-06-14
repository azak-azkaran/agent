package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"testing"
	"time"

	cqueue "github.com/enriquebris/goconcurrentqueue"
	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map"
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
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

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

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestMainInit(t *testing.T) {
	fmt.Println("running: TestMainInit")
	testconfig := readConfig(t)
	hostname, err := os.Hostname()
	require.NoError(t, err)

	// Test with flags
	var args []string
	args = append(args, "--address="+MAIN_TEST_ADDRESS)
	args = append(args, "--pathdb="+MAIN_TEST_PATHDB)
	args = append(args, "--duration="+MAIN_TEST_DURATION)
	args = append(args, "--mount-duration="+MAIN_TEST_MOUNT_DURATION)
	args = append(args, "--mount-allow="+MAIN_TEST_MOUNT_ALLOW)

	err = Init(testconfig.config, args)
	require.NoError(t, err)
	assert.Equal(t, AgentConfiguration.Hostname, hostname)
	assert.Equal(t, AgentConfiguration.Address, MAIN_TEST_ADDRESS)
	assert.Equal(t, AgentConfiguration.PathDB, MAIN_TEST_PATHDB)
	assert.Equal(t, AgentConfiguration.MountAllow, false)
	assert.Equal(t, AgentConfiguration.MountDuration, MAIN_TEST_MOUNT_DURATION)

	dur, err := time.ParseDuration("1h30m")
	assert.NoError(t, err)

	assert.Equal(t, AgentConfiguration.TimeBetweenStart, dur)
	length := len(os.Args)

	os.Args = os.Args[:length-1]

	// Test with Environment variables
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_PATHDB", MAIN_TEST_PATHDB)
	os.Setenv("AGENT_DURATION", MAIN_TEST_DURATION)
	os.Setenv("AGNET_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGNET_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)

	err = Init(testconfig.config, args)
	require.NoError(t, err)
	assert.Equal(t, AgentConfiguration.Hostname, hostname)
	assert.Equal(t, AgentConfiguration.Address, MAIN_TEST_ADDRESS)
	assert.Equal(t, AgentConfiguration.PathDB, MAIN_TEST_PATHDB)
	assert.Equal(t, AgentConfiguration.TimeBetweenStart, dur)
	assert.Equal(t, AgentConfiguration.MountAllow, false)
	assert.Equal(t, AgentConfiguration.MountDuration, MAIN_TEST_MOUNT_DURATION)
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

	ConcurrentQueue = cqueue.NewFIFO()
	jobmap = cmap.New()
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	os.Setenv("AGNET_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGNET_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)
	err := Init(testconfig.config, os.Args)
	require.NoError(t, err)

	_, err = Unseal(testconfig.config, testconfig.secret)
	require.NoError(t, err)

	AgentConfiguration.DB = InitDB("", "", true)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	go fun()

	time.Sleep(1 * time.Millisecond)
	go Start("5s", false)
	time.Sleep(1 * time.Millisecond)

	tokenMessage := TokenMessage{
		Token: "randomtoken",
	}
	reqBody, err := json.Marshal(tokenMessage)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post(REST_TEST_TOKEN,
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

	AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)
}

func TestMainMain(t *testing.T) {
	fmt.Println("running: TestMainMain")
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	os.Setenv("AGENT_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGENT_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)
	multipleKey = true
	sealStatus = true
	Progress = 0

	go main()
	time.Sleep(1 * time.Second)
	AgentConfiguration.VaultConfig = testconfig.config

	resp, err := http.Get(REST_TEST_PING)
	require.NoError(t, err)
	defer resp.Body.Close()

	for i := 1; i < 6; i++ {
		msg := VaultKeyMessage{
			Key:   "test" + strconv.Itoa(i),
			Share: i,
		}
		reqBody, err := json.Marshal(msg)
		require.NoError(t, err)

		fmt.Println("Sending Body:", string(reqBody))
		_, err = http.Post(REST_TEST_UNSEAL_KEY,
			"application/json", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
	}

	tokenMessage := TokenMessage{
		Token: "randomtoken",
	}
	reqBody, err := json.Marshal(tokenMessage)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	_, err = http.Post(REST_TEST_TOKEN,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	token, err := GetToken(AgentConfiguration.DB)
	assert.NoError(t, err)
	assert.Equal(t, "randomtoken", token)

	time.Sleep(5 * time.Second)
	assert.Eventually(t, checkContents, 20*time.Second, 1*time.Second)

	stopChan <- syscall.SIGINT
	time.Sleep(5 * time.Second)

	is, err := IsEmpty("./test/tmp-mount")
	assert.NoError(t, err)
	assert.True(t, is)

	assert.False(t, sealStatus)
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = RemoveContents(BACKUP_TEST_FOLDER)
	assert.NoError(t, err)
	assert.NoFileExists(t, BACKUP_TEST_CONF_FILE)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)
	multipleKey = false
}

func checkContents() bool {
	ok, err := IsEmpty("./test/tmp-mount")
	if err != nil {
		return false
	}

	if ok {
		return false
	}

	ok, err = IsEmpty(BACKUP_TEST_FOLDER)
	if err != nil {
		return false
	}
	if ok {
		return false
	}

	return true
}
