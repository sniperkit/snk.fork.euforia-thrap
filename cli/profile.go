/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/euforia/hclencoder"
	"github.com/sniperkit/snk.fork.thrap/store"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
	"github.com/sniperkit/snk.fork.thrap/utils"
	"gopkg.in/urfave/cli.v2"
)

func commandProfile() *cli.Command {
	return &cli.Command{
		Name:      "profile",
		Usage:     "Iteract with profiles",
		ArgsUsage: "<profile>",
		Action: func(ctx *cli.Context) error {

			ppath, err := utils.GetLocalPath("")
			if err != nil {
				return err
			}

			profs, err := store.LoadHCLFileProfileStorage(ppath)
			if err != nil {
				return err
			}

			var (
				profIn  = ctx.Args().Get(0)
				display interface{}
			)

			if profIn == "" {
				display = profs.Profiles
				os.Stdout.Write([]byte("\n"))
			} else {
				kv := strings.Split(profIn, "=")
				if len(kv) < 1 {
					cli.ShowCommandHelpAndExit(ctx, "profile", 1)
				}
				id := kv[0]

				prof := profs.Get(id)
				if prof == nil {
					return store.ErrProfileNotFound
				}
				display = map[string]*thrapb.Profile{id: prof}
				fmt.Println()
			}

			b, _ := hclencoder.Encode(&display)
			fmt.Printf("%s\n", b)

			return nil
		},
	}
}
