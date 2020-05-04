package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func MountGocryptfs(cryptoDir string, folder string, duration time.Duration, pwd string, allowOther bool) *exec.Cmd {
	var cmd *exec.Cmd
	if allowOther {
		cmd = exec.Command("gocryptfs", "-allow_other", "-i", duration.String(), cryptoDir, folder)
	} else {
		cmd = exec.Command("gocryptfs", "-i", duration.String(), cryptoDir, folder)
	}
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(pwd)
	return cmd
}

func IsEmpty(name string) (bool, error) {
	//from: https://stackoverflow.com/a/30708914/9447237
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
