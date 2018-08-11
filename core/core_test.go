/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sniperkit/snk.fork.thrap/config"
	"github.com/sniperkit/snk.fork.thrap/consts"
	"github.com/sniperkit/snk.fork.thrap/manifest"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
	"github.com/sniperkit/snk.fork.thrap/utils"
	"github.com/stretchr/testify/assert"
)

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ConfigureGlobal(t *testing.T) {
	opt := DefaultConfigureOptions()
	opt.NoPrompt = true
	opt.DataDir, _ = ioutil.TempDir("/tmp", "cg-")
	err := ConfigureGlobal(opt)
	if err != nil {
		t.Fatal(err)
	}

	cf := filepath.Join(opt.DataDir, consts.CredsFile)
	assert.True(t, utils.FileExists(cf))
	cf = filepath.Join(opt.DataDir, consts.ConfigFile)
	assert.True(t, utils.FileExists(cf))
}

func Test_NewCore(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("/tmp", "core-")

	opt := DefaultConfigureOptions()
	opt.NoPrompt = true
	opt.DataDir = tmpdir
	err := ConfigureGlobal(opt)
	if err != nil {
		t.Fatal(err)
	}

	conf := &Config{DataDir: tmpdir}
	c, err := NewCore(conf)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, c.regs)
	assert.NotNil(t, c.sec)
	assert.NotNil(t, c.vcs)
	assert.NotNil(t, c.orchs)
	assert.NotNil(t, c.packs)
	assert.NotNil(t, c.orchs["docker"])

	_, err = c.Stack(&thrapb.Profile{Orchestrator: "foo"})
	assert.Contains(t, err.Error(), errOrchNotLoaded.Error())
}

func Test_Core_Build(t *testing.T) {
	if !utils.FileExists("/var/run/docker.sock") {
		t.Skip("Skipping: docker file descriptor not found")
	}

	tmpdir, _ := ioutil.TempDir("/tmp", "core-")

	opt := DefaultConfigureOptions()
	opt.NoPrompt = true
	opt.DataDir = tmpdir
	err := ConfigureGlobal(opt)
	if err != nil {
		t.Fatal(err)
	}

	conf := &Config{DataDir: tmpdir, ThrapConfig: &config.ThrapConfig{
		Registry: map[string]*config.RegistryConfig{
			"ecr": &config.RegistryConfig{
				ID:   "ecr",
				Addr: "foobar.com",
			},
		},
		Orchestrator: map[string]*config.OrchestratorConfig{
			"docker": &config.OrchestratorConfig{},
		},
	}}
	c, err := NewCore(conf)
	if err != nil {
		t.Fatal(err)
	}

	stack, err := manifest.LoadManifest("../test-fixtures/thrap.hcl")
	if err != nil {
		t.Fatal(err)
	}

	errs := stack.Validate()
	if len(errs) > 0 {
		fatal(t, utils.FlattenErrors(errs))
	}

	st, err := c.Stack(thrapb.DefaultProfile())
	if err != nil {
		t.Fatal(err)
	}
	err = st.Build(context.Background(), stack, BuildOptions{Workdir: "../"})
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Core_populateFromImageConf(t *testing.T) {

	if !utils.FileExists("/var/run/docker.sock") {
		t.Skip("Skipping: docker file descriptor not found")
	}

	tmpdir, _ := ioutil.TempDir("/tmp", "core-")
	defer os.RemoveAll(tmpdir)

	opt := DefaultConfigureOptions()
	opt.NoPrompt = true
	opt.DataDir = tmpdir
	err := ConfigureGlobal(opt)
	if err != nil {
		t.Fatal(err)
	}

	conf := &Config{DataDir: tmpdir, ThrapConfig: &config.ThrapConfig{
		Registry: map[string]*config.RegistryConfig{
			"ecr": &config.RegistryConfig{
				ID:   "ecr",
				Addr: "foobar.com",
			},
		},
		Orchestrator: map[string]*config.OrchestratorConfig{
			"docker": &config.OrchestratorConfig{},
		},
	}}
	c, err := NewCore(conf)
	if err != nil {
		t.Fatal(err)
	}

	stack, err := manifest.LoadManifest("../test-fixtures/thrap.hcl")
	if err != nil {
		t.Fatal(err)
	}

	stack.Validate()

	st, err := c.Stack(thrapb.DefaultProfile())
	if err != nil {
		t.Fatal(err)
	}

	st.populateFromImageConf(stack)
	assert.Equal(t, 1, len(stack.Components["vault"].Ports))
	assert.Equal(t, 5, len(stack.Components["consul"].Ports))
	assert.True(t, stack.Components["consul"].HasVolumeTarget("/consul/data"))
}

func Test_Core_Assembler(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("/tmp", "core.stack-")
	defer os.RemoveAll(tmpdir)

	opt := DefaultConfigureOptions()
	opt.NoPrompt = true
	opt.DataDir = tmpdir
	err := ConfigureGlobal(opt)
	fatal(t, err)

	lconf, err := config.ReadProjectConfig("../")
	fatal(t, err)
	conf := &Config{
		DataDir:     tmpdir,
		ThrapConfig: lconf,
		Logger:      DefaultLogger(os.Stdout),
	}

	c, err := NewCore(conf)
	if err != nil {
		t.Fatal(err)
	}

	stack, err := manifest.LoadManifest("../thrap.yml")
	if err != nil {
		t.Fatal(err)
	}
	stack.Validate()

	st, err := c.Stack(thrapb.DefaultProfile())
	if err != nil {
		t.Fatal(err)
	}
	sasm, err := st.Assembler("../", stack)
	if err != nil {
		t.Fatal(err)
	}

	err = sasm.Assemble()
	if err != nil {
		t.Fatal(err)
	}

	casm := sasm.ComponentAsm("registry")
	assert.NotNil(t, casm)
	fmt.Println(sasm.ComponentAsm("nomad").Dockerfile())
}
