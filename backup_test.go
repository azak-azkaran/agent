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

func TestBackupDoBackup(t *testing.T) {
	fmt.Println("running: TestBackupDoBackup")
	t.Cleanup(clear)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	env := []string{
		RESTIC_PASSWORD + "test",
		RESTIC_REPOSITORY + BACKUP_FOLDER,
	}

	cmd := InitRepo(env)
	require.DirExists(t, BACKUP_FOLDER)
	err = RunJob(cmd, "test")
	assert.NoError(t, err)

	cmd = Backup(pwd, env, EXCLUDE_FILE, 2000, 2000)
	assert.Contains(t, cmd.String(), "restic backup ")
	assert.Contains(t, cmd.String(), pwd)
	assert.Contains(t, cmd.String(), "--exclude-file "+EXCLUDE_FILE)
	assert.Contains(t, cmd.String(), "--tag -o s3.connections=10")
	assert.Contains(t, cmd.String(), "--limit-upload 2000")
	assert.Contains(t, cmd.String(), "--limit-download 2000")

	err = RunJob(cmd, "backup")
	assert.NoError(t, err)
	is, err := IsEmpty(BACKUP_FOLDER)
	assert.NoError(t, err)
	assert.False(t, is)
	assert.FileExists(t, "./test/Backup/config")

}

func TestBackupExistsRepo(t *testing.T) {
	fmt.Println("running: TestBackupExistsRepo")
	t.Cleanup(clear)
	env := []string{
		RESTIC_PASSWORD + "hallo",
		RESTIC_REPOSITORY + BACKUP_FOLDER,
	}
	cmd := ExistsRepo(env)
	assert.Contains(t, cmd.String(), "restic snapshots")
	err := RunJob(cmd, "test")
	assert.Error(t, err)

	cmd = InitRepo(env)
	require.NoFileExists(t, BACKUP_CONF_FILE)
	err = RunJob(cmd, "test")
	assert.NoError(t, err)

	cmd = ExistsRepo(env)
	err = RunJob(cmd, "test")
	assert.NoError(t, err)

	v, ok := jobmap.Get("test")
	require.True(t, ok)
	job := v.(Job)
	fmt.Println(job.Stdout.String())
}

func TestBackupInitRepo(t *testing.T) {
	fmt.Println("running: TestBackupInitRepo")
	t.Cleanup(clear)
	env := []string{
		RESTIC_PASSWORD + "hallo",
		RESTIC_REPOSITORY + BACKUP_FOLDER,
	}
	cmd := InitRepo(env)
	require.NoFileExists(t, BACKUP_CONF_FILE)
	assert.Contains(t, cmd.String(), "restic init")
	assert.Contains(t, cmd.Env, RESTIC_REPOSITORY+BACKUP_FOLDER)

	require.DirExists(t, BACKUP_FOLDER)
	err := RunJob(cmd, "test")
	assert.NoError(t, err)
}
