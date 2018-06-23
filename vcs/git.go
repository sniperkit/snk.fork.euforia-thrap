package vcs

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/euforia/thrap/utils"
	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

const (
	gitUserConfigFile = ".gitconfig"
	gitIgnoresFile    = ".gitignore"
	defaultRemoteName = "origin"
)

// DefaultGitRemoteURL returns the default url scheme scheme for remote access
func DefaultGitRemoteURL(addr, owner, name string) string {
	return "ssh://git@" + addr + "/" + owner + "/" + name
}

// GitVCS implements a git version controlled interface
type GitVCS struct {
	globalUser  string
	globalEmail string
}

func NewGitVCS() *GitVCS {
	return &GitVCS{}
}

// ID returns the vcs provider id
func (g *GitVCS) ID() string {
	return "git"
}

// Init initializes the git provider with the config info in the users home dir
func (g *GitVCS) Init(conf map[string]interface{}) error {
	fpath, err := homedir.Expand("~/" + gitUserConfigFile)
	if err != nil {
		return err
	}

	if utils.FileExists(fpath) {
		err = g.loadFromGitConfig(fpath)
	}

	return err
}

// loadFromGitConfig loads settings from the git config file in the users home
// directory.  It does not return an error is the name and email are missing
func (g *GitVCS) loadFromGitConfig(fpath string) error {
	f, err := ini.Load(fpath)
	if err != nil {
		return err
	}

	sec := f.Section("user")

	nkey, err := sec.GetKey("name")
	if err == nil {
		g.globalUser = nkey.Value()
	}

	nkey, err = sec.GetKey("email")
	if err == nil {
		g.globalEmail = nkey.Value()
	}

	return nil
}

// GlobalUser satisfies the VCS interface
func (g *GitVCS) GlobalUser() string {
	return g.globalUser
}

// GlobalEmail satisfies the VCS interface
func (g *GitVCS) GlobalEmail() string {
	return g.globalEmail
}

// IgnoresFile satisfies the VCS interface
func (g *GitVCS) IgnoresFile() string {
	return gitIgnoresFile
}

// Create creates a new Repository returning a repo of the vcs' type. Each call
// only fills in missing pieces so multiple calls will not corrupt
func (g *GitVCS) Create(repo *Repository, opt Option) (interface{}, error) {
	var (
		gitRepo *git.Repository
		err     error
	)

	if len(opt.Path) == 0 {
		return nil, errors.New("path not specified")
	}

	gitRepo, err = git.PlainInit(opt.Path, false)

	// We still check remote
	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			return nil, err
		}

		// Open so we can set remote
		gitRepo, err = git.PlainOpen(opt.Path)
		if err != nil {
			return nil, err
		}

	}

	err = g.setupRemote(gitRepo, opt.Remote)

	return gitRepo, err
}

// Delete removes the .git directory leaving all other files intact
func (g *GitVCS) Delete(repo *Repository, opt Option) error {
	abspath := filepath.Join(opt.Path, ".git")
	return os.RemoveAll(abspath)
}

// setup remote configuration on local instance
func (g *GitVCS) setupRemote(repo *git.Repository, remote string) (err error) {
	if len(remote) == 0 {
		return
	}

	_, err = parseURL(remote)
	if err != nil {
		return
	}

	rmt, err := repo.Remote(defaultRemoteName)
	if err == nil {
		conf := rmt.Config()
		if len(conf.URLs) > 0 {
			// Its has been externally set to we don't touch
			return
		}
		conf.URLs = []string{remote}
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: defaultRemoteName,
		URLs: []string{remote},
	})

	return err
}

func parseURL(uri string) (*url.URL, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https", "ssh":
	default:
		err = fmt.Errorf("unsupported URL scheme: '%s'", u.Scheme)
	}

	return u, err
}
