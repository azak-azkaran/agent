package main

import (
	"context"
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
var forbidden bool = false

var Progress = 0
var Hostname string

func StartServer(t *testing.T, address string) {
	if running {
		Sugar.Info("Server already running")
		return
	}
	gin.SetMode(gin.TestMode)
	server = &http.Server{
		Addr:    strings.TrimPrefix(address, "http://"),
		Handler: createHandler(),
	}
	go func() {
		running = true
		Sugar.Info("Starting MOCK server at: ", server.Addr)
		err := server.ListenAndServe()
		require.Equal(t, http.ErrServerClosed, err)
		running = false
	}()
	time.Sleep(1 * time.Millisecond)
	t.Cleanup(StopServer)
}

func StopServer() {
	Sugar.Info("stopping server")
	if running {
		err := server.Shutdown(context.Background())
		if err != nil {
			Sugar.Fatal(err)
		}
	}
	time.Sleep(1 * time.Millisecond)
}

func createHandler() http.Handler {
	r := gin.Default()
	r.GET("/v1/sys/seal-status", test_seal_status)
	r.PUT("/v1/sys/unseal", test_unseal)
	r.PUT("/v1/sys/seal", func(c *gin.Context) {
		Sugar.Info("MOCK-Server: called seal")
		sealStatus = true
		c.JSON(http.StatusOK, nil)
	})
	r.GET("/v1/restic/data/resticpath", test_restic)
	r.GET("/v1/restic/data/forbidden", test_forbidden)
	r.GET("/v1/config/:name", func(c *gin.Context) {
		name := c.Param("name")

		if name == Hostname || name == "configpath" {
			test_config(c)
			return
		}
		Sugar.Info("MOCK-Server: invalid config")
		var arr []string
		c.JSON(404, gin.H{"error": arr})
	})
	r.GET("/v1/gocrypt/data/random-config-path", test_gocrypt)
	r.GET("/v1/gocrypt/data/gocryptpath", test_gocrypt)
	r.GET("/v1/git/data/gitpath", test_git)
	r.GET("/v1/git/data/vimrc", test_vimrc)
	r.PUT("/v1/auth/approle/login", test_login)
	return r
}

func test_gocrypt(c *gin.Context) {
	Sugar.Info("MOCK-Server: called gocrypt")
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
	Sugar.Info("MOCK-Server: called config")
	var msg vault.Secret
	data := make(map[string]interface{})

	pwd, err := os.Getwd()
	if err != nil {
		Sugar.Error(err)
	}

	if forbidden {
		data["restic"] = "forbidden"
	} else {
		data["restic"] = "resticpath"
	}
	data["gocryptfs"] = VAULT_TEST_CONFIGPATH
	data["git"] = "gitpath,vimrc"
	data["home"] = pwd
	msg.Data = data
	c.JSON(http.StatusOK, msg)
}

func test_vimrc(c *gin.Context) {
	Sugar.Info("MOCK-Server: called git")
	var msg vault.Secret
	data := make(map[string]interface{})

	data["repo"] = GIT_TEST_REPO_VIMRC
	data["dir"] = GIT_TEST_FOLDER_VIMRC
	data["personal_token"] = ""
	msg.Data = data
	c.JSON(http.StatusOK, msg)
}

func test_git(c *gin.Context) {
	Sugar.Info("MOCK-Server: called git")
	var msg vault.Secret
	data := make(map[string]interface{})

	data["repo"] = GIT_TEST_REPO
	data["dir"] = GIT_TEST_FOLDER
	data["personal_token"] = ""
	msg.Data = data
	c.JSON(http.StatusOK, msg)
}

func test_seal_status(c *gin.Context) {
	Sugar.Info("MOCK-Server: called seal-status")
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
	Sugar.Info("MOCK-Server: called unseal")
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

func test_forbidden(c *gin.Context) {
	Sugar.Info("MOCK-Server: called forbidden")
	var msg vault.Secret
	data := make(map[string]interface{})
	secret := make(map[string]string)
	data["data"] = secret
	msg.Data = data
	c.JSON(http.StatusForbidden, msg)
}

func test_restic(c *gin.Context) {
	Sugar.Info("MOCK-Server: called resticpath")
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
}

func test_login(c *gin.Context) {
	Sugar.Info("MOCK-Server: called login")
	msg := "{\"request_id\":\"requestid\",\"lease_id\":\"\",\"renewable\":false,\"lease_duration\":0,\"data\":null,\"wrap_info\":null,\"warnings\":null,\"auth\":{\"client_token\":\"" + VAULT_TEST_TOKEN + "\",\"accessor\":\"accessorid\",\"policies\":[\"default\",\"secret access\"],\"token_policies\":[\"default\",\"secret access\"],\"metadata\":{\"role_name\":\"agent\"},\"lease_duration\":3600,\"renewable\":true,\"entity_id\":\"entity_id\",\"token_type\":\"service\",\"orphan\":true}}"
	c.String(http.StatusOK, msg)
}
