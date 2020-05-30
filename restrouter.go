package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

const (
	ERROR_MODE      = "Backup Mode:"
	ERROR_STATUS    = "GetStatus:"
	ERROR_LOG       = "GetLogs:"
	ERROR_ISSEALED  = "IsSealed:"
	ERROR_UNSEAL    = "Unseal:"
	ERROR_SEAL      = "Seal:"
	ERROR_RUNBACKUP = "RunBackupJob:"
	ERROR_ENQUEUE   = "Enqueue:"
	ERROR_CONFIG    = "GetConfigFromVault:"
	ERROR_BINDING   = "BindJSON:"
	ERROR_PREFIX    = "ERROR: "
	JSON_MESSAGE    = "message"
)

type TokenMessage struct {
	Token string `json:"token" binding:"required"`
}

type BackupMessage struct {
	Mode  string `json:"mode" binding:"required"`
	Token string `json:"token" binding:"required"`
	Run   bool   `json:"run"`
	Test  bool   `json:"test"`
	Debug bool   `json:"debug"`
}

type MountMessage struct {
	Token string `json:"token" binding:"required"`
	Run   bool   `json:"run"`
	Test  bool   `json:"test"`
	Debug bool   `json:"debug"`
}

func HandleBackup(cmd *exec.Cmd, mode string, function func(*exec.Cmd, string) error, c *gin.Context) {
	if err := function(cmd, mode); err != nil {
		log.Println(ERROR_PREFIX + err.Error())
		enqueue(err.Error(), c)
		c.JSON(http.StatusInternalServerError, gin.H{
			JSON_MESSAGE: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func HandleMountFolders(config []GocryptConfig, function func(*exec.Cmd, string) error, c *gin.Context, buffer bytes.Buffer) {
	out, ok := MountFolders(config, function)
	if !ok {
		for _, err := range out {
			log.Println(ERROR_PREFIX + err.Error())
			enqueue(err.Error(), c)
			buffer.WriteString(err.Error())
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			JSON_MESSAGE: buffer.String(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: buffer.String(),
	})
}

func enqueue(v interface{}, c *gin.Context) {
	err := ConcurrentQueue.Enqueue(v)
	if err != nil {
		log.Println(ERROR_PREFIX+ERROR_ENQUEUE, err.Error())
	}
}

func returnErr(err error, source string, c *gin.Context) {
	log.Println(ERROR_PREFIX+source, err.Error())
	c.JSON(http.StatusInternalServerError, gin.H{
		JSON_MESSAGE: err.Error(),
	})
}

func postUnseal(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(&msg)
	if err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	resp, err := Unseal(AgentConfiguration.VaultConfig, msg.Token)
	if err != nil {
		returnErr(err, ERROR_UNSEAL, c)
		return
	}

	log.Println("INFO: Vault seal is: ", resp.Sealed)
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: resp.Sealed,
	})
}

func postSeal(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(&msg)
	if err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}
	AgentConfiguration.Token = msg.Token
	err = Seal(AgentConfiguration.VaultConfig, AgentConfiguration.Token)
	if err != nil {
		returnErr(err, ERROR_SEAL, c)
		return
	}

	AgentConfiguration.Token = msg.Token
	log.Println("INFO: Vault Sealed")
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: true,
	})
}

func getIsSealed(c *gin.Context) {
	b, err := IsSealed(AgentConfiguration.VaultConfig)
	if err != nil {
		returnErr(err, ERROR_ISSEALED, c)
		return
	}

	log.Println("INFO: Vault seal is: ", b)
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: b,
	})
}

func getLog(c *gin.Context) {
	totalElements := ConcurrentQueue.GetLen()

	var buffer bytes.Buffer
	for i := 0; i < totalElements; i++ {
		m, err := ConcurrentQueue.Get(i)
		if err != nil {
			returnErr(errors.New(buffer.String()+"\n"+err.Error()), ERROR_LOG, c)
			return
		}
		buffer.WriteString(m.(string))
	}
	log.Println("INFO: Log: ", buffer.String())
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: buffer.String(),
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
			cmd := v.(*exec.Cmd)
			buffer.WriteString("Job: " + k + " Status: " + cmd.ProcessState.String())
		} else {
			buffer.WriteString("Job: " + k + " Error while retrieving")
		}
	}
	if jobmap.IsEmpty() {
		buffer.WriteString("No Job started")
	}

	log.Println("INFO: Get Status: ", buffer.String())
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: buffer.String(),
	})
}

func postMount(c *gin.Context) {
	var msg MountMessage

	if err := c.BindJSON(&msg); err != nil {
		returnErr(err, ERROR_BINDING, c)
		return
	}

	config, err := GetConfigFromVault(msg.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil || config.Agent.Gocryptfs == nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	if msg.Debug {
		log.Println("Config", config.Gocrypt)
	}
	var buffer bytes.Buffer

	if msg.Test {
		log.Println("Test Mode")
		HandleMountFolders(config.Gocrypt, DontRun, c, buffer)
		return
	}

	if msg.Run {
		HandleMountFolders(config.Gocrypt, RunJob, c, buffer)
		return
	} else {
		HandleMountFolders(config.Gocrypt, RunJobBackground, c, buffer)
		return
	}
}

func postBackup(c *gin.Context) {
	var msg BackupMessage
	if err := c.BindJSON(&msg); err != nil {
		log.Println(msg)
		returnErr(err, ERROR_BINDING, c)
		return
	}

	config, err := GetConfigFromVault(msg.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil || config.Restic == nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	var cmd *exec.Cmd
	switch msg.Mode {
	case "init":
		cmd = InitRepo(config.Restic.Environment)
	case "exist":
		cmd = ExistsRepo(config.Restic.Environment)
	case "check":
		cmd = CheckRepo(config.Restic.Environment)
	case "backup":
		cmd = Backup(
			config.Restic.Path,
			config.Restic.Environment,
			config.Restic.ExcludePath,
			2000,
			2000)
	default:
		log.Println("Not supported Mode:", msg.Mode)
		returnErr(err, ERROR_MODE, c)
		return
	}
	if msg.Debug {
		log.Println("Command: ", cmd.String())
		log.Println("Config", config.Restic)
	}

	if msg.Test {
		HandleBackup(cmd, msg.Mode, DontRun, c)
		return
	}

	if msg.Run {
		HandleBackup(cmd, msg.Mode, RunJob, c)
		return
	} else {
		HandleBackup(cmd, msg.Mode, RunJobBackground, c)
		return
	}
}
func RunRestServer(address string) (*http.Server, func()) {
	server := &http.Server{
		Addr:    "localhost:8081",
		Handler: CreateRestHandler(),
	}
	fun := func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			log.Println("Agent server closed happily...")
		} else if err != nil {
			log.Println("Agent server closed horribly...\n", err)
		}
	}
	return server, fun
}

func CreateRestHandler() http.Handler {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			JSON_MESSAGE: "pong",
		})
	})

	r.POST("/unseal", postUnseal)
	r.POST("/seal", postSeal)
	r.POST("/mount", postMount)
	r.POST("/backup", postBackup)
	r.GET("/is_sealed", getIsSealed)
	r.GET("/status", getStatus)
	r.GET("/logs", getLog)
	return r
}
