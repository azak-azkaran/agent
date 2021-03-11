package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clear() {
	viper.Reset()
	pwd, _ := os.Getwd()
	test_folder := strings.ReplaceAll(BACKUP_TEST_FOLDER, HOME, pwd)
	err := RemoveContents(test_folder)
	if err != nil {
		fmt.Println("Error cleaning up: ", err.Error())
	}

	test_folder = strings.ReplaceAll(GIT_TEST_FOLDER, HOME, pwd)
	err = RemoveContents(test_folder)
	if err != nil {
		fmt.Println("Error cleaning up: ", err.Error())
	}
	os.Remove(test_folder)

	test_folder = strings.ReplaceAll(GIT_TEST_FOLDER_VIMRC, HOME, pwd)
	err = RemoveContents(test_folder)
	if err != nil {
		fmt.Println("Error cleaning up: ", err.Error())
	}
	os.Remove(test_folder)
	forbidden = false


	os.Remove("AGENT_ADDRESS")
	os.Remove("AGENT_DURATION")
	os.Remove("AGENT_PATHDB")
	os.Remove("AGENT_MOUNT_DURATION")
	os.Remove("AGENT_MOUNT_ALLOW")
	os.Remove("AGENT_VAULT_ROLE_ID")
	os.Remove("AGENT_VAULT_SECRET_ID")
}

