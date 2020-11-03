package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
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

	// Second Clone for test if repo exists error is ignored
	err = GitClone(GIT_TEST_REPO, GIT_TEST_FOLDER, pwd, "test")
	require.Error(t, err)
	assert.Error(t, git.ErrRepositoryAlreadyExists, err)
}

func TestGitPull(t *testing.T) {
	fmt.Println("running: TestGitPull")
	t.Cleanup(clear)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	test_folder := strings.ReplaceAll(GIT_TEST_FOLDER, HOME, pwd)
	require.NoDirExists(t, test_folder)

	r, err := git.PlainClone(test_folder, false, &git.CloneOptions{
		URL: GIT_TEST_REPO,
	})
	require.NoError(t, err)
	require.DirExists(t, test_folder)
	remote, err := r.Remote(GIT_REMOTE_NAME)
	assert.Error(t, git.ErrRemoteNotFound, err)
	assert.Nil(t, remote)

	err = GitPull(GIT_TEST_FOLDER, pwd, "")
	assert.Error(t, git.ErrRemoteNotFound, err)
	err = GitCreateRemote(GIT_TEST_FOLDER, pwd, GIT_TEST_REPO)
	assert.NoError(t, err)

	err = GitPull(GIT_TEST_FOLDER, pwd, "")
	assert.NoError(t, err)
	remote, err = r.Remote(GIT_REMOTE_NAME)
	assert.NoError(t, err)
	assert.NotNil(t, remote)
}

func TestGitCreateRemote(t *testing.T) {
	fmt.Println("running: TestGitCreateRemote")
	t.Cleanup(clear)

	t.Cleanup(clear)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	test_folder := strings.ReplaceAll(GIT_TEST_FOLDER, HOME, pwd)
	require.NoDirExists(t, test_folder)

	r, err := git.PlainClone(test_folder, false, &git.CloneOptions{
		URL: GIT_TEST_REPO,
	})
	require.NoError(t, err)
	require.DirExists(t, test_folder)

	remote, err := r.Remote(GIT_REMOTE_NAME)
	assert.Error(t, git.ErrRemoteNotFound, err)
	assert.Nil(t, remote)

	err = GitCreateRemote(GIT_TEST_FOLDER, pwd, GIT_TEST_REPO)
	assert.NoError(t, err)

	remote, err = r.Remote(GIT_REMOTE_NAME)
	assert.NoError(t, err)
	assert.NotNil(t, remote)
}
