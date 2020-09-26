package main

import (
	"errors"
	"log"
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func GitClone(repo string, dir string, home string, personal string) error {
	var r *git.Repository
	var err error

	dir = strings.ReplaceAll(dir, HOME, home)

	log.Println("Checkout out to Dir: ", dir)
	log.Println("Checkout out to Repo: ", repo)
	if personal == "" {
		r, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL:      repo,
			Progress: os.Stdout,
		})
	} else {

		r, err = git.PlainClone(dir, false, &git.CloneOptions{
			// The intended use of a GitHub personal access token is in replace of your password
			// because access tokens can easily be revoked.
			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
			Auth: &http.BasicAuth{
				Username: "abc123", // yes, this can be anything except an empty string
				Password: personal,
			},
			URL:      repo,
			Progress: os.Stdout,
		})

	}
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

func GitPull(repo string, dir string, home string, personal string) error {
	return errors.New("Function not implemented")
}