func TestMainInit(t *testing.T) {
	fmt.Println("running: TestMainInit")
	t.Cleanup(clear)
	testconfig := readConfig(t)
	hostname, err := os.Hostname()
	require.NoError(t, err)

	// Test with flags
	var args []string
	args = append(args, "--address="+MAIN_TEST_ADDRESS)
	args = append(args, "--pathdb="+MAIN_TEST_PATHDB)
	args = append(args, "--duration="+MAIN_TEST_DURATION)
	args = append(args, "--mount_duration="+MAIN_TEST_MOUNT_DURATION)
	args = append(args, "--mount_allow="+MAIN_TEST_MOUNT_ALLOW)
	args = append(args, "--vault_role_id="+VAULT_TEST_ROLE_ID)
	args = append(args, "--vault_secret_id="+VAULT_TEST_SECRET_ID)

	err = Init(testconfig.config, args)
	require.NoError(t, err)
	assert.Equal(t, hostname, AgentConfiguration.Hostname)
	assert.Equal(t, MAIN_TEST_ADDRESS, AgentConfiguration.Address)
	assert.Equal(t, MAIN_TEST_PATHDB, AgentConfiguration.PathDB)
	assert.Equal(t, false, AgentConfiguration.MountAllow)
	assert.Equal(t, MAIN_TEST_MOUNT_DURATION, AgentConfiguration.MountDuration)
	assert.Equal(t, VAULT_TEST_SECRET_ID, AgentConfiguration.SecretID)
	assert.Equal(t, VAULT_TEST_ROLE_ID, AgentConfiguration.RoleID)
	assert.True(t, AgentConfiguration.useLogin)

	dur, err := time.ParseDuration("1h30m")
	assert.NoError(t, err)

	assert.Equal(t, AgentConfiguration.TimeBetweenStart, dur)
	length := len(os.Args)

	os.Args = os.Args[:length-1]

	// Test with Environment variables
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_PATHDB", MAIN_TEST_PATHDB)
	os.Setenv("AGENT_DURATION", MAIN_TEST_DURATION)
	os.Setenv("AGENT_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGENT_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)

	err = Init(testconfig.config, nil)
	require.NoError(t, err)
	assert.Equal(t, hostname, AgentConfiguration.Hostname)
	assert.Equal(t, MAIN_TEST_ADDRESS, AgentConfiguration.Address)
	assert.Equal(t, MAIN_TEST_PATHDB, AgentConfiguration.PathDB)
	assert.Equal(t, dur, AgentConfiguration.TimeBetweenStart)
	assert.Equal(t, false, AgentConfiguration.MountAllow)
	assert.Equal(t, MAIN_TEST_MOUNT_DURATION, AgentConfiguration.MountDuration)
	assert.False(t, AgentConfiguration.useLogin)
}

func TestMainStart(t *testing.T) {
	fmt.Println("running: TestMainStart")
	t.Cleanup(clear)

	jobmap = cmap.New()
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	os.Setenv("AGENT", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGENT", MAIN_TEST_MOUNT_ALLOW)
	err := Init(testconfig.config, os.Args)
	require.NoError(t, err)

	_, err = Unseal(testconfig.config, testconfig.secret)
	require.NoError(t, err)

	AgentConfiguration.DB = InitDB("", "", true)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	go fun()
	Start()

	time.Sleep(1 * time.Millisecond)
	assert.Eventually(t, func() bool { return checkJobmap("check") },
		time.Duration(20*time.Second), time.Duration(1*time.Second))

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)
}

func TestMainMain(t *testing.T) {
	fmt.Println("running: TestMainMain")
	t.Cleanup(func() {
		os.Remove(MAIN_TEST_KEYFILE_PATH)
		clear()
	})
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	os.Setenv("AGENT_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGENT_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)
	os.Setenv("AGENT_VAULT_KEY_FILE", MAIN_TEST_KEYFILE_PATH)
	multipleKey = true
	sealStatus = true
	Progress = 0

	home, err := os.Getwd()
	require.NoError(t, err)

	f, err := os.Create(MAIN_TEST_KEYFILE_PATH)
	require.NoError(t, err)
	w := bufio.NewWriter(f)
	key := "test"
	for i := 1; i < 6; i++ {
		_, err := w.WriteString(key + strconv.Itoa(i) + "\n")
		assert.NoError(t, err)
	}
	err = w.Flush()
	assert.NoError(t, err)
	require.FileExists(t, MAIN_TEST_KEYFILE_PATH)

	go main()
	time.Sleep(1 * time.Second)
	AgentConfiguration.VaultConfig = testconfig.config

	sendingGet(t, REST_TEST_PING, http.StatusOK)

	time.Sleep(10 * time.Second)
	assert.Eventually(t, checkContents, 120*time.Second, 1*time.Second)
	assert.Eventually(t, func() bool { return checkJobmap("check") },
		120*time.Second, 1*time.Second)

	stopChan <- syscall.SIGINT
	time.Sleep(10 * time.Second)

	test_mountpath := strings.ReplaceAll(GOCRYPT_TEST_MOUNTPATH, "~", home)

	err = IsEmpty(home, test_mountpath)
	assert.NoError(t, err)

	assert.False(t, sealStatus)
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)
	multipleKey = false
}

func TestMainCheckKeyFile(t *testing.T) {
	fmt.Println("Running: TestMainCheckKeyFile")
	t.Cleanup(func() {
		os.Remove(MAIN_TEST_KEYFILE_PATH)
		clear()
	})
	AgentConfiguration.DB = InitDB("", "", true)
	require.NotNil(t, AgentConfiguration.DB)

	require.NoFileExists(t, MAIN_TEST_KEYFILE_PATH)
	err := CheckKeyFile(MAIN_TEST_KEYFILE_PATH)
	assert.Error(t, err)

	f, err := os.Create(MAIN_TEST_KEYFILE_PATH)
	require.NoError(t, err)
	w := bufio.NewWriter(f)

	key := "test"
	for i := 1; i < 6; i++ {
		_, err := w.WriteString(key + strconv.Itoa(i) + "\n")
		assert.NoError(t, err)
	}
	err = w.Flush()
	assert.NoError(t, err)
	require.FileExists(t, MAIN_TEST_KEYFILE_PATH)

	err = CheckKeyFile(MAIN_TEST_KEYFILE_PATH)
	assert.NoError(t, err)

	ok := CheckSealKey(AgentConfiguration.DB, 5)
	assert.True(t, ok)
}

func TestMainBackupRepositoryExists(t *testing.T) {
	fmt.Println("running: TestMainBackupRepositoryExists")
	t.Cleanup(clear)
	jobmap = cmap.New()
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	os.Setenv("AGENT_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGENT_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)
	err := Init(testconfig.config, os.Args)
	require.NoError(t, err)

	_, err = Unseal(testconfig.config, testconfig.secret)
	require.NoError(t, err)

	pwd, err := os.Getwd()
	require.NoError(t, err)

	AgentConfiguration.DB = InitDB("", "", true)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	go fun()
	time.Sleep(1 * time.Millisecond)

	BackupRepositoryExists(VAULT_TEST_TOKEN)

	assert.Eventually(t, func() bool {

		path := strings.ReplaceAll(BACKUP_TEST_CONF_FILE, HOME, pwd)
		stat, err := os.Stat(path)
		if err != nil {
			return false
		}

		return !stat.IsDir()
	},
		time.Duration(20*time.Second), time.Duration(1*time.Second))

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)

}

func TestMainCheckBackupRepository(t *testing.T) {
	fmt.Println("running: TestMainCheckBackupRepository")
	t.Cleanup(clear)

	jobmap = cmap.New()
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	os.Setenv("AGENT_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGENT_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)
	os.Setenv("AGENT_VAULT_ROLE_ID", VAULT_TEST_ROLE_ID)
	os.Setenv("AGENT_VAULT_SECRET_ID", VAULT_TEST_SECRET_ID)

	err := Init(testconfig.config, os.Args)
	require.NoError(t, err)

	_, err = Unseal(testconfig.config, testconfig.secret)
	require.NoError(t, err)

	AgentConfiguration.DB = InitDB("", "", true)
	time.Sleep(1 * time.Millisecond)
	CheckBackupRepository()

	assert.Eventually(t, func() bool {
		timestamp, err := GetTimestamp(AgentConfiguration.DB)
		if err != nil {
			return false
		}

		return timestamp != time.Unix(0, 0)
	},
		time.Duration(20*time.Second), time.Duration(1*time.Second))

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)
}

func TestMainGitCheckout(t *testing.T) {
	fmt.Println("Running: TestMainGitCheckout")
	t.Cleanup(clear)
	jobmap = cmap.New()
	gin.SetMode(gin.TestMode)
	testconfig := readConfig(t)
	os.Setenv("AGENT_ADDRESS", MAIN_TEST_ADDRESS)
	os.Setenv("AGENT_DURATION", testconfig.Duration)
	os.Setenv("AGENT_PATHDB", "./test/DB")
	os.Setenv("AGENT_MOUNT_DURATION", MAIN_TEST_MOUNT_DURATION)
	os.Setenv("AGENT_MOUNT_ALLOW", MAIN_TEST_MOUNT_ALLOW)
	err := Init(testconfig.config, os.Args)
	require.NoError(t, err)

	_, err = Unseal(testconfig.config, testconfig.secret)
	require.NoError(t, err)

	pwd, err := os.Getwd()
	require.NoError(t, err)

	AgentConfiguration.DB = InitDB("", "", true)
	server, fun := RunRestServer(MAIN_TEST_ADDRESS)
	go fun()
	time.Sleep(1 * time.Millisecond)

	GitCheckout()

	assert.Eventually(t, func() bool {
		path := strings.ReplaceAll(GIT_TEST_FOLDER, HOME, pwd)
		f, err := os.Stat(path + "/.git")
		if err != nil {
			return false
		}
		return f.IsDir()
	},
		time.Duration(20*time.Second), time.Duration(1*time.Second))

	GitCheckout()

	assert.Eventually(t, func() bool { return checkJobmap("pull") },
		time.Duration(20*time.Second), time.Duration(1*time.Second))

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	AgentConfiguration.DB.Close()
	assert.NoError(t, err)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
	time.Sleep(1 * time.Millisecond)

}

func checkJobmap(jobname string) bool {
	for item := range jobmap.IterBuffered() {
		if strings.Contains(item.Key, jobname) {
			j := item.Val.(*Job)
			return j.IsFinished()
		}
	}
	return false
}

func checkContents() bool {
	home, err := os.Getwd()
	if err != nil {
		return false
	}

	err = IsEmpty(home, GOCRYPT_TEST_MOUNTPATH)
	if err != nil {
		return false
	}

	err = IsEmpty(home, BACKUP_TEST_FOLDER)
	if err != nil {
		return false
	}

	return true
}
