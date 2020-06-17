package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func Backup(path string, env []string, exclude string, upload int, download int) *exec.Cmd {
	var bud strings.Builder
	excludes := strings.Split(exclude, "\n")

	bud.WriteString("restic backup ")
	bud.WriteString(path)
	bud.WriteString(" -x ")
	for _, v := range excludes {
		bud.WriteString(" --exclude=\"")
		bud.WriteString(v)
		bud.WriteString("\"")
	}
	bud.WriteString(" --tag 'full-home'")
	bud.WriteString(" --limit-upload ")
	bud.WriteString(strconv.Itoa(upload))
	bud.WriteString(" --limit-download ")
	bud.WriteString(strconv.Itoa(download))
	command := bud.String()
	//" --quiet "

	return createCmd(command, env)
}
