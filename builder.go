package main

import (
	"fmt"
	"github.com/libgit2/git2go"
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
	repoUrl := path.Join(GitEndpoint, self.git.Group, fmt.Sprintf("%s.git", self.git.Name))
	storePath := path.Join(self.workdir, self.name)
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
	filePath := path.Join(self.workdir, "Dockerfile")
	codePath := path.Join(self.workdir, self.name)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("FROM %s\n\n", self.git.Base))
	f.WriteString(fmt.Sprintf("ADD %s /%s\n\n", codePath, self.name))

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
