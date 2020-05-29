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

func setupRestrouterTest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", testconfig.config.Address)
	err := Init(testconfig.config, os.Args)
	require.NoError(t, err)
}

func TestCreateRestHandler(t *testing.T) {
	fmt.Println("running: TestCreateRestHandler")
	setupRestrouterTest(t)
	server := &http.Server{
		Addr:    "localhost:8081",
		Handler: CreateRestHandler(),
	}
	go func() {
		err := server.ListenAndServe()
		assert.Equal(t, http.ErrServerClosed, err)
	}()

	if ConcurrentQueue.GetLen() > 0 {
		_, err := ConcurrentQueue.Dequeue()
		assert.NoError(t, err)
	}
	err := server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestRunRestServer(t *testing.T) {
	fmt.Println("running: TestRunRestServer")
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	resp, err := http.Get("http://localhost:8081/ping")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr := strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":\"pong\"}")

	if ConcurrentQueue.GetLen() > 0 {
		_, err = ConcurrentQueue.Dequeue()
		assert.NoError(t, err)
	}
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestHandleSeal(t *testing.T) {
	fmt.Println("running: TestRunRestServer")
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)
	msg := TokenMessage{
		Token: "randomtoken",
	}
	reqBody, err := json.Marshal(msg)
	require.NoError(t, err)

	//check seal
	resp, err := http.Get("http://localhost:8081/is_sealed")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr := strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":false}")

	//seal vault
	resp, err = http.Post("http://localhost:8081/seal",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr = strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":true}")

	// check seal
	resp, err = http.Get("http://localhost:8081/is_sealed")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr = strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":true}")

	// unseal vault
	resp, err = http.Post("http://localhost:8081/unseal",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr = strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":false}")

	if ConcurrentQueue.GetLen() > 0 {
		_, err = ConcurrentQueue.Dequeue()
		assert.NoError(t, err)
	}
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestPostBackup(t *testing.T) {
	fmt.Println("running: TestPostBackup")
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")
	backupMsg := BackupMessage{
		Mode:  "backup",
		Test:  true,
		Run:   true,
		Token: "randomtoken",
	}

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	reqBody, err := json.Marshal(backupMsg)
	require.NoError(t, err)

	fmt.Println("Sending Body:", string(reqBody))
	resp, err := http.Post("http://localhost:8081/backup",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, ConcurrentQueue)

	backupMsg.Mode = "init"
	backupMsg.Test = false
	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)
	fmt.Println("Sending Body:", string(reqBody))

	resp, err = http.Post("http://localhost:8081/backup",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	backupMsg.Mode = "exist"
	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)
	fmt.Println("Sending Body:", string(reqBody))

	resp, err = http.Post("http://localhost:8081/backup",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	backupMsg.Mode = "backup"
	backupMsg.Run = false
	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)
	fmt.Println("Sending Body:", string(reqBody))

	resp, err = http.Post("http://localhost:8081/backup",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	time.Sleep(1 * time.Millisecond)
	assert.Eventually(t, func() bool {
		v, ok := jobmap.Get(backupMsg.Mode)
		require.True(t, ok)
		require.NotNil(t, v)
		j := v.(Job)
		return j.Cmd.Process != nil
	},
		time.Duration(25*time.Second), time.Duration(1*time.Second))
	//assert.EqualValues(t, ConcurrentQueue.GetLen(), 3)

	v, _ := jobmap.Get(backupMsg.Mode)
	require.NotNil(t, v)
	cmd := v.(Job)

	fmt.Println(cmd.Stdout.String())

	if ConcurrentQueue.GetLen() > 0 {
		_, err = ConcurrentQueue.Dequeue()
		assert.NoError(t, err)
	}
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = RemoveContents(BACKUP_FOLDER)
	assert.NoError(t, err)
	assert.NoFileExists(t, BACKUP_CONF_FILE)
}

func TestPostMount(t *testing.T) {
	fmt.Println("running: TestPostMount")
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")
	mountMsg := MountMessage{
		Test:  true,
		Token: "randomtoken",
	}
	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)
	reqBody, err := json.Marshal(mountMsg)
	require.NoError(t, err)

	resp, err := http.Post("http://localhost:8081/mount",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	if ConcurrentQueue.GetLen() > 0 {
		v, err := ConcurrentQueue.Dequeue()
		assert.NoError(t, err)
		fmt.Println(v)
	}
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}
func TestStatus(t *testing.T) {
	fmt.Println("running: TestStatus")
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	err := ConcurrentQueue.Enqueue("test")
	assert.NoError(t, err)

	resp, err := http.Get("http://localhost:8081/status")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, ConcurrentQueue)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}
