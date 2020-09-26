package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

var server *http.Server
var running bool = false
var sealStatus bool = false
var multipleKey bool = false

var Progress = 0
var Hostname string

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
		log.Println("Starting MOCK server at: ", server.Addr)
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
	r.GET("/v1/sys/seal-status", test_seal_status)
	r.PUT("/v1/sys/unseal", test_unseal)
	r.PUT("/v1/sys/seal", func(c *gin.Context) {
		log.Println("MOCK-Server: called seal")
		sealStatus = true
		c.JSON(http.StatusOK, nil)
	})
	r.GET("/v1/restic/data/resticpath", func(c *gin.Context) {
		log.Println("MOCK-Server: called resticpath")
		var msg vault.Secret
		data := make(map[string]interface{})
		secret := make(map[string]string)
		secret["path"] = "~/"
		secret["repo"] = VAULT_TEST_BACKUP_PATH
		secret["pw"] = VAULT_TEST_PASSWORD
		secret["exclude"] = VAULT_TEST_BACKUP_EXCLUDE_FILE
		secret["access_key"] = VAULT_TEST_BACKUP_ACCESS_KEY
		secret["secret_key"] = VAULT_TEST_BACKUP_SECRET_KEY
		data["data"] = secret
		msg.Data = data
		c.JSON(http.StatusOK, msg)
	})
	//r.GET("/v1/config/"+Hostname, config)
	//r.GET("/v1/config/configpath", config)
	r.GET("/v1/config/:name", func(c *gin.Context) {
		name := c.Param("name")

		if name == Hostname || name == "configpath" {
			test_config(c)
			return
		}
		log.Println("MOCK-Server: invalid config")
		var arr []string
		c.JSON(404, gin.H{"error": arr})
	})
	r.GET("/v1/gocrypt/data/random-config-path", test_gocrypt)
	r.GET("/v1/gocrypt/data/gocryptpath", test_gocrypt)
	return r
}

func test_gocrypt(c *gin.Context) {
	log.Println("MOCK-Server: called gocrypt")
	var msg vault.Secret
	data := make(map[string]interface{})
	secret := make(map[string]interface{})

	secret["path"] = VAULT_TEST_PATH
	secret["mount-path"] = VAULT_TEST_MOUNTPATH
	secret["pw"] = VAULT_TEST_PASSWORD
	secret["duration"] = "5s"
	secret["notempty"] = "true"
	data["data"] = secret
	msg.Data = data
	c.JSON(http.StatusOK, msg)
}

func test_config(c *gin.Context) {
	log.Println("MOCK-Server: called config")
	var msg vault.Secret
	data := make(map[string]interface{})

	pwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	data["restic"] = "resticpath"
	data["gocryptfs"] = VAULT_TEST_CONFIGPATH
	data["home"] = pwd
	msg.Data = data
	c.JSON(http.StatusOK, msg)
}

func test_seal_status(c *gin.Context) {
	log.Println("MOCK-Server: called seal-status")
	var msg vault.SealStatusResponse
	if multipleKey {
		msg.T = 3
		msg.N = 5
		msg.Progress = Progress
	}

	if msg.Progress >= msg.T {
		msg.Progress = 0
	}
	msg.Sealed = sealStatus
	c.JSON(http.StatusOK, msg)

}

func test_unseal(c *gin.Context) {
	log.Println("MOCK-Server: called unseal")
	var msg vault.SealStatusResponse

	if multipleKey {
		Progress = Progress + 1
		msg.T = 3
		msg.N = 5
		msg.Progress = Progress
		sealStatus = msg.Progress <= msg.T-1
	} else {
		sealStatus = false
	}

	if Progress >= msg.T {
		msg.Progress = 0
		Progress = 0
	}
	msg.Sealed = sealStatus

	c.JSON(http.StatusOK, msg)
}
