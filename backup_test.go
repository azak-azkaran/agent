package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	BACKUP_FOLDER    = "./test/Backup"
	EXCLUDE_FILE     = "./test/exclude"
	BACKUP_CONF_FILE = "./test/Backup/config"
)

func RemoveContents(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func clear() {
	err := RemoveContents(BACKUP_FOLDER)
	if err != nil {
		fmt.Println("Error cleaning up: ", err.Error())
	}
}

func TestBackup(t *testing.T) {
	fmt.Println("running: TestBackup")
	t.Cleanup(clear)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	cmd := Backup(pwd, BACKUP_FOLDER, "test", EXCLUDE_FILE, 2000, 2000)
	assert.Contains(t, cmd.String(), "restic backup "+
		pwd+" "+
		"-x --exclude-file "+EXCLUDE_FILE+" --tag -o s3.connections=10 --limit-upload 2000 --limit-download 2000")
}

func TestExistsRepo(t *testing.T) {
	fmt.Println("running: TestExistsRepo")
	t.Cleanup(clear)
	cmd := ExistsRepo(BACKUP_FOLDER, "hallo")
	assert.Contains(t, cmd.String(), "restic snapshots")
	err := RunJob(cmd, "test")
	assert.Error(t, err)

	cmd = InitRepo(BACKUP_FOLDER, "hallo")
	require.NoFileExists(t, BACKUP_CONF_FILE)
	err = RunJob(cmd, "test")
	assert.NoError(t, err)

	cmd = ExistsRepo(BACKUP_FOLDER, "hallo")
	err = RunJob(cmd, "test")
	assert.NoError(t, err)

	v, ok := jobmap.Get("test")
	require.True(t, ok)
	job := v.(Job)
	fmt.Println(job.Stdout.String())
}

func TestInitRepo(t *testing.T) {
	fmt.Println("running: TestInitRepo")
	t.Cleanup(clear)
	cmd := InitRepo(BACKUP_FOLDER, "hallo")
	require.NoFileExists(t, BACKUP_CONF_FILE)
	assert.Contains(t, cmd.String(), "restic init")

	require.DirExists(t, BACKUP_FOLDER)
	err := RunJob(cmd, "test")
	assert.NoError(t, err)
}
