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
	err := Init(testconfig.config, os.Args[2:])
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

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestPostBackup(t *testing.T) {
	fmt.Println("running: TestPostBackup")
	setupRestrouterTest(t)
	server, fun := RunRestServer("localhost:8081")
	backupMsg := BackupMessage{
		Mode: "backup",
		Run:  false,
	}
	msg := TokenMessage{
		Token: "randomtoken",
	}

	go fun()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)

	reqBody, err := json.Marshal(msg)
	require.NoError(t, err)

	_, err = http.Post("http://localhost:8081/config",
		"application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	reqBody, err = json.Marshal(backupMsg)
	require.NoError(t, err)

	resp, err := http.Post("http://localhost:8081/backup",
		"application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEmpty(t, ConcurrentQueue)
}
