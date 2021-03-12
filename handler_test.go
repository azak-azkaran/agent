package main

import (
	"fmt"
	"testing"
	"bytes"
	"os/exec"

	"github.com/stretchr/testify/assert"
)

func TestHandleMount(t *testing.T) {
	fmt.Println("running: TestHandleMount")

	cmd := exec.Command("return -1")
	job := CreateJobFromCommand(cmd, "test")
	var buffer bytes.Buffer
	assert.False(t,HandleMount(job, true, false, true, buffer))
}
