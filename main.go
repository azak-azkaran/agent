package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgraph-io/badger/v2"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	cmap "github.com/orcaman/concurrent-map"
)

var AgentConfiguration Configuration
var stopChan = make(chan os.Signal, 2)
var restServerAgent *http.Server

type Configuration struct {
	Agent            *AgentConfig
	Restic           *ResticConfig
	Gocrypt          []GocryptConfig
	Git              []GitConfig
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
	VaultKeyFile     string
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

	err = viper.BindEnv(MAIN_VAULT_KEY_FILE)
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

	if viper.IsSet(MAIN_VAULT_KEY_FILE) {
		confi.VaultKeyFile = viper.GetString(MAIN_VAULT_KEY_FILE)
	} else {
		confi.VaultKeyFile = ""
	}

	if viper.IsSet(MAIN_VAULT_ADDRESS) {
		confi.VaultConfig.Address = viper.GetString(MAIN_VAULT_ADDRESS)
	}

	log.Println("Agent initalzing on: ", confi.Hostname)
	log.Println("Agent Configuration:",
		"\nAddress: ", confi.Address,
		"\nPath to DB:", confi.PathDB,
		"\nTime Between Backup Runs: ", confi.TimeBetweenStart,
		"\nVault Address: ", confi.VaultConfig.Address,
		"\nVault KeyFile path: ", confi.VaultKeyFile,
		"\nMount Duration: ", confi.MountDuration,
		"\nMount AllowOther: ", confi.MountAllow,
	)
}

