package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func createCmd(command string, env []string, home string) *exec.Cmd {
	//https://stackoverflow.com/a/43246464/9447237
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = os.Environ()

	for _, v := range env {
		if strings.Contains(v, HOME) {
			cmd.Env = append(cmd.Env, strings.ReplaceAll(v, HOME, home))
		} else {
			cmd.Env = append(cmd.Env, v)
		}
	}
	return cmd
}

func InitRepo(env []string, home string) *exec.Cmd {
	return createCmd("restic init", env, home)
}

func ExistsRepo(env []string, home string) *exec.Cmd {
	return createCmd("restic snapshots", env, home)
}

func CheckRepo(env []string, home string) *exec.Cmd {
	return createCmd("restic check", env, home)
}

func UnlockRepo(env []string, home string) *exec.Cmd {
	return createCmd("restic unlock", env, home)
}

func ListRepo(env []string, home string) *exec.Cmd {
	return createCmd("restic snapshots", env, home)
}

func ForgetRepoDetail(env []string, home string, daily int, monthly int, yearly int) *exec.Cmd {
	var bud strings.Builder
	bud.WriteString("restic forget --prune --keep-daily ")
	bud.WriteString(strconv.Itoa(daily))
	bud.WriteString(" --keep-monthly ")
	bud.WriteString(strconv.Itoa(monthly))
	bud.WriteString(" --keep-yearly ")
	bud.WriteString(strconv.Itoa(yearly))
	bud.WriteString(" --prune")
	return createCmd(bud.String(), env, home)
}

func ForgetRep(env []string, home string) *exec.Cmd {
	return ForgetRepoDetail(env, home, 7, 12, 3)
}

func Backup(path string, env []string, home string, exclude string, upload int, download int) *exec.Cmd {
	var bud strings.Builder

	//test_mountpath := strings.ReplaceAll(GOCRYPT_TEST_MOUNTPATH, "~", home)
	path = strings.ReplaceAll(path, HOME, home)
	exclude = strings.ReplaceAll(exclude, HOME, home)

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
	return createCmd(command, env, home)
}
