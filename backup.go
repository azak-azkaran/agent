package main

import (
	"os"
	"os/exec"
	"strconv"
)

func createCmd(command string, env []string) *exec.Cmd {
	//https://stackoverflow.com/a/43246464/9447237
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)
	return cmd
}

func InitRepo(env []string) *exec.Cmd {
	return createCmd("restic init", env)
}

func ExistsRepo(env []string) *exec.Cmd {
	return createCmd("restic snapshots", env)
}

func CheckRepo(env []string) *exec.Cmd {
	return createCmd("restic check", env)
}

func Backup(path string, env []string, excludeFile string, upload int, download int) *exec.Cmd {
	command := "restic backup " + path + " -x " +
		" --exclude-file " + excludeFile +
		" --tag 'full-home' " + //"-o s3.connections=10" +
		" --limit-upload " + strconv.Itoa(upload) +
		" --limit-download " + strconv.Itoa(download) //+
	//" --quiet "

	return createCmd(command, env)
}
