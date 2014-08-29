package main

import (
	"bytes"
	"fmt"
	"github.com/CMGS/go-dockerclient"
	"github.com/juju/utils/tar"
	"github.com/libgit2/git2go"
	"os"
	"path"
)

var GitEndpoint, GitWorkDir string

type Builder struct {
	name    string
	workdir string
	build   *BuildInfo
}

func (self *Builder) checkout(repo *git.Repository, opts *git.CheckoutOpts) error {
	obj, err := repo.RevparseSingle(self.build.Version)
	if err != nil {
		return err
	}
	defer obj.Free()

	commit, err := repo.LookupCommit(obj.Id())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}
	if err := repo.CheckoutTree(tree, opts); err != nil {
		return err
	}
	return nil
}

func (self *Builder) FetchCode() error {
	repoUrl := UrlJoin(GitEndpoint, self.build.Group, fmt.Sprintf("%s.git", self.build.Name))
	storePath := path.Join(self.workdir, self.name)
	repo, err := git.Clone(repoUrl, storePath, &git.CloneOptions{})
	logger.Debug(repoUrl, storePath)
	if err != nil {
		return err
	}

	opts := &git.CheckoutOpts{Strategy: git.CheckoutForce}
	if err := self.checkout(repo, opts); err != nil {
		return err
	}

	return os.RemoveAll(path.Join(storePath, ".git"))
}

func (self *Builder) CreateDockerFile() error {
	filePath := path.Join(self.workdir, "Dockerfile")
	logger.Debug(filePath)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("FROM %s\n\n", self.build.Base))
	f.WriteString("ENV NBE 1\n")
	f.WriteString(fmt.Sprintf("ADD %s /%s\n", self.name, self.name))
	f.WriteString(fmt.Sprintf("WORKDIR /%s\n", self.name))
	f.WriteString(fmt.Sprintf("RUN %s\n", self.build.Build))

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}

func (self *Builder) CreateTar() error {
	tarPath := path.Join(self.workdir, fmt.Sprintf("%s.tar.gz", self.name))
	filePath := path.Join(self.workdir, "Dockerfile")
	codePath := path.Join(self.workdir, self.name)
	logger.Debug(tarPath)

	file, _ := os.Create(tarPath)
	defer file.Close()
	if _, err := tar.TarFiles([]string{filePath, codePath}, file, self.workdir); err != nil {
		return err
	}
	return nil
}

func (self *Builder) Clear() {
	defer os.RemoveAll(self.workdir)
	images, err := Docker.ListImages(false)
	if err != nil {
		logger.Debug(err)
	}
	for _, image := range images {
		flag := false
		for _, tag := range image.RepoTags {
			if tag == "<none>:<none>" {
				flag = true
			}
		}
		if flag {
			logger.Debug(image)
			Docker.RemoveImage(image.ID)
		}
	}
}

func (self *Builder) BuildImage() error {
	tarPath := path.Join(self.workdir, fmt.Sprintf("%s.tar.gz", self.name))
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	name := fmt.Sprintf("%s:%s", UrlJoin(RegEndpoint, self.name), self.build.Version)

	opts := docker.BuildImageOptions{
		Name:                name,
		NoCache:             false,
		SuppressOutput:      true,
		RmTmpContainer:      true,
		ForceRmTmpContainer: true,
		InputStream:         file,
		OutputStream:        os.Stdout,
	}

	if err := client.BuildImage(opts); err != nil {
		return err
	}
	return nil
}

func (self *Builder) PushImage() error {
	url := UrlJoin(RegEndpoint, self.name)
	buf := bytes.Buffer{}
	defer logger.Debug(buf.String())
	if err := Docker.PushImage(docker.PushImageOptions{url, self.build.Version, RegEndpoint, &buf}, nil); err != nil {
		return err
	}
	return nil
}
