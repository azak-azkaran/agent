package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strconv"

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

func HandleGit(mode string, v GitConfig, run bool, printOutput bool, home string) (bool, error) {
	var job Job
	switch mode {
	case "clone":
		job = CreateJobFromFunction(func() error {
			return GitClone(v.Rep, v.Directory, home, v.PersonalToken)
		}, mode+" "+v.Name)
	case "pull":
		job = CreateJobFromFunction(func() error {
			err := GitCreateRemote(v.Directory, home, v.Rep)
			if err != nil {
				return err
			} else {
				return GitPull(v.Directory, home, v.PersonalToken)
			}
		}, mode+" "+v.Name)
	default:
		return false, errors.New("Not supported Mode: " + mode)
	}

	var err error

	if run {
		err = job.RunJob(printOutput)
	} else {
		err = job.RunJobBackground(printOutput)
	}

	return err == nil, err
}

func HandleBackup(cmd *exec.Cmd, name string, printOutput bool, test bool, run bool, c *gin.Context) {
	job := CreateJobFromCommand(cmd, name)
	var err error
	if test {
		err = job.DontRun(printOutput)
	} else {
		if run {
			err = job.RunJob(printOutput)
		} else {
			err = job.RunJobBackground(printOutput)
		}
	}

	if err != nil {
		returnErr(err, ERROR_RUNBACKUP, c)
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}

}

func handleError(job Job, err error, errMsg string, buffer bytes.Buffer) bool {
	if err != nil {
		m := ERROR_PREFIX + errMsg + err.Error()
		if job.Stderr != nil {
			m = "\t" + job.Stderr.String()
		}
		log.Println(m)
		buffer.WriteString(m)
		return false
	}
	return true
}

func HandleMount(job Job, printOutput bool, test bool, run bool, c *gin.Context, buffer bytes.Buffer) bool {
	var err error
	if test {
		err = job.DontRun(printOutput)
		return handleError(job, err, ERROR_RUNMOUNT, buffer)
	} else {

		if run {
			err = job.RunJob(printOutput)
			return handleError(job, err, ERROR_RUNMOUNT, buffer)

		} else {
			err = job.RunJobBackground(printOutput)
			return handleError(job, err, ERROR_RUNMOUNT, buffer)
		}
	}
}

func HandleMountFolders(cmds []*exec.Cmd, printOutput bool, test bool, run bool, c *gin.Context, buffer bytes.Buffer) {
	ok := true
	for k, v := range cmds {
		job := CreateJobFromCommand(v, "mount"+strconv.Itoa(k))
		if !HandleMount(job, printOutput, test, run, c, buffer) {
			ok = false
		}
	}
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			JSON_MESSAGE: buffer.String(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			JSON_MESSAGE: buffer.String(),
		})
	}
}

func returnErr(err error, source string, c *gin.Context) {
	log.Println(ERROR_PREFIX+source, err.Error())

	var code int
	if source == ERROR_CONFIG {
		code = http.StatusForbidden
	} else {
		code = http.StatusInternalServerError
	}
	c.JSON(code, gin.H{
		JSON_MESSAGE: err.Error(),
	})
}

