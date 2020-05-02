package main

import (
	vault "github.com/hashicorp/vault/api"
	//badger "github.com/dgraph-io/badger/v2"
	"github.com/robfig/cron/v3"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func RunJob(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func MountGocryptfs(cryptoDir string, folder string, duration time.Duration, pwd string, allow_other bool) *exec.Cmd {
	var cmd *exec.Cmd
	if allow_other {
		cmd = exec.Command("gocryptfs", "-allow_other", "-i", duration.String(), cryptoDir, folder)
	} else {
		cmd = exec.Command("gocryptfs", "-i", duration.String(), cryptoDir, folder)
	}
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(pwd)
	return cmd
}

func main() {
	c := cron.New()
	c.Stop() // Stop the scheduler (does not stop any jobs already running).

	log.Print("Please enter Token: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		log.Fatal(err)
	}
	token := strings.TrimSpace(string(password))
	log.Printf("token %s", token)

	resp, err := IsSealed(vault.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Check if Vault is unsealed")
	if resp {
		log.Println("Vault is sealed")
	}
}
