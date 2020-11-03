package main

import (
	"log"
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func GitClone(repo string, dir string, home string, personal string) error {
	var r *git.Repository
	var err error

	dir = strings.ReplaceAll(dir, HOME, home)
	log.Println("Checkout out to Repo: ", repo)
	log.Println("Checkout out to Dir: ", dir)

	cloneOptions := git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
	}

	if personal != "" {
		cloneOptions.Auth = &http.BasicAuth{
			// The intended use of a GitHub personal access token is in replace of your password
			// because access tokens can easily be revoked.
			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
			Username: "abc123", // yes, this can be anything except an empty string
			Password: personal,
		}
	}

	r, err = git.PlainClone(dir, false, &cloneOptions)
	if err != nil {
		return err
	}

	ref, err := r.Head()
	if err != nil {
		return err
	}
	log.Println("Checkout out Ref: ", ref)
	return nil
}

func GitCreateRemote(dir string, home string, repoUrl string) error {
	path := strings.ReplaceAll(dir, HOME, home)
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}
	_, err = r.Remote(GIT_REMOTE_NAME)
	if err != nil && err == git.ErrRemoteNotFound {
		log.Println("Adding remote: ", GIT_REMOTE_NAME)
		_, err = r.CreateRemote(&config.RemoteConfig{
			Name: GIT_REMOTE_NAME,
			URLs: []string{repoUrl},
		})
		return err
	}
	return err
}

func GitPull(dir string, home string, personal string) error {
	path := strings.ReplaceAll(dir, HOME, home)
	log.Println("Pulling from: ", path)
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	pullOptions := git.PullOptions{
		RemoteName: GIT_REMOTE_NAME,
	}

	if personal != "" {
		pullOptions.Auth = &http.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			Password: personal,
		}
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.Pull(&pullOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	// Print the latest commit that was just pulled
	ref, err := r.Head()
	if err != nil {
		return err
	}
	log.Println("Checkout out Ref: ", ref)
	return nil
}
