package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitClone(t *testing.T) {
	fmt.Println("running: TestGitClone")
	t.Cleanup(clear)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	test_folder := strings.ReplaceAll(GIT_TEST_FOLDER, HOME, pwd)
	require.NoDirExists(t, test_folder)

	err = GitClone(GIT_TEST_REPO, GIT_TEST_FOLDER, pwd, "")
	assert.NoError(t, err)
	assert.DirExists(t, test_folder)
}

func TestGitPull(t *testing.T) {
	fmt.Println("running: TestGitPull")
	t.Cleanup(clear)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	test_folder := strings.ReplaceAll(GIT_TEST_FOLDER, HOME, pwd)
	require.NoDirExists(t, test_folder)

	err = GitClone(GIT_TEST_REPO, GIT_TEST_FOLDER, pwd, "")
	assert.NoError(t, err)
	assert.DirExists(t, test_folder)

	err = GitPull(GIT_TEST_FOLDER, pwd)
	assert.NoError(t, err)
}
