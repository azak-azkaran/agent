package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCreateRestHandler(t *testing.T) {
	fmt.Println("running: TestCreateRestHandler")
	gin.SetMode(gin.TestMode)
	endServer := &http.Server{
		Addr:    "localhost:8081",
		Handler: CreateRestHandler(),
	}
	go func() {
		err := endServer.ListenAndServe()
		assert.Equal(t, http.ErrServerClosed, err)
	}()

	err := endServer.Shutdown(context.Background())
	assert.NoError(t, err)
}
