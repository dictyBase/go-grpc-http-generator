package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-github/v28/github"

	"github.com/urfave/cli"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func cloneGitRepo(repo, branch string, isTag bool) (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "gclone")
	if err != nil {
		return "", fmt.Errorf("error in creating temp dir %s", err)
	}
	cloneOpt := &git.CloneOptions{
		URL:           repo,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	}
	if isTag {
		cloneOpt.ReferenceName = plumbing.NewTagReferenceName(branch)
	}
	_, err = git.PlainClone(dir, false, cloneOpt)
	return dir, err
}

func getValidatorProto(c *cli.Context) (string, error) {
	dir, out, err := createValidatorTemp("validator", c.String("validator-proto-local-path"))
	if err != nil {
		return "", err
	}
	client := github.NewClient(nil)
	ghf, _, _, err := client.Repositories.GetContents(
		context.Background(),
		c.String("validator-repo-owner"),
		c.String("validator-repo-name"),
		c.String("validator-proto-path"),
		&github.RepositoryContentGetOptions{
			Ref: c.String("validator-repo-version"),
		})
	if err != nil {
		return "", fmt.Errorf("error in downloading file from given path %s", err)
	}
	ct, err := ghf.GetContent()
	if err != nil {
		return "", fmt.Errorf("error in getting content from file in repo %s", err)
	}
	out = filepath.Join(out, c.String("validator-proto-path"))
	if err := ioutil.WriteFile(out, []byte(ct), 0644); err != nil {
		return "", fmt.Errorf("error in writing content to file %s %s", out, err)
	}
	return dir, nil
}
