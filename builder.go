package main

import (
	"fmt"
	"github.com/CMGS/go-dockerclient"
	"github.com/juju/utils/tar"
	"github.com/libgit2/git2go"
	"os"
	"path"
)

type Builder struct {
	name           string
	workDir        string
	codeDir        string
	dockerFilePath string
	tarPath        string
	repoURL        string
	registryURL    string
	repoTag        string
	build          *BuildInfo
}

func NewBuilder(name string, build *BuildInfo) *Builder {
	builder := Builder{name: name, build: build}
	builder.workDir = path.Join(config.Git.WorkDir, name, build.Version)
	builder.repoURL = UrlJoin(config.Git.Endpoint, build.Group, fmt.Sprintf("%s.git", build.Name))
	builder.codeDir = path.Join(builder.workDir, name)
	builder.dockerFilePath = path.Join(builder.workDir, "Dockerfile")
	builder.tarPath = path.Join(builder.workDir, fmt.Sprintf("%s.tar.gz", name))
	builder.registryURL = UrlJoin(config.Docker.Registry, name)
	builder.repoTag = fmt.Sprintf("%s:%s", builder.registryURL, build.Version)
	return &builder
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

func (self *Builder) Build() error {
	defer self.clear()

	if err := self.fetchCode(); err != nil {
		return err
	}
	if err := self.createDockerFile(); err != nil {
		return err
	}
	if err := self.createTar(); err != nil {
		return err
	}
	if err := self.buildImage(); err != nil {
		return err
	}
	if err := self.pushImage(); err != nil {
		return err
	}
	return nil
}

func (self *Builder) fetchCode() error {
	repo, err := git.Clone(self.repoURL, self.codeDir, &git.CloneOptions{})
	logger.Debug(self.repoURL, self.codeDir)
	if err != nil {
		return err
	}

	opts := &git.CheckoutOpts{Strategy: git.CheckoutForce}
	if err := self.checkout(repo, opts); err != nil {
		return err
	}

	return os.RemoveAll(path.Join(self.codeDir, ".git"))
}

func (self *Builder) createDockerFile() error {
	logger.Debug(self.dockerFilePath)
	f, err := os.Create(self.dockerFilePath)
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

func (self *Builder) createTar() error {
	logger.Debug(self.tarPath)

	file, _ := os.Create(self.tarPath)
	defer file.Close()
	if _, err := tar.TarFiles([]string{self.dockerFilePath, self.codeDir}, file, self.workDir); err != nil {
		return err
	}
	return nil
}

func (self *Builder) buildImage() error {
	file, err := os.Open(self.tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	opts := docker.BuildImageOptions{
		Name:                self.repoTag,
		NoCache:             false,
		SuppressOutput:      true,
		RmTmpContainer:      true,
		ForceRmTmpContainer: true,
		InputStream:         file,
		OutputStream:        GetBuffer(),
	}

	if err := Docker.BuildImage(opts); err != nil {
		return err
	}
	return nil
}

func (self *Builder) pushImage() error {
	url := UrlJoin(config.Docker.Registry, self.name)
	if err := Docker.PushImage(
		docker.PushImageOptions{url, self.build.Version, config.Docker.Registry, GetBuffer()},
		docker.AuthConfiguration{}); err != nil {
		return err
	}
	return nil
}

func (self *Builder) clear() {
	defer os.RemoveAll(self.workDir)
	images, err := Docker.ListImages(false)
	if err != nil {
		logger.Debug(err)
	}
	for _, image := range images {
		flag := false
		for _, tag := range image.RepoTags {
			if tag == "<none>:<none>" || tag == self.repoTag {
				flag = true
			}
		}
		if flag {
			logger.Debug(image)
			Docker.RemoveImage(image.ID)
		}
	}
}
