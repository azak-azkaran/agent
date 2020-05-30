package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	GOCRYPT_TEST_MOUNTPATH = "./test/tmp-mount"
	GOCRYPT_TEST_FILE      = "./test/tmp-mount/test"
	GOCRYPT_TEST_FOLDER    = "./test/tmp"
)

var testRun *testing.T
var count int = 0

func TestGocryptfsMountGocryptfs(t *testing.T) {
	fmt.Println("running: TestGocryptfsMountGocryptfs")
	idletime, err := time.ParseDuration("3s")
	assert.NoError(t, err)
	require.DirExists(t, GOCRYPT_TEST_FOLDER)

	_ = os.Mkdir(GOCRYPT_TEST_MOUNTPATH, 0700)
	require.DirExists(t, GOCRYPT_TEST_MOUNTPATH, "Folder creation failed")

	cmd := MountGocryptfs("./test/tmp", GOCRYPT_TEST_MOUNTPATH, idletime, "hallo", false)

	assert.Contains(t, cmd.String(), "gocryptfs -i 3s ./test/tmp ./test/tmp-mount",
		// clear location of executable
		strings.TrimPrefix(strings.TrimPrefix(cmd.String(), "/usr/local/bin/"), "/usr/bin/"))

	err = RunJob(cmd, "test")
	assert.NoError(t, err)

	require.FileExists(t, GOCRYPT_TEST_FILE)
	b, err := ioutil.ReadFile(GOCRYPT_TEST_FILE) // just pass the file name
	assert.NoError(t, err)
	assert.Equal(t, "testfile\n", string(b))
	time.Sleep(4 * time.Second)
	assert.NoFileExists(t, GOCRYPT_TEST_FILE)
}

func TestGocryptfsMountFolders(t *testing.T) {
	fmt.Println("running: TestGocryptfsMountFolders")
	idletime, err := time.ParseDuration("3s")
	assert.NoError(t, err)
	require.DirExists(t, GOCRYPT_TEST_FOLDER)

	_ = os.Mkdir(GOCRYPT_TEST_MOUNTPATH, 0700)
	require.DirExists(t, GOCRYPT_TEST_MOUNTPATH, "Folder creation failed")

	config := GocryptConfig{
		MountPoint: GOCRYPT_TEST_MOUNTPATH,
		Path:       GOCRYPT_TEST_FOLDER,
		AllowOther: false,
		Password:   "hallo",
		Duration:   idletime,
	}
	var configs []GocryptConfig
	configs = append(configs, config, config)

	testRun = t
	errs, ok := MountFolders(configs, CheckCmd)
	assert.True(t, ok)
	assert.Empty(t, errs)
}

func CheckCmd(cmd *exec.Cmd, v string) error {
	count++
	b := assert.Contains(testRun, cmd.String(), "gocryptfs -i 3s ./test/tmp ./test/tmp-mount",
		// clear location of executable
		strings.TrimPrefix(strings.TrimPrefix(cmd.String(), "/usr/local/bin/"), "/usr/bin/"))
	if b {
		return nil
	} else {
		return errors.New("Fail")
	}
}

func TestGocryptfsAbsolutePath(t *testing.T) {
	fmt.Println("running: TestGocryptfsAbsolutePath")
	dir, err := os.UserHomeDir()
	assert.NoError(t, err)

	path := AbsolutePath("~/test")
	assert.True(t, strings.HasPrefix(path, dir))
	path = AbsolutePath("./test")
	assert.False(t, strings.HasPrefix(path, dir))
}

func TestGocryptfsIsEmpty(t *testing.T) {
	fmt.Println("running: TestGocryptfsIsEmpty")
	is, err := IsEmpty("./test")
	assert.NoError(t, err)
	assert.False(t, is)

	_ = os.Mkdir(GOCRYPT_TEST_MOUNTPATH, 0700)
	require.DirExists(t, GOCRYPT_TEST_MOUNTPATH, "Folder creation failed")
	is, err = IsEmpty("./test/tmp-mount")
	assert.NoError(t, err)
	assert.True(t, is)
}
