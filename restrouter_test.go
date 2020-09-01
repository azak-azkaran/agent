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
	log.Println("Agent rest server startet on: ", server.Addr)

	resp, err := http.Get(REST_TEST_PING)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr := strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":\"pong\"}")

	err = server.Shutdown(context.Background())
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
	log.Println("Agent rest server startet on: ", server.Addr)
	msg := TokenMessage{
		Token: "randomtoken",
	}
	reqBody, err := json.Marshal(msg)
	require.NoError(t, err)

	//check seal
	log.Println("address:", AgentConfiguration.VaultConfig.Address)
	resp, err := http.Get(REST_TEST_IS_SEALED)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr := strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":false}")

	//seal vault
	resp, err = http.Post(REST_TEST_SEAL,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr = strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":true}")

	// check seal
	resp, err = http.Get(REST_TEST_IS_SEALED)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr = strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":true}")

	// unseal vault
	resp, err = http.Post(REST_TEST_UNSEAL,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr = strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":false}")

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)
}

func TestRestPostBackup(t *testing.T) {
	fmt.Println("running: TestRestPostBackup")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	backupMsg := BackupMessage{
		Mode:  "backup",
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
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	backupMsg.Mode = "init"
	backupMsg.Test = false
	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)
	fmt.Println("Sending Body:", string(reqBody))

	resp, err = http.Post(REST_TEST_BACKUP,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	backupMsg.Mode = "exist"
	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)
	fmt.Println("Sending Body:", string(reqBody))

	resp, err = http.Post(REST_TEST_BACKUP,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	backupMsg.Mode = "backup"
	backupMsg.Run = false
	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)
	fmt.Println("Sending Body:", string(reqBody))

	resp, err = http.Post(REST_TEST_BACKUP,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	time.Sleep(1 * time.Millisecond)
	assert.Eventually(t, func() bool {
		v, ok := jobmap.Get(backupMsg.Mode)
		require.True(t, ok)
		require.NotNil(t, v)
		j := v.(Job)

		_, err := os.Stat(BACKUP_TEST_CONF_FILE)
		return j.Cmd.Process != nil && err == nil
	},
		time.Duration(25*time.Second), time.Duration(1*time.Second))

	v, _ := jobmap.Get(backupMsg.Mode)
	require.NotNil(t, v)
	cmd := v.(Job)

	fmt.Println(cmd.Stdout.String())

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = RemoveContents(BACKUP_TEST_FOLDER)
	assert.NoError(t, err)
	assert.NoFileExists(t, BACKUP_TEST_CONF_FILE)
}

func TestRestPostMount(t *testing.T) {
	fmt.Println("running: TestRestPostMount")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	mountMsg := MountMessage{
		Test:  true,
		Token: "randomtoken",
		Debug: true,
	}
	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)
	reqBody, err := json.Marshal(mountMsg)
	require.NoError(t, err)

	resp, err := http.Post(REST_TEST_MOUNT,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mountMsg.Test = false
	reqBody, err = json.Marshal(mountMsg)
	require.NoError(t, err)

	resp, err = http.Post(REST_TEST_MOUNT,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Eventually(t, func() bool {
		_, err := os.Stat(GOCRYPT_TEST_FILE)
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)
	require.FileExists(t, GOCRYPT_TEST_FILE)
	b, err := ioutil.ReadFile(GOCRYPT_TEST_FILE) // just pass the file name
	assert.NoError(t, err)
	assert.Equal(t, "testfile\n", string(b))

	time.Sleep(7 * time.Second)
	assert.NoFileExists(t, GOCRYPT_TEST_FILE)

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
	tokenMessage := TokenMessage{
		Token: "randomtoken",
	}

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	reqBody, err := json.Marshal(tokenMessage)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post(REST_TEST_TOKEN,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(REST_TEST_TOKEN)
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

	msg := VaultKeyMessage{
		Key:   "test1",
		Share: 1,
	}

	go fun()
	time.Sleep(1 * time.Millisecond)
	ok := CheckSealKey(AgentConfiguration.DB, 1)
	assert.False(t, ok)

	log.Println("Agent rest server startet on: ", server.Addr)

	reqBody, err := json.Marshal(msg)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post(REST_TEST_UNSEAL_KEY,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ok = CheckSealKey(AgentConfiguration.DB, 1)
	assert.True(t, ok)

	err = AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

}
func TestRestBindings(t *testing.T) {
	fmt.Println("Running: TestRestBindings")
	setupRestrouterTest(t)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	msg := DummyMessage{
		Message: "test",
	}

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	reqBody, err := json.Marshal(msg)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post(REST_TEST_TOKEN,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Post(REST_TEST_UNSEAL_KEY,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Post(REST_TEST_UNSEAL,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Post(REST_TEST_BACKUP,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Post(REST_TEST_MOUNT,
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}
