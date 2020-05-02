package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

const (
	BACKUP_FOLDER = "./test/Backup"
	EXCLUDE_FILE  = "./test/exclude"
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

func TestBackup(t *testing.T) {
	fmt.Println("running: TestBackup")

	cmd := Backup(BACKUP_FOLDER, "test", EXCLUDE_FILE, 2000, 2000)
	assert.Contains(t, cmd.String(), "restic backup ~/* ~/.* -x --exclude-file "+EXCLUDE_FILE+" --tag -o s3.connections=10 --limit-upload 2000 --limit-download 2000")
}

func TestExistsRepo(t *testing.T) {
	fmt.Println("running: TestExistsRepo")
	cmd := ExistsRepo(BACKUP_FOLDER, "hallo")
	assert.Contains(t, cmd.String(), "restic snapshots")
	_, err := RunJob(cmd)
	assert.Error(t, err)

	cmd = InitRepo(BACKUP_FOLDER, "hallo")
	require.NoFileExists(t, BACKUP_FOLDER+"/config")
	_, err = RunJob(cmd)
	assert.NoError(t, err)

	cmd = ExistsRepo(BACKUP_FOLDER, "hallo")
	s, err := RunJob(cmd)
	assert.NoError(t, err)
	fmt.Println(s)
	RemoveContents(BACKUP_FOLDER)
	assert.NoFileExists(t, BACKUP_FOLDER+"/config")
}

func TestInitRepo(t *testing.T) {
	fmt.Println("running: TestInitRepo")
	cmd := InitRepo(BACKUP_FOLDER, "hallo")
	require.NoFileExists(t, BACKUP_FOLDER+"/config")
	assert.Contains(t, cmd.String(), "restic init")

	require.DirExists(t, BACKUP_FOLDER)
	_, err := RunJob(cmd)
	assert.NoError(t, err)
	RemoveContents(BACKUP_FOLDER)
	assert.NoFileExists(t, BACKUP_FOLDER+"/config")
}
