package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DummyMessage struct {
	Message string
}

func setupRestrouterTest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	err := Init(testconfig.config, os.Args)
	require.NoError(t, err)

	if AgentConfiguration.DB == nil {
		AgentConfiguration.DB = InitDB("", "", true)
	}
}

func TestRestCreateRestHandler(t *testing.T) {
	fmt.Println("running: TestRestCreateRestHandler")
	setupRestrouterTest(t)
	server := &http.Server{
		Addr:    MAIN_TEST_ADDRESS,
		Handler: CreateRestHandler(),
	}
	go func() {
		err := server.ListenAndServe()
		assert.Equal(t, http.ErrServerClosed, err)
	}()

	err := server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)
}

func TestRestRunRestServer(t *testing.T) {
	fmt.Println("running: TestRestRunRestServer")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

	go fun()
	time.Sleep(1 * time.Millisecond)

	bodyStr := sendingGet(t, REST_TEST_PING, http.StatusOK)
	assert.Equal(t, bodyStr, "{\"message\":\"pong\"}")

	err := server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)
}

func TestRestHandleSeal(t *testing.T) {
	fmt.Println("running: TestRestRunRestServer")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

	go fun()
	time.Sleep(1 * time.Millisecond)
	msg := TokenMessage{
		Token: "randomtoken",
	}
	bodyStr := sendingGet(t, REST_TEST_IS_SEALED, http.StatusOK)
	assert.Equal(t, bodyStr, "{\"message\":false}")

	//seal vault
	bodyStr = sendingPost(t, REST_TEST_SEAL, http.StatusOK, msg)
	assert.Equal(t, bodyStr, "{\"message\":true}")

	// check seal
	bodyStr = sendingGet(t, REST_TEST_IS_SEALED, http.StatusOK)
	assert.Equal(t, bodyStr, "{\"message\":true}")

	// unseal vault
	bodyStr = sendingPost(t, REST_TEST_UNSEAL, http.StatusOK, msg)
	assert.Equal(t, bodyStr, "{\"message\":false}")

	err := server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)
}

func TestRestPostBackup(t *testing.T) {
	fmt.Println("running: TestRestPostBackup")
	clear()
	t.Cleanup(clear)
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

	go fun()
	time.Sleep(1 * time.Millisecond)

	msg := BackupMessage{
		Mode:        "backup",
		Test:        true,
		Run:         true,
		Debug:       true,
		PrintOutput: true,
		Token:       "randomtoken",
	}
	sendingPost(t, REST_TEST_BACKUP, http.StatusOK, msg)

	msg.Mode = "init"
	msg.Test = false
	sendingPost(t, REST_TEST_BACKUP, http.StatusOK, msg)

	msg.Mode = "exist"
	sendingPost(t, REST_TEST_BACKUP, http.StatusOK, msg)
	time.Sleep(1 * time.Millisecond)

	msg.Mode = "backup"
	msg.Run = false
	sendingPost(t, REST_TEST_BACKUP, http.StatusOK, msg)

	msg.Mode = "check"
	msg.Run = true
	sendingPost(t, REST_TEST_BACKUP, http.StatusInternalServerError, msg)

	msg.Mode = "unlock"
	sendingPost(t, REST_TEST_BACKUP, http.StatusOK, msg)

	msg.Mode = "list"
	sendingPost(t, REST_TEST_BACKUP, http.StatusOK, msg)

	assert.Eventually(t, func() bool {
		pwd, err := os.Getwd()
		require.NoError(t, err)

		test_conf := strings.ReplaceAll(BACKUP_TEST_CONF_FILE, HOME, pwd)

		_, err = os.Stat(test_conf)
		return err == nil
	},
		time.Duration(25*time.Second), time.Duration(1*time.Second))

	var v interface{}
	for value := range jobmap.IterBuffered() {
		if strings.Contains(value.Key, msg.Mode) {
			v = value.Val
			break
		}
	}

	require.NotNil(t, v)
	cmd := v.(Job)

	fmt.Println(cmd.Stdout.String())

	err := server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)
}

