package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ()

func RemoveContents(dir string) error {
	fmt.Println("RemoveContents: ", dir)
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
	pwd, _ := os.Getwd()
	test_folder := strings.ReplaceAll(BACKUP_TEST_FOLDER, HOME, pwd)
	err := RemoveContents(test_folder)
	if err != nil {
		fmt.Println("Error cleaning up: ", err.Error())
	}
}

func TestBackupDoBackup(t *testing.T) {
	fmt.Println("running: TestBackupDoBackup")
	clear()
	t.Cleanup(clear)
	pwd, err := os.Getwd()
	require.NoError(t, err)

	test_folder := strings.ReplaceAll(BACKUP_TEST_FOLDER, HOME, pwd)
	test_exclude := strings.ReplaceAll(BACKUP_TEST_EXCLUDE_FILE, HOME, pwd)
	test_conf := strings.ReplaceAll(BACKUP_TEST_CONF_FILE, HOME, pwd)

	env := []string{
		RESTIC_PASSWORD + "test",
		RESTIC_REPOSITORY + test_folder,
	}

	err = os.MkdirAll(test_folder, os.ModePerm)
	assert.NoError(t, err)

	cmd := InitRepo(env, pwd)
	require.DirExists(t, test_folder)
	err = RunJob(cmd, "test", true)
	assert.NoError(t, err)

	cmd = Backup("~/", env, pwd, test_exclude, 2000, 2000)
	assert.Contains(t, cmd.String(), "restic backup ")
	assert.Contains(t, cmd.String(), pwd)
	assert.Contains(t, cmd.String(), "--exclude=\""+pwd+"/*.go\"")
	assert.Contains(t, cmd.String(), "--exclude=\""+pwd+"/test/exclude\"")
	assert.Contains(t, cmd.String(), "--limit-upload 2000")
	assert.Contains(t, cmd.String(), "--limit-download 2000")

	fmt.Println(cmd.String())
	err = RunJob(cmd, "backup", false)
	assert.NoError(t, err)
	err = IsEmpty(pwd, BACKUP_TEST_FOLDER)
	assert.NoError(t, err)
	assert.FileExists(t, test_conf)

}

func TestBackupExistsRepo(t *testing.T) {
	fmt.Println("running: TestBackupExistsRepo")
	t.Cleanup(clear)
	env := []string{
		RESTIC_PASSWORD + "hallo",
		RESTIC_REPOSITORY + BACKUP_TEST_FOLDER,
	}

	pwd, err := os.Getwd()
	require.NoError(t, err)

	test_folder := strings.ReplaceAll(BACKUP_TEST_FOLDER, HOME, pwd)
	test_conf := strings.ReplaceAll(BACKUP_TEST_CONF_FILE, HOME, pwd)

	err = os.MkdirAll(test_folder, os.ModePerm)
	assert.NoError(t, err)

	cmd := ExistsRepo(env, pwd)
	assert.Contains(t, cmd.String(), "restic snapshots")
	err = RunJob(cmd, "test", false)
	assert.Error(t, err)

	cmd = InitRepo(env, pwd)
	require.NoFileExists(t, test_conf)

	err = RunJob(cmd, "test", true)
	assert.NoError(t, err)

	cmd = ExistsRepo(env, pwd)
	err = RunJob(cmd, "test", false)
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
		RESTIC_REPOSITORY + BACKUP_TEST_FOLDER,
	}
	pwd, err := os.Getwd()
	require.NoError(t, err)

	test_folder := strings.ReplaceAll(BACKUP_TEST_FOLDER, HOME, pwd)

	err = os.MkdirAll(test_folder, os.ModePerm)
	cmd := InitRepo(env, pwd)
	require.NoFileExists(t, BACKUP_TEST_CONF_FILE)
	assert.Contains(t, cmd.String(), "restic init")

	require.DirExists(t, test_folder)
	err = RunJob(cmd, "test", false)
	assert.NoError(t, err)
}
