/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package core

import (
	"crypto/ecdsa"
	"log"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sniperkit/snk.fork.thrap/config"
	"github.com/sniperkit/snk.fork.thrap/consts"
	"github.com/sniperkit/snk.fork.thrap/crt"
	"github.com/sniperkit/snk.fork.thrap/orchestrator"
	"github.com/sniperkit/snk.fork.thrap/packs"
	"github.com/sniperkit/snk.fork.thrap/registry"
	"github.com/sniperkit/snk.fork.thrap/secrets"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
	"github.com/sniperkit/snk.fork.thrap/vcs"
)

var (
	errProviderNotConfigured = errors.New("provider not configured")
	errPacksDirMissing       = errors.New("packs directory missing")
	errDataDirMissing        = errors.New("data directory missing")
	errOrchNotLoaded         = errors.New("orchestrator not loaded")
	errRegNotLoaded          = errors.New("registry not loaded")
)

const (
	// Temporary default
	defaultPacksRepoURL = "https://github.com/sniperkit/snk.fork.thrap-packs.git"
)

// Core is the thrap core
type Core struct {
	conf  *config.ThrapConfig
	creds *config.CredsConfig

	// Remote VCS github etc.
	vcs vcs.VCS

	// Loaded registries
	regs map[string]registry.Registry

	// Secrets engine
	sec secrets.Secrets

	// Deployment orchestrator
	orchs map[string]orchestrator.Orchestrator

	// Loaded extension packs
	packs *packs.Packs

	// Container runtime. Currently docker
	crt *crt.Docker

	sst StackStorage
	ist IdentityStorage

	// Load keypair. Currently 1 per core
	kp *ecdsa.PrivateKey

	// Logger
	log *log.Logger
}

// NewCore loads the core engine with the global configs
func NewCore(conf *Config) (*Core, error) {
	c := &Core{}
	err := c.loadConfigs(conf)
	if err != nil {
		return nil, err
	}

	// Init CRT
	c.crt, err = crt.NewDocker()
	if err != nil {
		return nil, err
	}

	err = c.initKeyPair(conf.DataDir)
	if err != nil {
		return nil, err
	}

	err = c.initPacks(filepath.Join(conf.DataDir, consts.PacksDir))
	if err != nil {
		return nil, err
	}

	err = c.initProviders()
	if err == nil {
		err = c.initStores(conf.DataDir)
	}

	return c, err
}

// Config returns the currently loaded config.  This is the merged global and
// local config
func (core *Core) Config() *config.ThrapConfig {
	return core.conf
}

// Packs returns a pack instance containing the currently loaded packs
func (core *Core) Packs() *packs.Packs {
	return core.packs
}

// Stack returns a Stack instance that can be used to perform operations
// against a stack.  It is loaded pased on the profile provided.  All
// stack fields or constructed based on the profile
func (core *Core) Stack(profile *thrapb.Profile) (*Stack, error) {
	orch, ok := core.orchs[profile.Orchestrator]
	if !ok {
		return nil, errors.Wrap(errOrchNotLoaded, profile.Orchestrator)
	}

	stack := &Stack{
		crt:   core.crt,
		orch:  orch,
		conf:  core.conf.Clone(),
		vcs:   core.vcs,
		packs: core.packs,
		sst:   core.sst,
		log:   core.log,
	}

	// The registry may be empty for local builds
	if profile.Registry != "" {
		reg, ok := core.regs[profile.Registry]
		if !ok {
			return nil, errors.Wrap(errRegNotLoaded, profile.Registry)
		}
		stack.reg = reg
	}

	return stack, nil
}

// Identity returns an Identity instance to perform operations against
// identities
func (core *Core) Identity() *Identity {
	return &Identity{
		store: core.ist,
		log:   core.log,
	}
}

// KeyPair returns the public-private key currently held by the core
func (core *Core) KeyPair() *ecdsa.PrivateKey {
	return core.kp
}
