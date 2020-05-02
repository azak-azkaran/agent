package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
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

	cmd := MountGocryptfs("./test/tmp", GOCRYPT_TEST_MOUNTPATH, idletime, "hallo", false)

	assert.Equal(t, "gocryptfs -i 3s ./test/tmp ./test/tmp-mount",
		// clear location of executable
		strings.TrimPrefix(strings.TrimPrefix(cmd.String(), "/usr/local/bin/"), "/usr/bin/"))

	out, err := RunJob(cmd)
	fmt.Println(out)
	assert.NoError(t, err)

	require.FileExists(t, GOCRYPT_TEST_FILE)
	b, err := ioutil.ReadFile(GOCRYPT_TEST_FILE) // just pass the file name
	assert.NoError(t, err)
	assert.Equal(t, "testfile\n", string(b))
	time.Sleep(4 * time.Second)
	assert.NoFileExists(t, GOCRYPT_TEST_FILE)
}
