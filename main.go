package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/dgraph-io/badger/v2"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"errors"
	"log"
	"os"
	"os/exec"
	"time"

	cqueue "github.com/enriquebris/goconcurrentqueue"
	cmap "github.com/orcaman/concurrent-map"
)

var AgentConfiguration Configuration
var ConcurrentQueue *cqueue.FIFO
var jobmap cmap.ConcurrentMap

type Configuration struct {
	Agent            *AgentConfig
	Restic           *ResticConfig
	Gocrypt          []GocryptConfig
	VaultConfig      *vault.Config
	DB               *badger.DB
	Hostname         string
	Address          string
	Token            string
	PathDB           string
	TimeBetweenStart time.Duration
	Timer            *time.Timer
	MountAllow       bool
	MountDuration    string
}

type Job struct {
	Cmd    *exec.Cmd
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
}

func HandleError(err error) bool {
	if err != nil {
		log.Println("ERROR: ", err)
		err = ConcurrentQueue.Enqueue("ERROR: " + err.Error())
		if err != nil {
			log.Println("ERROR: Failed to enqueue: ", err)
		}
		return false
	}
	return true
}

func Log(toQueue string) {
	log.Println("INFO: " + toQueue)
	err := ConcurrentQueue.Enqueue(toQueue)
	if err != nil {
		log.Println("ERROR: Failed to enqueue, ", err)
	}
}

func QueueJobStatus(job Job) {
	if ConcurrentQueue == nil {
		ConcurrentQueue = cqueue.NewFIFO()
	}

	if job.Cmd.Process == nil {
		log.Println("Process not found")
		return
	}

	if job.Stdout.Len() > 0 {
		Log(job.Stdout.String())
	} else {
		Log("No Output in stdout")
	}

	if job.Stderr.Len() > 0 {
		Log(job.Stderr.String())
	} else {
		Log("No Output in stderr")
	}

}

func AddJob(cmd *exec.Cmd, name string) Job {
	if jobmap == nil {
		jobmap = cmap.New()
	}
	if jobmap.Has(name) {
		v, ok := jobmap.Get(name)
		if ok {
			oldCmd := v.(Job)
			if oldCmd.Cmd.Process != nil {
				log.Println("Found job:", name, "\tPID: ", oldCmd.Cmd.Process.Pid)
			}
		}
	}

	job := Job{
		Cmd:    cmd,
		Stdout: new(bytes.Buffer),
		Stderr: new(bytes.Buffer),
	}

	cmd.Stdout = job.Stdout
	cmd.Stderr = job.Stderr
	jobmap.Set(name, job)
	return job
}

func RunJob(cmd *exec.Cmd, name string) error {
	job := AddJob(cmd, name)
	log.Println("Starting job: ", name)
	return doJob(job)
}

func doJob(job Job) error {
	err := job.Cmd.Run()
	QueueJobStatus(job)
	return err
}

func RunJobBackground(cmd *exec.Cmd, name string) error {
	go func() {
		log.Println("Starting job in background: ", name)
		job := AddJob(cmd, name)
		err := doJob(job)
		HandleError(err)
	}()
	return nil
}

func DontRun(cmd *exec.Cmd, name string) error {
	job := AddJob(cmd, name)
	log.Println("Not Runing: ", job)
	QueueJobStatus(job)
	return nil
}

func bindEnviorment() error {
	viper.SetEnvPrefix("agent")
	err := viper.BindEnv(MAIN_ADDRESS)
	if err != nil {
		return err
	}

	err = viper.BindEnv(MAIN_PATHDB)
	if err != nil {
		return err
	}

	err = viper.BindEnv(MAIN_TIME_DURATION)
	if err != nil {
		return err
	}

	err = viper.BindEnv(MAIN_MOUNT_DURATION)
	if err != nil {
		return err
	}
	err = viper.BindEnv(MAIN_MOUNT_ALLOW)
	if err != nil {
		return err
	}

	return nil
}

func parseConfiguration(confi *Configuration) {
	if viper.IsSet(MAIN_ADDRESS) {
		confi.Address = viper.GetString(MAIN_ADDRESS)
	} else {
		confi.Address = "localhost:8081"
	}

	if viper.IsSet(MAIN_PATHDB) {
		confi.PathDB = viper.GetString(MAIN_PATHDB)
	} else {
		confi.PathDB = "/opt/agent/db"
	}

	if viper.IsSet(MAIN_TIME_DURATION) {
		dur, err := time.ParseDuration(viper.GetString(MAIN_TIME_DURATION))
		if err != nil {
			log.Println("Error parsing duration: ", err)
			dur = 30 * time.Minute
		}
		confi.TimeBetweenStart = dur
	} else {
		confi.TimeBetweenStart = 30 * time.Minute
	}

	if viper.IsSet(MAIN_MOUNT_DURATION) {
		confi.MountDuration = viper.GetString(MAIN_MOUNT_DURATION)
	} else {
		confi.MountDuration = ""
	}

	if viper.IsSet(MAIN_MOUNT_ALLOW) {
		confi.MountAllow = viper.GetBool(MAIN_MOUNT_ALLOW)
	} else {
		confi.MountAllow = false
	}

	log.Println("Agent initalzing on: ", confi.Hostname)
	log.Println("Agent Configuration:",
		"\nAddress: ", confi.Address,
		"\nPath to DB:", confi.PathDB,
		"\nTime Between Backup Runs: ", confi.TimeBetweenStart,
		"\nVaultAddress: ", confi.VaultConfig.Address,
	)
}

