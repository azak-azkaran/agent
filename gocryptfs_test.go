package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRun *testing.T
var count int = 0

func TestGocryptfsMountGocryptfs(t *testing.T) {
	fmt.Println("running: TestGocryptfsMountGocryptfs")
	idletime, err := time.ParseDuration("3s")
	assert.NoError(t, err)
	home, err := os.Getwd()
	require.NoError(t, err)

	test_folder := strings.ReplaceAll(GOCRYPT_TEST_FOLDER, "~", home)
	test_mountpath := strings.ReplaceAll(GOCRYPT_TEST_MOUNTPATH, "~", home)

	require.DirExists(t, test_folder)

	_ = os.Mkdir(test_mountpath, 0700)
	require.DirExists(t, test_mountpath, "Folder creation failed")

	cmd := MountGocryptfs("~/test/tmp", test_mountpath, home, idletime, "hallo", false)

	assert.Contains(t, cmd.String(), "gocryptfs -i 3s", "/test/tmp", "/test/tmp-mount",
		// clear location of executable
		strings.TrimPrefix(strings.TrimPrefix(cmd.String(), "/usr/local/bin/"), "/usr/bin/"))

	job := CreateJobFromCommand(cmd, "test")
	err = job.RunJob(false)
	assert.NoError(t, err)

	require.FileExists(t, test_mountpath+GOCRYPT_TEST_FILE)
	b, err := ioutil.ReadFile(test_mountpath + GOCRYPT_TEST_FILE) // just pass the file name
	assert.NoError(t, err)
	assert.Equal(t, "testfile\n", string(b))
	time.Sleep(4 * time.Second)
	assert.NoFileExists(t, test_mountpath+GOCRYPT_TEST_FILE)
}

func TestGocryptfsMountFolders(t *testing.T) {
	fmt.Println("running: TestGocryptfsMountFolders")
	idletime, err := time.ParseDuration("3s")
	assert.NoError(t, err)

	home, err := os.Getwd()
	require.NoError(t, err)

	test_folder := strings.ReplaceAll(GOCRYPT_TEST_FOLDER, "~", home)
	test_mountpath := strings.ReplaceAll(GOCRYPT_TEST_MOUNTPATH, "~", home)
	require.DirExists(t, test_folder)

	_ = os.Mkdir(test_mountpath, 0700)
	require.DirExists(t, test_mountpath, "Folder creation failed")

	config := GocryptConfig{
		MountPoint:    GOCRYPT_TEST_MOUNTPATH,
		Path:          GOCRYPT_TEST_FOLDER,
		AllowOther:    false,
		Password:      "hallo",
		MountDuration: idletime,
	}
	var configs []GocryptConfig
	configs = append(configs, config, config)

	testRun = t
	cmds := MountFolders(home, configs)
	assert.NotEmpty(t, cmds)
	for k, v := range cmds {
		err = CheckCmd(v, "mount"+strconv.Itoa(k))
		assert.NoError(t, err)
	}
}

func CheckCmd(cmd *exec.Cmd, v string) error {
	count++
	b := assert.Contains(testRun, cmd.String(), "gocryptfs -i 3s ", "/test/tmp ", "/test/tmp-mount",
		// clear location of executable
		strings.TrimPrefix(strings.TrimPrefix(cmd.String(), "/usr/local/bin/"), "/usr/bin/"))
	if b {
		return nil
	} else {
		return errors.New("Fail")
	}
}

func TestGocryptfsIsEmpty(t *testing.T) {
	fmt.Println("running: TestGocryptfsIsEmpty")
	home, err := os.Getwd()
	require.NoError(t, err)

	test_mountpath := strings.ReplaceAll(GOCRYPT_TEST_MOUNTPATH, "~", home)

	err = IsEmpty(home, "./test")
	assert.NoError(t, err)

	_ = os.Mkdir(test_mountpath, 0700)
	require.DirExists(t, test_mountpath, "Folder creation failed")
	err = IsEmpty(home, GOCRYPT_TEST_MOUNTPATH)
	assert.NoError(t, err)
}
