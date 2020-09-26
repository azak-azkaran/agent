package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobAddJob(t *testing.T) {
	fmt.Println("running: TestJobAddJob")
	cmd := exec.Command("echo", "hallo")
	CreateJobFromCommand(cmd, "test")
	assert.NotNil(t, jobmap)
	assert.NotEmpty(t, jobmap)
}

func TestJobRunJobBackground(t *testing.T) {
	fmt.Println("running: TestJobRunJobBackground")
	cmd := exec.Command("echo", "hallo")

	job := CreateJobFromCommand(cmd, "test")
	err := job.RunJobBackground(false)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		v, ok := jobmap.Get("test")
		require.True(t, ok)
		job := v.(Job)
		return job.Cmd.Process != nil
	}, time.Duration(4*time.Second), time.Duration(1*time.Second))
	v, ok := jobmap.Get("test")
	require.True(t, ok)
	job = v.(Job)

	assert.Equal(t, "hallo\n", job.Stdout.String())
	assert.Equal(t, "", job.Stderr.String())
}

func TestJobRunJob(t *testing.T) {
	fmt.Println("running: TestJobRunJob")
	cmd := exec.Command("echo", "hallo")

	job := CreateJobFromCommand(cmd, "test")
	err := job.RunJob(false)
	assert.NoError(t, err)

	v, ok := jobmap.Get("test")
	require.True(t, ok)
	job = v.(Job)
	assert.Equal(t, "hallo\n", job.Stdout.String())
	assert.Equal(t, "", job.Stderr.String())

	cmd = exec.Command("printenv")
	cmd.Env = []string{"TEST=hallo"}

	job = CreateJobFromCommand(cmd, "test")
	err = job.RunJob(true)
	assert.NoError(t, err)

	v, ok = jobmap.Get("test")
	require.True(t, ok)
	job = v.(Job)
	assert.Equal(t, "TEST=hallo\n", job.Stdout.String())
	assert.Equal(t, "", job.Stderr.String())
}

func TestJobQueueStatus(t *testing.T) {
	fmt.Println("running: TestJobQueueStatus")

	cmd := exec.Command("echo", "hallo")

	job := CreateJobFromCommand(cmd, "test1")
	var infoBuffer bytes.Buffer

	log.SetOutput(&infoBuffer)
	job.QueueStatus()

	assert.NotEmpty(t, infoBuffer)

	log.SetOutput(os.Stdout)
	log.Println("test: ", infoBuffer.String())
}
