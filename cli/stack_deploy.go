/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package cli

import (
	"fmt"

	"github.com/sniperkit/snk.fork.thrap/manifest"
	"github.com/sniperkit/snk.fork.thrap/orchestrator"
	"github.com/sniperkit/snk.fork.thrap/store"
	"github.com/sniperkit/snk.fork.thrap/utils"
	"github.com/sniperkit/snk.fork.thrap/vcs"
	"gopkg.in/urfave/cli.v2"
)

func commandStackDeploy() *cli.Command {
	return &cli.Command{
		Name:  "deploy",
		Usage: "Deploy stack",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dryrun",
				Aliases: []string{"dry"},
				Usage:   "perform a dry run",
				Value:   false,
			},
		},
		Action: func(ctx *cli.Context) error {
			stack, err := manifest.LoadManifest("")
			if err != nil {
				return err
			}
			// Load stack version
			lpath, err := utils.GetLocalPath("")
			if err != nil {
				return err
			}

			// Load profiles
			profs, err := store.LoadHCLFileProfileStorage(lpath)
			if err != nil {
				return err
			}

			// Load request profile
			profName := ctx.String("profile")
			prof := profs.Get(profName)
			if prof == nil {
				return fmt.Errorf("profile not found: %s", profName)
			}

			stack.Version = vcs.GetRepoVersion(lpath).String()
			fmt.Println(stack.ID, stack.Version)

			cr, err := loadCore(ctx)
			if err != nil {
				return err
			}

			opt := orchestrator.RequestOptions{Dryrun: ctx.Bool("dryrun")}
			st, err := cr.Stack(prof)
			if err != nil {
				return err
			}

			return st.Deploy(stack, opt)
		},
	}
}
