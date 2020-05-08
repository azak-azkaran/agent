package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ERROR_PREFIX = "ERROR: "
	JSON_MESSAGE = "message"
)

type TokenMessage struct {
	Token string `json:"token" binding:"required"`
}

type BackupMessage struct {
	Mode  string `json:"mode" binding:"required"`
	Run   bool   `json:"run" binding:"required"`
	Token string `json:"token" binding:"required"`
}

type MountMessage struct {
	Run   bool   `json:"run" binding:"required"`
	Token string `json:"token" binding:"required"`
}

func postUnseal(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println(ERROR_PREFIX+"BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			JSON_MESSAGE: err.Error(),
		})
		return
	}

	resp, err := Unseal(AgentConfiguration.VaultConfig, msg.Token)
	if err != nil {
		log.Println(ERROR_PREFIX+"Unseal:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			JSON_MESSAGE: err.Error(),
		})
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
		log.Println(ERROR_PREFIX+"BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			JSON_MESSAGE: err.Error(),
		})
		return
	}
	AgentConfiguration.Token = msg.Token
	err = Seal(AgentConfiguration.VaultConfig, AgentConfiguration.Token)
	if err != nil {
		log.Println(ERROR_PREFIX+"Seal:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			JSON_MESSAGE: err.Error(),
		})
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
		log.Println(ERROR_PREFIX+"IsSealed: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			JSON_MESSAGE: err.Error(),
		})
		return
	}

	log.Println("INFO: Vault seal is: ", b)
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: b,
	})
}

func getStatus(c *gin.Context) {

}

//func postConfig(c *gin.Context) {
//	var msg TokenMessage
//	err := c.BindJSON(&msg)
//	if err != nil {
//		log.Println(ERROR_PREFIX+"BindJSON:", err.Error())
//		c.JSON(http.StatusBadRequest, gin.H{
//			JSON_MESSAGE: err.Error(),
//		})
//		return
//	}
//	AgentConfiguration.Token = msg.Token
//	config, err := GetConfigFromVault(AgentConfiguration.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
//	if err != nil {
//		log.Println(ERROR_PREFIX+"GetConfigFromVault:", err.Error())
//		c.JSON(http.StatusBadRequest, gin.H{
//			JSON_MESSAGE: err.Error(),
//		})
//		return
//	}
//	AgentConfiguration.Agent = config.Agent
//	AgentConfiguration.Gocrypt = config.Gocrypt
//	AgentConfiguration.Restic = config.Restic
//	c.JSON(http.StatusOK, gin.H{
//		JSON_MESSAGE: "Recieved Agent configuration from Vault",
//	})
//}

func postMount(c *gin.Context) {
	var msg *MountMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println(ERROR_PREFIX+"BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			JSON_MESSAGE: err.Error(),
		})
		return
	}

	config, err := GetConfigFromVault(AgentConfiguration.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil || config.Agent.Gocryptfs == nil {
		log.Println(JSON_MESSAGE + "Config missing")
		c.JSON(http.StatusBadRequest, gin.H{JSON_MESSAGE: errors.New("Config missing").Error()})
		return
	}

	if msg.Run {
		out, err := MountFolders(config.Gocrypt, RunJob)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
		}
		err = ConcurrentQueue.Enqueue(out)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
		}
	}
}

func postInitBackupt(c *gin.Context) {
	var msg *BackupMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println("ERROR: BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			JSON_MESSAGE: err.Error(),
		})
		return
	}

	config, err := GetConfigFromVault(AgentConfiguration.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil || config.Restic == nil {
		log.Println(ERROR_PREFIX + "Config missing")
		c.JSON(http.StatusBadRequest, gin.H{JSON_MESSAGE: errors.New("Config missing").Error()})
		return
	}

	cmd := InitRepo(config.Restic.Path, config.Restic.Password)
	if msg.Run {
		exists_out, err := RunJob(cmd)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
		}
		err = ConcurrentQueue.Enqueue(exists_out)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
		}
	}
}

func postBackup(c *gin.Context) {
	var msg *BackupMessage
	err := c.BindJSON(&msg)
	if err != nil {
		log.Println("ERROR: BindJSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			JSON_MESSAGE: err.Error(),
		})
		return
	}

	config, err := GetConfigFromVault(AgentConfiguration.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil || config.Restic == nil {
		log.Println(ERROR_PREFIX + "Config missing")
		c.JSON(http.StatusBadRequest, gin.H{JSON_MESSAGE: errors.New("Config missing").Error()})
		return
	}

	cmd := ExistsRepo(config.Restic.Path, config.Restic.Password)
	if msg.Run {
		exists_out, err := RunJob(cmd)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{JSON_MESSAGE: err.Error()})
			return
		}
		err = ConcurrentQueue.Enqueue(exists_out)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{JSON_MESSAGE: err.Error()})
			return
		}
	}

	switch msg.Mode {
	case "check":

		cmd = CheckRepo(config.Restic.Path,
			config.Restic.Password,
		)
		err = ConcurrentQueue.Enqueue(cmd)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{JSON_MESSAGE: err.Error()})
			return
		}
	case "backup":
		fallthrough
	default:
		cmd = Backup(config.Restic.Path,
			config.Restic.Password,
			config.Restic.ExcludePath,
			2000,
			2000)

		err = ConcurrentQueue.Enqueue(cmd)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{JSON_MESSAGE: err.Error()})
			return
		}
	}

	if msg.Run {
		backup_out, err := RunJob(cmd)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
		}
		err = ConcurrentQueue.Enqueue(backup_out)
		if err != nil {
			log.Println(ERROR_PREFIX + err.Error())
		}
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
	r.POST("/initbackup", postInitBackupt)
	r.GET("/is_sealed", getIsSealed)
	r.GET("/status", getStatus)
	return r
}
