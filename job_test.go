package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobAddJob(t *testing.T) {
	fmt.Println("running: TestJobAddJob")
	t.Cleanup(clear)
	cmd := exec.Command("echo", "hallo")
	CreateJobFromCommand(cmd, "test")
	assert.NotNil(t, jobmap)
	assert.NotEmpty(t, jobmap)
}

func TestJobRunJobBackground(t *testing.T) {
	fmt.Println("running: TestJobRunJobBackground")
	t.Cleanup(clear)
	cmd := exec.Command("echo", "hallo")

	job := CreateJobFromCommand(cmd, "test")
	err := job.RunJobBackground(false)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		v, ok := jobmap.Get("test")
		require.True(t, ok)
		j := v.(*Job)
		return j.Cmd.Process != nil
	}, time.Duration(4*time.Second), time.Duration(1*time.Second))
	v, ok := jobmap.Get("test")
	require.True(t, ok)
	j := v.(*Job)

	assert.Equal(t, "hallo\n", j.Stdout.String())
	assert.Equal(t, "", j.Stderr.String())
}

func TestJobRunJob(t *testing.T) {
	fmt.Println("running: TestJobRunJob")
	t.Cleanup(clear)
	cmd := exec.Command("echo", "hallo")

	job := CreateJobFromCommand(cmd, "test")
	err := job.RunJob(false)
	assert.NoError(t, err)

	v, ok := jobmap.Get("test")
	require.True(t, ok)
	j := v.(*Job)
	assert.Equal(t, "hallo\n", j.Stdout.String())
	assert.Equal(t, "", j.Stderr.String())

	cmd = exec.Command("printenv")
	cmd.Env = []string{"TEST=hallo"}

	job = CreateJobFromCommand(cmd, "test")
	err = job.RunJob(true)
	assert.NoError(t, err)

	v, ok = jobmap.Get("test")
	require.True(t, ok)
	j = v.(*Job)
	assert.Equal(t, "TEST=hallo\n", j.Stdout.String())
	assert.Equal(t, "", j.Stderr.String())
}

func TestJobQueueStatus(t *testing.T) {
	fmt.Println("running: TestJobQueueStatus")
	t.Cleanup(clear)

	cmd := exec.Command("echo", "hallo")

	job := CreateJobFromCommand(cmd, "test1")
	var infoBuffer bytes.Buffer

	GetLogger().SetOutput(&infoBuffer)
	job.QueueStatus()

	assert.NotEmpty(t, infoBuffer)

	GetLogger().SetOutput(os.Stdout)
	Sugar.Info("test: ", infoBuffer.String())
}
