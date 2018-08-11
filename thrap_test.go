/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package thrap

import (
	"fmt"
	"testing"

	"github.com/sniperkit/snk.fork.thrap/utils"
	"github.com/sniperkit/snk.fork.thrap/vcs"
	"gopkg.in/src-d/go-git.v4"
)

func Test_foo(t *testing.T) {
	vcsp := vcs.NewGitVCS()
	vcsp.Init(nil)
	cwd, _ := utils.GetLocalPath("")
	r, _ := vcsp.Open(&vcs.Repository{Name: "thrap"}, vcs.Option{Path: cwd})
	repo := r.(*git.Repository)
	wt, _ := repo.Worktree()
	fmt.Println(wt.Status())
}
