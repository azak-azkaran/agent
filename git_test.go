package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

	_, err = git.PlainOpen(test_folder)
	assert.NoError(t, err)

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

	ref, err := r.Head()
	require.NoError(t, err)
	head := ref.Hash().String()

	remote, err := r.Remote(GIT_REMOTE_NAME)
	assert.Error(t, git.ErrRemoteNotFound, err)
	assert.Nil(t, remote)

	w, err := r.Worktree()
	assert.NoError(t, err)
	assert.NotNil(t, w)

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(GIT_TEST_COMMIT),
	})
	assert.NoError(t, err)
	ref, err = r.Head()
	assert.NoError(t, err)
	assert.Equal(t, GIT_TEST_COMMIT, ref.Hash().String())

	err = GitPull(GIT_TEST_FOLDER, pwd, "")
	assert.Error(t, git.ErrRemoteNotFound, err)
	err = GitCreateRemote(GIT_TEST_FOLDER, pwd, GIT_TEST_REPO)
	assert.NoError(t, err)

	err = GitPull(GIT_TEST_FOLDER, pwd, "")
	assert.NoError(t, err)
	remote, err = r.Remote(GIT_REMOTE_NAME)
	assert.NoError(t, err)
	assert.NotNil(t, remote)

	ref, err = r.Head()
	assert.NoError(t, err)
	assert.Equal(t, head, ref.Hash().String())
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
