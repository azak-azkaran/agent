package main

import (
	"bytes"

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
	Agent       *AgentConfig
	Restic      *ResticConfig
	Gocrypt     []GocryptConfig
	VaultConfig *vault.Config
	Hostname    string
	Address     string
	Token       string
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

func Init(vaultConfig *vault.Config, args []string) error {
	ConcurrentQueue = cqueue.NewFIFO()
	jobmap = cmap.New()
	addressCommend := pflag.NewFlagSet("address", pflag.ContinueOnError)
	addressCommend.String("address", "localhost:8081", "the addess on which rest server of the agent is startet")
	viper.SetEnvPrefix("agent")
	err := viper.BindEnv("address")
	if err != nil {
		return err
	}
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	log.Println("Agent initalzing on: ", hostname)
	AgentConfiguration = Configuration{
		VaultConfig: vaultConfig,
		Hostname:    hostname,
	}
	err = addressCommend.Parse(args)
	if err != nil {
		return err
	}

	if viper.IsSet("address") {
		AgentConfiguration.Address = viper.GetString("address")
	} else {
		AgentConfiguration.Address = "localhost:8081"
	}
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

func main() {
	err := Init(vault.DefaultConfig(), os.Args)
	//log.Print("Please enter Token: ")
	//password, err := terminal.ReadPassword(int(syscall.Stdin))
	HandleError(err)
	_, fun := RunRestServer("localhost:8081")
	fun()

}
