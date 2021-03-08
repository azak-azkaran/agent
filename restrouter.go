package main

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TokenMessage struct {
	Token string `json:"token" binding:"required"`
}

type VaultKeyMessage struct {
	Key   string `json:"key" binding:"required"`
	Share int    `json:"share" binding:"required"`
}

type BackupMessage struct {
	Mode        string `json:"mode" binding:"required"`
	Token       string `json:"token" binding:"required"`
	Run         bool   `json:"run"`
	Test        bool   `json:"test"`
	Debug       bool   `json:"debug"`
	PrintOutput bool   `json:"print"`
	DryRun      bool   `json:"dryrun"`
}

type MountMessage struct {
	Token       string `json:"token" binding:"required"`
	Run         bool   `json:"run"`
	Test        bool   `json:"test"`
	Debug       bool   `json:"debug"`
	PrintOutput bool   `json:"print"`
}

type GitMessage struct {
	Mode        string `json:"mode" binding:"required"`
	Token       string `json:"token" binding:"required"`
	Run         bool   `json:"run"`
	Debug       bool   `json:"debug"`
	PrintOutput bool   `json:"print"`
}

func handleError(job Job, err error, errMsg string, buffer bytes.Buffer) bool {
	if err != nil {
		m := errMsg + err.Error()
		if job.Stderr != nil {
			m = "\t" + job.Stderr.String()
		}
		Sugar.Error(m)
		buffer.WriteString(m)
		return false
	}
	return true
}

func returnErr(err error, source string, c *gin.Context) {
	Sugar.Error(source, err.Error())

	var code int
	if source == ERROR_CONFIG {
		code = http.StatusForbidden
	} else {
		code = http.StatusInternalServerError
	}
	c.JSON(code, gin.H{
		REST_JSON_MESSAGE: err.Error(),
	})
}

func postUnseal(c *gin.Context) {
	var msg TokenMessage
	if err := c.BindJSON(&msg); err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	sealed, err := DoUnseal(msg.Token)
	if  err != nil {
		returnErr(err, ERROR_UNSEAL, c)
	}
	c.JSON(http.StatusOK, gin.H{
		REST_JSON_MESSAGE: sealed,
	})
}

func postSeal(c *gin.Context) {
	var msg TokenMessage
	if err := c.BindJSON(&msg); err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	err := DoSeal(msg.Token)
	if err != nil{
		returnErr(err, ERROR_SEAL, c)
		return
	}

		c.JSON(http.StatusOK, gin.H{
		REST_JSON_MESSAGE: true,
	})
}

func getIsSealed(c *gin.Context) {
	b, err := IsSealed(AgentConfiguration.VaultConfig)
	if err != nil {
		returnErr(err, ERROR_ISSEALED, c)
		return
	}

	Sugar.Info(REST_VAULT_SEAL_MESSAGE, b)
	c.JSON(http.StatusOK, gin.H{
		REST_JSON_MESSAGE: b,
	})
}

func getLog(c *gin.Context) {
	Sugar.Info("Log: ", "not implemented")
	c.JSON(http.StatusOK, gin.H{
		REST_JSON_MESSAGE: "not implemented",
	})
}

func getStatus(c *gin.Context) {
	if jobmap == nil {
		returnErr(errors.New("ConcurrentMap not initialized"), ERROR_STATUS, c)
		return
	}
	var buffer bytes.Buffer
	for _, k := range jobmap.Keys() {
		v, ok := jobmap.Get(k)
		if ok {
			cmd := v.(*Job)
			buffer.WriteString("Job: " + k + " Status: " + cmd.Cmd.ProcessState.String())
		} else {
			buffer.WriteString("Job: " + k + " Error while retrieving")
		}
	}
	if jobmap.IsEmpty() {
		buffer.WriteString("No Job started")
	}

	Sugar.Info("Get Status: ", buffer.String())
	c.JSON(http.StatusOK, gin.H{
		REST_JSON_MESSAGE: buffer.String(),
	})
}

func postMount(c *gin.Context) {
	var msg MountMessage
	if err := c.BindJSON(&msg); err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	str, err := DoMount(msg.Token, msg.Debug, msg.PrintOutput, msg.Test, msg.Run)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			REST_JSON_MESSAGE: str,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		REST_JSON_MESSAGE: str,
	})
}

func postBackup(c *gin.Context) {
	var msg BackupMessage
	if err := c.BindJSON(&msg); err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	if err := DoBackup(msg.Token, msg.Mode, msg.PrintOutput, msg.Debug, msg.Test, msg.Run); err != nil {
		returnErr(err, ERROR_RUNBACKUP, c)
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func postUnsealKey(c *gin.Context) {
	var msg VaultKeyMessage
	if err := c.BindJSON(&msg); err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	ok, err := PutSealKey(AgentConfiguration.DB, msg.Key, msg.Share)

	if err != nil {
		returnErr(err, ERROR_PUT_SEAL_KEY, c)
		return
	}
	if ok {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{})
	}
}

func postGit(c *gin.Context) {
	var msg GitMessage
	if err := c.BindJSON(&msg); err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	str, err := DoGit(msg.Token, msg.Mode, msg.Run, msg.PrintOutput)
	if err != nil {
		returnErr(err, ERROR_GIT, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		REST_JSON_MESSAGE: str,
	})
}

func RunRestServer(address string) (*http.Server, func()) {
	server := &http.Server{
		Addr:    address,
		Handler: CreateRestHandler(),
	}
	fun := func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			Sugar.Info("Agent server closed happily...")
		} else if err != nil {
			Sugar.Error("Agent server closed horribly...\n", err)
		}
	}
	return server, fun
}

func CreateRestHandler() http.Handler {
	r := gin.New()
	r.Use(gin.LoggerWithConfig(*logconfig))
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			REST_JSON_MESSAGE: "pong",
		})
	})

	r.POST("/unsealkey", postUnsealKey)
	r.POST("/unseal", postUnseal)
	r.POST("/seal", postSeal)
	r.POST("/mount", postMount)
	r.POST("/backup", postBackup)
	r.POST("/git", postGit)
	r.GET("/is_sealed", getIsSealed)
	r.GET("/status", getStatus)
	return r
}
