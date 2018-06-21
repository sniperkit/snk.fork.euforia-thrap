package asm

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/euforia/hclencoder"
	"github.com/euforia/pseudo/scope"
	"github.com/euforia/thrap/consts"
	"github.com/euforia/thrap/devpack"
	"github.com/euforia/thrap/dockerfile"
	"github.com/euforia/thrap/thrapb"
	"github.com/euforia/thrap/utils"
	"github.com/euforia/thrap/vars"
	"github.com/euforia/thrap/vcs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type StackAsm struct {
	// vars available to the assembler as a whole
	vars scope.Variables
	// available devpacks
	packs *devpack.DevPacks

	vcs     vcs.VCS
	gitrepo *git.Repository

	worktree *git.Worktree

	stack *thrapb.Stack
}

// NewStackAsm returns a new stack assembler
func NewStackAsm(stack *thrapb.Stack,
	vcsp vcs.VCS, gitrepo *git.Repository,
	globalVars scope.Variables, packs *devpack.DevPacks) (*StackAsm, error) {

	asm := &StackAsm{
		vcs:     vcsp,
		gitrepo: gitrepo,
		packs:   packs,
		vars:    globalVars,
		stack:   stack,
	}

	// Add stack scope vars
	asm.vars = vars.MergeScopeVars(asm.vars, stack.ScopeVars())

	wtree, err := gitrepo.Worktree()
	if err == nil {
		asm.worktree = wtree
	}

	return asm, err
}

func (asm *StackAsm) Assemble() (err error) {
	st := asm.stack

	for _, cmpt := range st.Components {
		if cmpt.IsBuildable() {

			if err = asm.assembleDevComp(cmpt); err != nil {
				return err
			}

		} else {
			//
			// TODO: handle other components
			//
		}

	}

	return
}

// Commit writes out the manifest locally and commits all changes with the VCS
func (asm *StackAsm) Commit() error {
	return asm.vcsCommit()
}

func (asm *StackAsm) assembleDevComp(cmpt *thrapb.Component) error {
	langpack, err := asm.packs.Load(cmpt.Language.Lang())
	if err != nil {
		return err
	}

	casm := newDevCompAssembler(cmpt, langpack)
	if err = casm.assemble(asm.vars); err != nil {
		return err
	}

	//
	// TODO: introspect dockerFile
	//

	return asm.materializeDevComp(casm)

}

// TODO: change logic
func (asm *StackAsm) materializeDevComp(casm *devCompAssembler) error {

	files := make(map[string][]byte, 2)
	files[casm.comp.Build.Dockerfile] = []byte(casm.dockerfile.String())

	// TODO: check if file exists. load and confirm values
	files[dockerfile.DockerIgnoresFile] = []byte(strings.Join(casm.dockerignores, "\n"))

	// Default files to write independent of language pack
	name, content := asm.vcsIgnoreFile(casm.pack.IgnoreExts...)
	files[name] = content

	if _, ok := casm.files[consts.DefaultReadmeFile]; !ok {
		files[consts.DefaultReadmeFile] = asm.readmeFile()
	}

	return asm.writeFiles(casm.files, files)
}

// write all files
func (asm *StackAsm) writeFiles(filemaps ...map[string][]byte) error {
	var err error

	for _, files := range filemaps {

		for k, v := range files {
			err = asm.writeFile(k, v, false)
			if err != nil {
				break
			}
		}

		if err != nil {
			break
		}

	}

	return err
}

func (asm *StackAsm) writeFile(basename string, contents []byte, force bool) error {
	var (
		fs     = asm.worktree.Filesystem
		fsroot = fs.Root()
		path   = filepath.Join(fsroot, basename)
		err    error
	)

	if !utils.FileExists(path) || force {

		bk := filepath.Base(basename)
		if bk != basename {
			fs.MkdirAll(filepath.Dir(basename), 0755)
		}

		// fmt.Println("~ Writing:", basename)
		err = ioutil.WriteFile(path, contents, 0644)
		if err == nil {
			_, err = asm.worktree.Add(basename)
		}

	}

	return err
}

// returns the ignores base filename and its contents
func (asm *StackAsm) vcsIgnoreFile(add ...string) (string, []byte) {
	list := append(vcs.DefaultGitIgnores(), add...)
	content := strings.Join(list, "\n")
	return asm.vcs.IgnoresFile(), []byte(content)
}

func (asm *StackAsm) readmeFile() []byte {
	return []byte("# " + asm.stack.Name + "\n" + asm.stack.Description + "\n\n")
}

func (asm *StackAsm) WriteManifest() error {

	key := `manifest "` + asm.stack.ID + `"`
	out := map[string]interface{}{
		key: asm.stack,
	}

	b, err := hclencoder.Encode(&out)
	if err == nil {
		b = append(append([]byte("\n"), b...), []byte("\n")...)
		err = asm.writeFile(consts.DefaultManifestFile, b, false)
	}

	return err
}

func (asm *StackAsm) vcsCommit() error {
	// we set the signature to thrap as it performed the init
	commitOpt := &git.CommitOptions{
		Author: &object.Signature{
			Name:  "thrap",
			Email: "thrap",
			When:  time.Now(),
		},
	}
	_, err := asm.worktree.Commit("Initial commit", commitOpt)
	return err
}
