package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

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
	server := RunRestServer("localhost:8081")

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

func TestGetIsSealed(t *testing.T) {
	fmt.Println("running: TestRunRestServer")
	setupRestrouterTest(t)
	server := RunRestServer("localhost:8081")

	resp, err := http.Get("http://localhost:8081/is_sealed")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	bodyStr := strings.TrimSpace(string(body))
	assert.Equal(t, bodyStr, "{\"message\":false}")

	//resp, err := http.Post("http://localhost:8081/seal")

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}
