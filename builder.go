package main

import (
	"fmt"
	"github.com/libgit2/git2go"
	"os"
	"path"
)

var GitEndpoint, GitWorkDir string

type Builder struct {
	name string
	git  *GitInfo
}

func (self *Builder) Build() (string, error) {
	repoUrl := path.Join(GitEndpoint, fmt.Sprintf("%s.git", self.git.Name))
	storePath := path.Join(GitWorkDir, self.name)
	defer os.RemoveAll(storePath)
	repo, err := git.Clone(repoUrl, storePath, &git.CloneOptions{})
	if err != nil {
		return "", err
	}
	return "", nil
}

func (self *Builder) CreateDockerFile() error {
	return nil
}

func (self *Builder) BuildImage() (string, error) {
	return "", nil
}

func (self *Builder) PushImage() error {
	return nil
}
