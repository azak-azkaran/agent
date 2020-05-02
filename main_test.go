package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

const (
	GOCRYPT_TEST_MOUNTPATH = "./test/tmp-mount"
	GOCRYPT_TEST_FILE      = "./test/tmp-mount/test"
)

func TestMountGocryptfs(t *testing.T) {
	fmt.Println("running: TestMountGocryptfs")
	idletime, err := time.ParseDuration("3s")
	assert.NoError(t, err)
	require.DirExists(t, "./test/tmp")

	_ = os.Mkdir(GOCRYPT_TEST_MOUNTPATH, 0700)
	require.DirExists(t, GOCRYPT_TEST_MOUNTPATH, "Folder creation failed")

	cmd := MountGocryptfs("./test/tmp", GOCRYPT_TEST_MOUNTPATH, idletime, "hallo")

	assert.Equal(t, "/usr/local/bin/gocryptfs -allow_other -i 3s ./test/tmp ./test/tmp-mount", cmd.String())
	_, err = RunJob(cmd)
	assert.NoError(t, err)

	require.FileExists(t, GOCRYPT_TEST_FILE)
	b, err := ioutil.ReadFile(GOCRYPT_TEST_FILE) // just pass the file name
	assert.NoError(t, err)
	assert.Equal(t, "testfile\n", string(b))
	time.Sleep(4 * time.Second)
	assert.NoFileExists(t, GOCRYPT_TEST_FILE)
}
