package main

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func MountFolders(config []GocryptConfig) []*exec.Cmd {
	var output []*exec.Cmd
	for i, folderconfig := range config {
		log.Println("Mounting: ", i, " with name: ", AbsolutePath(folderconfig.MountPoint), " Duration", folderconfig.MountDuration.String(), " AllowOther", folderconfig.AllowOther)
		cmd := MountGocryptfs(folderconfig.Path, folderconfig.MountPoint, folderconfig.MountDuration, folderconfig.Password, folderconfig.AllowOther, folderconfig.NotEmpty)
		empty, err := IsEmpty(folderconfig.MountPoint)
		if err != nil {
			log.Println("ERROR", err)
		} else {
			if !empty {
				log.Println(AbsolutePath(folderconfig.MountPoint), ": was not empty")
			} else {
				output = append(output, cmd)
			}
		}
	}
	return output
}

func MountGocryptfs(cryptoDir string, folder string, duration time.Duration, pwd string, allowOther bool, nonempty bool) *exec.Cmd {
	var cmd *exec.Cmd
	var command string

	command = "gocryptfs"
	if allowOther {
		command = command + " -allow_other"
	}
	if nonempty {
		command = command + " -nonempty"
	}
	if duration.String() != "0s" {
		command = command + " -i " + duration.String()
	}

	command = command + " " + AbsolutePath(cryptoDir) + " " + AbsolutePath(folder)
	cmd = exec.Command("bash", "-c", command)
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(pwd)
	return cmd
}

func AbsolutePath(path string) string {
	// from: https://stackoverflow.com/a/17617721/9447237
	dir, err := os.UserHomeDir()
	if strings.HasPrefix(path, "~/") && err == nil {
		path = strings.ReplaceAll(path, "~", dir)
	}
	return path
}

func IsEmpty(name string) (bool, error) {
	path := AbsolutePath(name)
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !stat.IsDir() {
		return false, errors.New("Not a folder")
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	//from: https://stackoverflow.com/a/30708914/9447237
	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