func postUnseal(c *gin.Context) {
	var msg TokenMessage
	if err := c.BindJSON(&msg); err != nil {
		log.Println(ERROR_BINDING, err.Error())
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
	if err := c.BindJSON(&msg); err != nil {
		log.Println(ERROR_BINDING, err.Error())
		return
	}

	AgentConfiguration.Token = msg.Token
	err := Seal(AgentConfiguration.VaultConfig, AgentConfiguration.Token)
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
	log.Println("INFO: Log: ", "not implemented")
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: "not implemented",
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

	log.Println("INFO: Get Status: ", buffer.String())
	c.JSON(http.StatusOK, gin.H{
		JSON_MESSAGE: buffer.String(),
	})
}

func postMount(c *gin.Context) {
	var msg MountMessage
	if err := c.BindJSON(&msg); err != nil {
		log.Println(ERROR_BINDING, err.Error())
		return
	}

	config, err := CreateConfigFromVault(msg.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	err = config.GetGocryptConfig()
	if err != nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	out := MountFolders(config.Agent.HomeFolder, config.Gocrypt)

	if msg.Debug {
		log.Println("Config", config.Gocrypt)
		for k, v := range out {
			log.Println("Command", k, ": ", v.String())
		}
	}
	var buffer bytes.Buffer

	HandleMountFolders(out, msg.PrintOutput, msg.Test, msg.Run, c, buffer)
}

func postBackup(c *gin.Context) {
	var msg BackupMessage
	if err := c.BindJSON(&msg); err != nil {
		log.Println(ERROR_BINDING, err.Error())
		return
	}

	config, err := CreateConfigFromVault(msg.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	err = config.GetResticConfig()
	if err != nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	log.Println("config:", config)

	var cmd *exec.Cmd
	switch msg.Mode {
	case "init":
		cmd = InitRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "exist":
		cmd = ExistsRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "check":
		cmd = CheckRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "backup":
		cmd = Backup(
			config.Restic.Path,
			config.Restic.Environment,
			config.Agent.HomeFolder,
			config.Restic.ExcludePath,
			2000,
			2000)
	case "unlock":
		cmd = UnlockRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "list":
		cmd = ListRepo(config.Restic.Environment, config.Agent.HomeFolder)
		msg.PrintOutput = true
	case "forget":
		cmd = ForgetRep(config.Restic.Environment, config.Agent.HomeFolder)
	default:
		returnErr(errors.New("Not supported Mode: "+msg.Mode), ERROR_MODE, c)
		return
	}
	if msg.Debug {
		log.Println("Command: ", cmd.String())
		log.Println("Config", config.Restic)
	}

	HandleBackup(cmd, msg.Mode, msg.PrintOutput, msg.Test, msg.Run, c)

}

func postToken(c *gin.Context) {
	var msg TokenMessage
	if err := c.BindJSON(&msg); err != nil {
		log.Println(ERROR_BINDING, err.Error())
		return
	}

	ok, err := PutToken(AgentConfiguration.DB, msg.Token)
	if err != nil {
		returnErr(err, ERROR_PUT_TOKEN, c)
		return
	}

	if ok {
		c.JSON(http.StatusOK, gin.H{})
		return
	} else {
		returnErr(errors.New("Storage Error PUT returned false"), ERROR_PUT_TOKEN, c)
		return
	}
}

func getToken(c *gin.Context) {
	token, err := GetToken(AgentConfiguration.DB)
	if err != nil {
		returnErr(err, ERROR_PUT_TOKEN, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func postUnsealKey(c *gin.Context) {
	var msg VaultKeyMessage
	if err := c.BindJSON(&msg); err != nil {
		log.Println(ERROR_BINDING, err.Error())
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
		log.Println(ERROR_BINDING, err.Error())
		return
	}

	config, err := CreateConfigFromVault(msg.Token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	err = config.GetGitConfig()
	if err != nil {
		returnErr(err, ERROR_CONFIG, c)
		return
	}

	var buffer bytes.Buffer
	ok := true
	for _, v := range config.Git {
		ok, err = HandleGit(msg.Mode, v, msg.Run, msg.PrintOutput, config.Agent.HomeFolder)
		if !ok && err != nil {
			buffer.WriteString("\nJob: " + v.Name + " " + err.Error())
		}
	}

	if ok {
		c.JSON(http.StatusOK, gin.H{
			JSON_MESSAGE: buffer.String(),
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			JSON_MESSAGE: buffer.String(),
		})
	}
}

func RunRestServer(address string) (*http.Server, func()) {
	server := &http.Server{
		Addr:    address,
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

	r.POST("/unsealkey", postUnsealKey)
	r.POST("/token", postToken)
	r.POST("/unseal", postUnseal)
	r.POST("/seal", postSeal)
	r.POST("/mount", postMount)
	r.POST("/backup", postBackup)
	r.POST("/git", postGit)
	r.GET("/is_sealed", getIsSealed)
	r.GET("/status", getStatus)
	r.GET("/logs", getLog)
	r.GET("/token", getToken)
	return r
}
