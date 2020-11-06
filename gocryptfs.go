package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func MountFolders(home string, config []GocryptConfig) []*exec.Cmd {
	var output []*exec.Cmd
	for _, folderconfig := range config {

		cmd := mount(home, folderconfig)
		err := IsEmpty(home, folderconfig.MountPoint)
		if err != nil {
			Sugar.Error("ERROR", err)
		} else {
			output = append(output, cmd)
		}
	}
	return output
}

func mount(home string, folderconfig GocryptConfig) *exec.Cmd {
	return MountGocryptfs(folderconfig.Path, folderconfig.MountPoint, home, folderconfig.MountDuration, folderconfig.Password, folderconfig.AllowOther)
}

func MountGocryptfs(cryptoDir string, folder string, home string, duration time.Duration, pwd string, allowOther bool) *exec.Cmd {
	var cmd *exec.Cmd
	var command string

	command = "gocryptfs"
	if allowOther {
		command = command + " -allow_other"
	}
	if duration.String() != "0s" {
		command = command + " -i " + duration.String()
	}

	cryptoDir = strings.ReplaceAll(cryptoDir, HOME, home)
	folder = strings.ReplaceAll(folder, HOME, home)

	command = command + " " + cryptoDir + " " + folder

	Sugar.Debug("Mounting: ", folder, " Duration", duration.String(), " AllowOther", allowOther)
	cmd = exec.Command("bash", "-c", command)
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(pwd)
	return cmd
}

func IsEmpty(home string, name string) error {
	path := strings.ReplaceAll(name, HOME, home)
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("Not a folder")
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	//from: https://stackoverflow.com/a/30708914/9447237
	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return errors.New(path + ": was not empty") // Either not empty or error, suits both cases
	}
	return err
}
