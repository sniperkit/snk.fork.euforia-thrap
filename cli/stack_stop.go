/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/sniperkit/snk.fork.thrap/manifest"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
	"github.com/sniperkit/snk.fork.thrap/utils"
	"gopkg.in/urfave/cli.v2"
)

func commandStackStop() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop stack components",
		Action: func(ctx *cli.Context) error {

			stack, err := manifest.LoadManifest("")
			if err != nil {
				return err
			}
			if errs := stack.Validate(); len(errs) > 0 {
				return utils.FlattenErrors(errs)
			}

			cr, err := loadCore(ctx)
			if err != nil {
				return err
			}

			stm, err := cr.Stack(thrapb.DefaultProfile())
			if err != nil {
				return err
			}

			var stop bool
			utils.PromptUntilNoError("Are you sure you want to stop "+stack.ID+" [y/N] ? ",
				os.Stdout, os.Stdin, func(in []byte) error {
					s := string(in)
					switch s {
					case "y", "Y", "yes", "Yes":
						stop = true
					}
					return nil
				})

			if stop {
				report := stm.Stop(context.Background(), stack)
				defaultPrintStackResults(report)
			} else {
				fmt.Println("Exiting!")
			}

			return nil
		},
	}
}

func commandStackDestroy() *cli.Command {
	return &cli.Command{
		Name:  "destroy",
		Usage: "Destroy stack components",
		Action: func(ctx *cli.Context) error {

			stack, err := manifest.LoadManifest("")
			if err != nil {
				return err
			}
			if errs := stack.Validate(); len(errs) > 0 {
				return utils.FlattenErrors(errs)
			}

			_, prof, err := loadProfile(ctx)
			if err != nil {
				return err
			}

			cr, err := loadCore(ctx)
			if err != nil {
				return err
			}

			stm, err := cr.Stack(prof)
			if err != nil {
				return err
			}

			var destroy bool
			utils.PromptUntilNoError("Are you sure you want to destroy "+stack.ID+" [y/N] ? ",
				os.Stdout, os.Stdin, func(in []byte) error {
					s := string(in)
					switch s {
					case "y", "Y", "yes", "Yes":
						destroy = true
					}
					return nil
				})

			if destroy {
				report := stm.Destroy(context.Background(), stack)
				defaultPrintStackResults(report)
			} else {
				fmt.Println("Exiting!")
			}

			return nil
		},
	}
}
