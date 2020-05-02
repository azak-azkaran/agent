package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"time"
)

func TestMountGocryptfs(t *testing.T) {
	fmt.Println("running: TestMountGocryptfs")
	idletime, err := time.ParseDuration("3s")
	assert.NoError(t, err)
	require.DirExists(t, "./test/tmp")
	require.DirExists(t, "./test/tmp-mount")
	cmd := MountGocryptfs("./test/tmp", "./test/tmp-mount", idletime, "hallo")

	assert.Equal(t, "/usr/local/bin/gocryptfs -allow_other -i 3s ./test/tmp ./test/tmp-mount", cmd.String())
	_, err = RunJob(cmd)
	assert.NoError(t, err)

	require.FileExists(t, "./test/tmp-mount/test")
	b, err := ioutil.ReadFile("./test/tmp-mount/test") // just pass the file name
	assert.NoError(t, err)
	assert.Equal(t, "testfile\n", string(b))
	time.Sleep(4 * time.Second)
	assert.NoFileExists(t, "./test/tmp-mount/test")
}
