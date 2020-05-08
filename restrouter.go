package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TokenMessage struct {
	Token string `json:"token" binding:"required"`
}

type BackupMessage struct {
	Mode string `json:"mode" binding:"required"`
	Run  bool   `json:"run" binding:"required"`
}

type MountMessage struct {
	Run bool `json:"run" binding:"required"`
}

func postUnseal(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println("ERROR: BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := Unseal(AgentConfiguration.VaultConfig, msg.Token)
	if err != nil {
		log.Println("ERROR: Unseal:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Println("INFO: Vault seal is: ", resp.Sealed)
	c.JSON(http.StatusOK, gin.H{
		"message": resp.Sealed,
	})
}

func postSeal(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println("ERROR: BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	AgentConfiguration.Token = msg.Token
	err = Seal(AgentConfiguration.VaultConfig, AgentConfiguration.Token)
	if err != nil {
		log.Println("ERROR: Seal:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	AgentConfiguration.Token = msg.Token
	log.Println("INFO: Vault Sealed")
	c.JSON(http.StatusOK, gin.H{
		"message": true,
	})
}

func getIsSealed(c *gin.Context) {
	b, err := IsSealed(AgentConfiguration.VaultConfig)
	if err != nil {
		log.Println("ERROR: IsSealed: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	log.Println("INFO: Vault seal is: ", b)
	c.JSON(http.StatusOK, gin.H{
		"message": b,
	})
}

func getStatus(c *gin.Context) {

}

func postConfig(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println("ERROR: BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	AgentConfiguration.Token = msg.Token
	config, err := GetConfigFromVault(AgentConfiguration.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil {
		log.Println("ERROR: GetConfigFromVault:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	AgentConfiguration.Agent = config.Agent
	AgentConfiguration.Gocrypt = config.Gocrypt
	AgentConfiguration.Restic = config.Restic
	c.JSON(http.StatusOK, gin.H{
		"message": "Recieved Agent configuration from Vault",
	})
}

func postMount(c *gin.Context) {
	var msg *MountMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println("ERROR: BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if AgentConfiguration.Gocrypt == nil {
		log.Println("ERROR: Config missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("Config missing").Error()})
		return
	}

	out, err := MountFolders(AgentConfiguration.Gocrypt, RunJob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": out,
	})
}

func postBackup(c *gin.Context) {
	var msg *BackupMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println("ERROR: BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if AgentConfiguration.Restic == nil {
		log.Println("ERROR: Config missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("Config missing").Error()})
		return
	}

	cmd := ExistsRepo(AgentConfiguration.Restic.Path, AgentConfiguration.Restic.Password)
	if msg.Run {
		exists_out, err := RunJob(cmd)
		if err != nil {
			log.Println("ERROR: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = ConcurrentQueue.Enqueue(exists_out)
		if err != nil {
			log.Println("ERROR: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	switch msg.Mode {
	case "check":
		cmd = CheckRepo(AgentConfiguration.Restic.Path,
			AgentConfiguration.Restic.Password,
		)
		err = ConcurrentQueue.Enqueue(cmd)
		if err != nil {
			log.Println("ERROR: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "backup":
		fallthrough
	default:
		cmd = Backup(AgentConfiguration.Restic.Path,
			AgentConfiguration.Restic.Password,
			AgentConfiguration.Restic.ExcludePath,
			2000,
			2000)

		err = ConcurrentQueue.Enqueue(cmd)
		if err != nil {
			log.Println("ERROR: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if msg.Run {
		go func() {
			backup_out, err := RunJob(cmd)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			err = ConcurrentQueue.Enqueue(backup_out)
			if err != nil {
				log.Println("ERROR: " + err.Error())
			}
		}()
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
			"message": "pong",
		})
	})

	r.POST("/unseal", postUnseal)
	r.POST("/seal", postSeal)
	r.POST("/mount", postMount)
	r.POST("/backup", postBackup)
	r.GET("/is_sealed", getIsSealed)
	r.GET("/status", getStatus)
	r.POST("/config", postConfig)
	return r
}