func Init(vaultConfig *vault.Config, args []string) error {
	ConcurrentQueue = cqueue.NewFIFO()
	jobmap = cmap.New()
	addressCommend := pflag.NewFlagSet("agent", pflag.ContinueOnError)
	addressCommend.String(MAIN_ADDRESS, "localhost:8081", "the addess on which rest server of the agent is startet")
	addressCommend.String(MAIN_PATHDB, "/opt/agent/db", "The path where to save the Database")
	addressCommend.String(MAIN_TIME_DURATION, "30m", "The duration between backups")
	addressCommend.String(MAIN_MOUNT_DURATION, "", "The Duration how long the gocrypt should be mounted")
	addressCommend.String(MAIN_MOUNT_ALLOW, "true", "If the gocrypt mount should be allowed by other users")

	err := bindEnviorment()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	config := Configuration{
		VaultConfig: vaultConfig,
		Hostname:    hostname,
	}

	err = viper.BindPFlags(addressCommend)
	if err != nil {
		return err
	}

	err = addressCommend.Parse(args)
	if err != nil {
		return err
	}

	parseConfiguration(&config)
	AgentConfiguration = config
	return nil
}

func GetConfigFromVault(token string, hostname string, vaultConfig *vault.Config) (*Configuration, error) {
	resp, err := IsSealed(vaultConfig)
	if err != nil {
		return nil, err
	}
	if resp {
		return nil, errors.New(ERROR_VAULT_SEALED)
	}

	var config Configuration
	agent, err := GetAgentConfig(vaultConfig, token, hostname)
	if err != nil {
		return nil, err
	}
	config.Agent = agent
	restic, err := GetResticConfig(vaultConfig, token, config.Agent.Restic)
	if err != nil {
		return nil, err
	}
	config.Restic = restic

	for _, name := range config.Agent.Gocryptfs {
		gocrypt, err := GetGocryptConfig(vaultConfig, token, name)
		if err != nil {
			return nil, err
		}
		gocrypt.AllowOther = true
		gocrypt.Duration, err = time.ParseDuration("0s")
		if err != nil {
			return nil, err
		}
		config.Gocrypt = append(config.Gocrypt, *gocrypt)
	}
	return &config, nil
}

func Start(Duration string, AllowOther bool) {
	if AgentConfiguration.DB == nil {
		log.Println(ERROR_DATABASE_NOT_FOUND)
		return
	}

	ok := CheckToken(AgentConfiguration.DB)
	if !ok {
		log.Println("Token is not set")
		return
	}

	token, err := GetToken(AgentConfiguration.DB)
	if err != nil {
		log.Println("Read token failed: ", err)
		return
	}

	mountMsg := MountMessage{
		Token:      token,
		Duration:   Duration,
		AllowOther: AllowOther,
	}

	reqBody, err := json.Marshal(mountMsg)
	if err != nil {
		log.Println(ERROR_UNMARSHAL, err)
		return
	}

	resp, err := http.Post("http://"+AgentConfiguration.Address+"/mount",
		"application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println(ERROR_SENDING_REQUEST, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(ERROR_READING_RESPONSE, err)
			return
		}
		bodyString := string(bodyBytes)
		log.Println("Error while mounting: ", bodyString)
	}

	backupMsg := BackupMessage{
		Mode:  "backup",
		Token: token,
	}

	reqBody, err = json.Marshal(backupMsg)
	if err != nil {
		log.Println(ERROR_UNMARSHAL, err)
		return
	}

	resp, err = http.Post("http://"+AgentConfiguration.Address+"/backup",
		"application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println(ERROR_SENDING_REQUEST, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(ERROR_READING_RESPONSE, err)
			return
		}
		bodyString := string(bodyBytes)
		log.Println("Error while backup: ", bodyString)
	} else {
		UpdateTimestamp(AgentConfiguration.DB, time.Now())
	}

}

func run() {
	Start(AgentConfiguration.MountDuration, AgentConfiguration.MountAllow)
	AgentConfiguration.Timer = time.AfterFunc(AgentConfiguration.TimeBetweenStart, run)
}

func main() {
	err := Init(vault.DefaultConfig(), os.Args)
	//log.Print("Please enter Token: ")
	//password, err := terminal.ReadPassword(int(syscall.Stdin))
	HandleError(err)
	AgentConfiguration.DB = InitDB(AgentConfiguration.PathDB, false)
	_, fun := RunRestServer(AgentConfiguration.Address)

	go func() {
		log.Println("Starting Run Function in 5 Seconds")
		time.AfterFunc(5*time.Second, run)
	}()
	log.Println("Starting the Rest Server")
	fun()
}
