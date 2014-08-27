package main

import (
	"fmt"
	"github.com/libgit2/git2go"
	"io/ioutil"
	"os"
	"path"
)

var GitEndpoint, GitWorkDir string

type Builder struct {
	name    string
	workdir string
	git     *GitInfo
}

func (self *Builder) FetchCode() error {
	repoUrl := path.Join(GitEndpoint, fmt.Sprintf("%s.git", self.git.Name))
	storePath := path.Join(self.workdir, self.git.Name)
	repo, err := git.Clone(repoUrl, storePath, &git.CloneOptions{})
	if err != nil {
		return err
	}

	opts := &git.CheckoutOpts{
		Strategy: git.CheckoutForce,
	}

	err = repo.CheckoutTree(self.git.Version, opts)
	if err != nil {
		return err
	}

	return os.RemoveAll(path.Join(storePath, ".git"))
}

func (self *Builder) CreateDockerFile() error {
	return nil
}

func (self *Builder) CleanSourceCode() error {
	return os.RemoveAll(self.workdir)
}

func (self *Builder) BuildImage() (string, error) {
	return "", nil
}

func (self *Builder) PushImage() error {
	return nil
}