func Init(vaultConfig *vault.Config, args []string) error {
	jobmap = cmap.New()
	addressCommend := pflag.NewFlagSet("agent", pflag.ContinueOnError)
	addressCommend.String(MAIN_ADDRESS, "localhost:8081", "the addess on which rest server of the agent is startet")
	addressCommend.String(MAIN_PATHDB, "/opt/agent/db", "The path where to save the Database")
	addressCommend.String(MAIN_TIME_DURATION, "30m", "The duration between backups")
	addressCommend.String(MAIN_MOUNT_DURATION, "", "The Duration how long the gocrypt should be mounted")
	addressCommend.String(MAIN_MOUNT_ALLOW, "true", "If the gocrypt mount should be allowed by other users")
	addressCommend.String(MAIN_VAULT_KEY_FILE, "", "File in which the vault keys are stored for easy save into Badger database")
	addressCommend.String(MAIN_VAULT_ADDRESS, "https://localhost:8200", "The address to the vault server")

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

func CheckKeyFile(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return errors.New(MAIN_ERROR_IS_DIR)
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()
	reader := bufio.NewScanner(f)

	var keys []string
	for reader.Scan() {
		keys = append(keys, reader.Text())
	}

	for k, v := range keys {
		_, err = PutSealKey(AgentConfiguration.DB, v, k+1)
		if err != nil {
			DropSealKeys(AgentConfiguration.DB)
			return err
		}
	}
	return nil
	//return errors.New("Not implemented yet")
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

	crypts := strings.Split(config.Agent.Gocryptfs, ",")
	for _, name := range crypts {
		gocrypt, err := GetGocryptConfig(vaultConfig, token, name)
		if err != nil {
			return nil, err
		}
		config.Gocrypt = append(config.Gocrypt, *gocrypt)
	}

	gits := strings.Split(config.Agent.Git, ",")
	for _, name := range gits {
		git, err := GetGitConfig(vaultConfig, token, name)
		if err != nil {
			return nil, err
		}
		config.Git = append(config.Git, *git)
	}
	return &config, nil
}

func checkRequirementsForBackup() (string, bool) {
	if AgentConfiguration.DB == nil {
		log.Println(ERROR_DATABASE_NOT_FOUND)
		return "", false
	}

	ok := CheckToken(AgentConfiguration.DB)
	if !ok {
		log.Println("Token is not set")
		return "", false
	}

	token, err := GetToken(AgentConfiguration.DB)
	if err != nil {
		log.Println("Read token failed: ", err)
		return "", false
	}
	return token, true
}

func CheckBackupRepository() {
	token, ok := checkRequirementsForBackup()
	if !ok {
		return
	}

	t, err := GetTimestamp(AgentConfiguration.DB)
	if err != nil {
		log.Println(ERROR_TIMESTAMP, err)
	}
	log.Println("Last Backup Check: ", t.String())

	t = t.Add(12 * time.Hour)
	now := time.Now()
	log.Println("Next Backup Check after: ", t.String())
	if now.After(t) {
		BackupRepositoryExists(token)
		backupMsg := BackupMessage{
			Mode:        "check",
			Token:       token,
			PrintOutput: true,
			Run:         true,
		}

		reqBody, err := json.Marshal(backupMsg)
		if err != nil {
			log.Println(ERROR_UNMARSHAL, err)
			return
		}

		ok, err = SendRequest(reqBody, MAIN_POST_BACKUP_ENDPOINT)
		if err != nil {
			return
		}
		if ok {
			_, err = UpdateTimestamp(AgentConfiguration.DB, time.Now())
			if err != nil {
				log.Println(err)
			}
			return
		}
	} else {
		log.Println(MAIN_MESSAGE_BACKUP_ALREADY, t.String())
	}

}

func mountFolders() {
	token, ok := checkRequirementsForBackup()
	if !ok {
		return
	}

	mountMsg := MountMessage{
		Token: token,
		Run:   true,
	}

	reqBody, err := json.Marshal(mountMsg)
	if err != nil {
		log.Println(ERROR_UNMARSHAL, err)
		return
	}

	ok, err = SendRequest(reqBody, MAIN_POST_MOUNT_ENDPOINT)
	if err != nil {
		return
	}
	if !ok {
		return
	}
}

func backup() {
	token, ok := checkRequirementsForBackup()
	if !ok {
		return
	}

	t, err := GetLastBackup(AgentConfiguration.DB)
	if err != nil {
		log.Println(ERROR_TIMESTAMP, err)
	}
	log.Println("Last Backup: ", t.String())

	t = t.Add(2 * time.Hour)
	now := time.Now()
	log.Println("Next Backup after: ", t.String())
	if now.After(t) {
		BackupRepositoryExists(token)
		backupMsg := BackupMessage{
			Mode:        "backup",
			Token:       token,
			PrintOutput: true,
			Run:         true,
		}

		reqBody, err := json.Marshal(backupMsg)
		if err != nil {
			log.Println(ERROR_UNMARSHAL, err)
			return
		}

		ok, err = SendRequest(reqBody, MAIN_POST_BACKUP_ENDPOINT)
		if err != nil {
			return
		}
		if ok {
			log.Println(MAIN_MESSAGE_BACKUP_SUCCESS)
			UpdateLastBackup(AgentConfiguration.DB, time.Now())
			return
		}

	}
}

func BackupRepositoryExists(token string) {
	backupMsg := BackupMessage{
		Mode:        "exist",
		Token:       token,
		Run:         true,
		PrintOutput: false,
	}

	reqBody, err := json.Marshal(backupMsg)
	if err != nil {
		log.Println(ERROR_UNMARSHAL, err)
		return
	}
	ok, err := SendRequest(reqBody, MAIN_POST_BACKUP_ENDPOINT)
	if err != nil {
		return
	}
	if ok {
		return
	}

	log.Println(MAIN_MESSAGE_BACKUP_INIT)
	backupMsg.Mode = "init"
	reqBody, err = json.Marshal(backupMsg)
	if err != nil {
		log.Println(ERROR_UNMARSHAL, err)
		return
	}

	ok, err = SendRequest(reqBody, MAIN_POST_BACKUP_ENDPOINT)
	if err != nil {
		log.Println(err)
		return
	}
	if ok {
		return
	}
}

func GitCheckout() {
	token, ok := checkRequirementsForBackup()
	if !ok {
		return
	}

	msg := GitMessage{
		Mode:        "clone",
		Token:       token,
		Run:         true,
		PrintOutput: true,
	}

	reqBody, err := json.Marshal(msg)
	if err != nil {
		log.Println(ERROR_UNMARSHAL, err)
		return
	}

	ok, err = SendRequest(reqBody, MAIN_POST_GIT_ENDPOINT)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	if ok {
		return
	} else {
		msg.Mode = "pull"
		reqBody, err := json.Marshal(msg)
		if err != nil {
			log.Println(ERROR_UNMARSHAL, err)
			return
		}

		SendRequest(reqBody, MAIN_POST_GIT_ENDPOINT)
	}

}

func Start() {
	log.Println("Waking from Sleep")
	mountFolders()
	GitCheckout()
	backup()
	CheckBackupRepository()

	log.Println("Going to Sleep")
}

func unsealVault(seal *vault.SealStatusResponse) {
	if CheckSealKey(AgentConfiguration.DB, seal.N) {
		log.Println(MAIN_MESSAGE_START_UNSEAL)
		values := GetSealKey(AgentConfiguration.DB, seal.T, seal.N)
		for _, v := range values {
			_, err := Unseal(AgentConfiguration.VaultConfig, v)
			if err != nil {
				log.Println(MAIN_ERROR_UNSEAL, err)
			}
		}
	} else {
		log.Println(MAIN_MESSAGE_NOT_ENOUGH_KEYS)

	}
}

func run() {
	seal, err := SealStatus(AgentConfiguration.VaultConfig)
	if err != nil {
		log.Println(MAIN_ERROR_CHECK_SEAL, err)
	}

	if seal.Sealed {
		log.Println(ERROR_VAULT_SEALED)
		unsealVault(seal)
	}

	Start()
	AgentConfiguration.Timer = time.AfterFunc(AgentConfiguration.TimeBetweenStart, run)
}

func SendRequest(reqBody []byte, endpoint string) (bool, error) {
	resp, err := http.Post(MAIN_POST_HTTP+AgentConfiguration.Address+endpoint,
		MAIN_POST_DATA_TYPE, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println(ERROR_SENDING_REQUEST, err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(ERROR_READING_RESPONSE, err)
			return false, err
		}
		bodyString := string(bodyBytes)
		log.Println("Error while sending to:", endpoint, ": ", bodyString)
		return false, nil
	}
	return true, nil
}

func main() {
	stopChan = make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt)
	go func() {
		<-stopChan
		log.Println("Stopping Agent Happly")
		if AgentConfiguration.Timer != nil {
			AgentConfiguration.Timer.Stop()
		}

		if AgentConfiguration.DB != nil {
			Close(AgentConfiguration.DB, 5*time.Millisecond)
		}

		err := restServerAgent.Shutdown(context.Background())
		if err != nil {
			log.Println(MAIN_ERROR_SHUTDOWN, err)
		}
	}()
	err := Init(vault.DefaultConfig(), os.Args)

	//log.Print("Please enter Token: ")
	//password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Println("ERROR", err)
	}
	AgentConfiguration.DB = InitDB(AgentConfiguration.PathDB, "", false)

	if AgentConfiguration.VaultKeyFile != "" {
		err = CheckKeyFile(AgentConfiguration.VaultKeyFile)

		if err != nil {
			log.Println("ERROR", err)
		}
	}
	var fun func()
	restServerAgent, fun = RunRestServer(AgentConfiguration.Address)

	go func() {
		log.Println(MAIN_MESSAGE_START_RUNNING, "\t", AgentConfiguration.Hostname)
		AgentConfiguration.Timer = time.AfterFunc(5*time.Second, run)
	}()

	log.Println(MAIN_MESSAGE_START_RESTSERVER)
	fun()
}
