package main

import (
	"context"
	"github.com/gin-gonic/gin"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

var server *http.Server
var running bool = false
var sealStatus bool = false

func StartServer(t *testing.T, address string) {
	if running {
		log.Println("Server already running")
		return
	}
	gin.SetMode(gin.TestMode)
	server = &http.Server{
		Addr:    strings.TrimPrefix(address, "http://"),
		Handler: createHandler(),
	}
	go func() {
		running = true
		log.Println("Starting server")
		err := server.ListenAndServe()
		require.Equal(t, http.ErrServerClosed, err)
		running = false
	}()
	time.Sleep(1 * time.Millisecond)
	t.Cleanup(StopServer)
}

func StopServer() {
	log.Println("stopping server")
	if running {
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(1 * time.Millisecond)
}

func createHandler() http.Handler {
	r := gin.Default()
	r.GET("/v1/sys/seal-status", func(c *gin.Context) {
		var msg vault.SealStatusResponse
		msg.Sealed = sealStatus
		c.JSON(http.StatusOK, msg)
	})
	r.PUT("/v1/sys/unseal", func(c *gin.Context) {
		var msg vault.SealStatusResponse
		sealStatus = false
		msg.Sealed = sealStatus
		c.JSON(http.StatusOK, msg)
	})
	r.PUT("/v1/sys/seal", func(c *gin.Context) {
		sealStatus = true
		c.JSON(http.StatusOK, nil)
	})
	r.GET("/v1/restic/data/resticpath", func(c *gin.Context) {
		var msg vault.Secret
		data := make(map[string]interface{})
		secret := make(map[string]string)
		secret["path"] = PATH
		secret["pw"] = PASSWORD
		data["data"] = secret
		msg.Data = data
		c.JSON(http.StatusOK, msg)
	})
	r.GET("/v1/gocrypt/data/gocryptpath", func(c *gin.Context) {
		var msg vault.Secret
		data := make(map[string]interface{})
		secret := make(map[string]string)
		secret["path"] = PATH
		secret["mount-path"] = MOUNTPATH
		secret["pw"] = PASSWORD
		data["data"] = secret
		msg.Data = data
		c.JSON(http.StatusOK, msg)
	})
	return r
}
