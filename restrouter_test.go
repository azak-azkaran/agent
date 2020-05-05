package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestCreateRestHandler(t *testing.T) {
	fmt.Println("running: TestCreateRestHandler")
	gin.SetMode(gin.TestMode)
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
	gin.SetMode(gin.TestMode)
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
