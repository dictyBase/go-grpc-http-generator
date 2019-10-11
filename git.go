package main

import (
	"fmt"
	"io/ioutil"
	"os"

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