func TestRestPostMount(t *testing.T) {
	fmt.Println("running: TestRestPostMount")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	mountMsg := MountMessage{
		Test:        true,
		Token:       "randomtoken",
		Debug:       true,
		PrintOutput: true,
	}
	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)
	reqBody, err := json.Marshal(mountMsg)
	require.NoError(t, err)

	resp, err := http.Post(REST_TEST_MOUNT,
		"application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mountMsg.Test = false
	reqBody, err = json.Marshal(mountMsg)
	require.NoError(t, err)

	resp, err = http.Post(REST_TEST_MOUNT,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	pwd, err := os.Getwd()
	require.NoError(t, err)
	test_folder := strings.ReplaceAll(GOCRYPT_TEST_MOUNTPATH+GOCRYPT_TEST_FILE, HOME, pwd)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Eventually(t, func() bool {

		_, err = os.Stat(test_folder)
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)

	require.FileExists(t, test_folder)
	b, err := ioutil.ReadFile(test_folder) // just pass the file name
	assert.NoError(t, err)
	assert.Equal(t, "testfile\n", string(b))

	time.Sleep(7 * time.Second)
	assert.NoFileExists(t, test_folder)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}
func TestRestStatus(t *testing.T) {
	fmt.Println("running: TestRestStatus")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	backupMsg := BackupMessage{
		Mode:  "init",
		Run:   true,
		Debug: true,
		Token: "randomtoken",
	}

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	reqBody, err := json.Marshal(backupMsg)
	require.NoError(t, err)
	_, err = http.Post(REST_TEST_BACKUP,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	backupMsg.Mode = "check"
	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)
	_, err = http.Post(REST_TEST_BACKUP,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	resp, err := http.Get(REST_TEST_STATUS)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = RemoveContents(BACKUP_TEST_FOLDER)
	assert.NoError(t, err)
	assert.NoFileExists(t, BACKUP_TEST_CONF_FILE)
}

func TestRestGetLog(t *testing.T) {
	fmt.Println("running: TestRestGetLog")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	backupMsg := BackupMessage{
		Mode:  "blub",
		Test:  true,
		Run:   true,
		Debug: true,
		Token: "randomtoken",
	}

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	reqBody, err := json.Marshal(backupMsg)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post(REST_TEST_BACKUP,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	resp, err = http.Get(REST_TEST_LOG)
	assert.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestRestPostToken(t *testing.T) {
	fmt.Println("running: TestRestPostToken")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

	go fun()
	time.Sleep(1 * time.Millisecond)

	tokenMessage := TokenMessage{
		Token: "randomtoken",
	}
	sendingPost(t, REST_TEST_TOKEN, http.StatusOK, tokenMessage)

	resp, err := http.Get(REST_TEST_TOKEN)
	assert.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyString := string(bodyBytes)
	assert.Contains(t, bodyString, tokenMessage.Token)

	ok := CheckToken(AgentConfiguration.DB)
	assert.True(t, ok)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestRestPostUnsealKey(t *testing.T) {
	fmt.Println("running: TestRestPostUnsealKey")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

	go fun()
	time.Sleep(1 * time.Millisecond)
	ok := CheckSealKey(AgentConfiguration.DB, 1)
	assert.False(t, ok)

	msg := VaultKeyMessage{
		Key:   "test1",
		Share: 1,
	}
	sendingPost(t, REST_TEST_UNSEAL_KEY, http.StatusOK, msg)

	ok = CheckSealKey(AgentConfiguration.DB, 1)
	assert.True(t, ok)

	err := AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

}
func TestRestBindings(t *testing.T) {
	fmt.Println("Running: TestRestBindings")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	msg := DummyMessage{
		Message: "test",
	}
	sendingPost(t, REST_TEST_TOKEN, http.StatusBadRequest, msg)
	sendingPost(t, REST_TEST_UNSEAL_KEY, http.StatusBadRequest, msg)
	sendingPost(t, REST_TEST_UNSEAL, http.StatusBadRequest, msg)
	sendingPost(t, REST_TEST_BACKUP, http.StatusBadRequest, msg)
	sendingPost(t, REST_TEST_MOUNT, http.StatusBadRequest, msg)

	err := server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestRestPostGit(t *testing.T) {
	fmt.Println("running: TestRestPostGit")
	t.Cleanup(clear)
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	mountMsg := GitMessage{
		Mode:        "clone",
		Token:       "randomtoken",
		Debug:       true,
		PrintOutput: true,
		Run:         true,
	}
	sendingPost(t, REST_TEST_GIT, http.StatusOK, mountMsg)

	mountMsg.Mode = "pull"
	sendingPost(t, REST_TEST_GIT, http.StatusOK, mountMsg)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	test_folder := strings.ReplaceAll(GIT_TEST_FOLDER, HOME, pwd)

	assert.DirExists(t, test_folder)
}

func sendingPost(t *testing.T, endpoint string, statusCode int, msg interface{}) string {
	reqBody, err := json.Marshal(msg)
	require.NoError(t, err)
	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post(endpoint,
		"application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, statusCode, resp.StatusCode)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(bodyBytes)
}

func sendingGet(t *testing.T, endpoint string, statusCode int) string {
	resp, err := http.Get(endpoint)
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, statusCode, resp.StatusCode)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(bodyBytes)
}
