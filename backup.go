package main

import (
	"os"
	"os/exec"
	"strconv"
)

const (
	RESTIC_REPOSITORY = "RESTIC_REPOSITORY="
	RESTIC_PASSWORD   = "RESTIC_PASSWORD="
)

func InitRepo(repo string, pwd string) *exec.Cmd {
	//restic -r "repo" init
	cmd := exec.Command("restic", "init")
	cmd.Env = append(os.Environ(),
		RESTIC_REPOSITORY+repo,
		RESTIC_PASSWORD+pwd,
	)
	return cmd
}

func ExistsRepo(repo string, pwd string) *exec.Cmd {
	cmd := exec.Command("restic", "snapshots")
	cmd.Env = append(os.Environ(),
		RESTIC_REPOSITORY+repo,
		RESTIC_PASSWORD+pwd,
	)
	return cmd
}

func CheckRepo(repo string, pwd string) *exec.Cmd {
	cmd := exec.Command("restic", "check")
	cmd.Env = append(os.Environ(),
		RESTIC_REPOSITORY+repo,
		RESTIC_PASSWORD+pwd,
	)
	return cmd
}

func Backup(repo string, pwd string, excludeFile string, upload int, download int) *exec.Cmd {
	//restic --verbose backup ~/* ~/.* -x \
	//            --exclude-file ~/Documents/backup/exclude_home \
	//            --tag 'full-home' \
	//            -o s3.connections=10 --limit-upload 2000 --limit-download 2000
	cmd := exec.Command("restic",
		"backup", "~/*", "~/.*", "-x",
		"--exclude-file", excludeFile,
		"--tag", "-o", "s3.connections=10",
		"--limit-upload", strconv.Itoa(upload),
		"--limit-download", strconv.Itoa(download))

	cmd.Env = append(os.Environ(),
		RESTIC_REPOSITORY+repo,
		RESTIC_PASSWORD+pwd,
	)
	//cmd.Stdin = strings.NewReader(pwd)
	return cmd
}
