package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
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

	cmd := Backup("./test/Backup", "test", "./test/exclude", 2000, 2000)
	assert.Contains(t, cmd.String(), "restic backup ~/* ~/.* -x --exclude-file ./test/exclude --tag -o s3.connections=10 --limit-upload 2000 --limit-download 2000")
}

func TestExistsRepo(t *testing.T) {
	fmt.Println("running: TestExistsRepo")
	cmd := ExistsRepo("./test/Backup", "hallo")
	assert.Contains(t, cmd.String(), "restic snapshots")
	_, err := RunJob(cmd)
	assert.Error(t, err)

	cmd = InitRepo("./test/Backup", "hallo")
	require.NoFileExists(t, "./test/Backup/config")
	_, err = RunJob(cmd)
	assert.NoError(t, err)

	cmd = ExistsRepo("./test/Backup", "hallo")
	s, err := RunJob(cmd)
	assert.NoError(t, err)
	fmt.Println(s)
	RemoveContents("./test/Backup/")
	assert.NoFileExists(t, "./test/Backup/config")
}

func TestInitRepo(t *testing.T) {
	fmt.Println("running: TestInitRepo")
	cmd := InitRepo("./test/Backup", "hallo")
	require.NoFileExists(t, "./test/Backup/config")
	assert.Contains(t, cmd.String(), "restic init")

	require.DirExists(t, "./test/Backup")
	_, err := RunJob(cmd)
	assert.NoError(t, err)
	RemoveContents("./test/Backup/")
	assert.NoFileExists(t, "./test/Backup/config")
}
