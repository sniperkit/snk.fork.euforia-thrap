/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/sniperkit/snk.fork.thrap/consts"
	"github.com/sniperkit/snk.fork.thrap/packs"
	"gopkg.in/urfave/cli.v2"
)

func commandPack() *cli.Command {
	return &cli.Command{
		Name:  "pack",
		Usage: "Pack operations",
		Subcommands: []*cli.Command{
			commandPackUpdate(),
			commandPackList(),
		},
	}
}

func commandPackUpdate() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update packs",
		Action: func(ctx *cli.Context) error {
			packdir := filepath.Join(consts.DefaultDataDir, consts.PacksDir)
			pks, err := packs.New(packdir)
			if err != nil {
				return err
			}

			return pks.Update()
		},
	}
}

func commandPackList() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Usage:   "List available packs",
		Aliases: []string{"ls"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Usage:   "filter by `type`",
			},
		},
		Action: func(ctx *cli.Context) error {
			packdir := filepath.Join(consts.DefaultDataDir, consts.PacksDir)
			pks, err := packs.New(packdir)
			if err != nil {
				return err
			}

			typ := ctx.String("type")
			switch typ {
			case "web":
				p := pks.Web()
				list, _ := p.List()
				for _, l := range list {
					fmt.Println(l)
				}

			case "dev":
				p := pks.Dev()
				list, _ := p.List()
				for _, l := range list {
					fmt.Println(l)
				}

			case "datastore":
				p := pks.Datastore()
				list, _ := p.List()
				for _, l := range list {
					fmt.Println(l)
				}

			case "":
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.StripEscape)
				fmt.Fprintf(w, "TYPE\tID\n")

				dp := pks.Dev()
				list, _ := dp.List()
				for _, l := range list {
					fmt.Fprintf(w, "dev\t%s\n", l)
				}

				dsp := pks.Datastore()
				list, _ = dsp.List()
				for _, l := range list {
					fmt.Fprintf(w, "datastore\t%s\n", l)
				}

				wp := pks.Web()
				list, _ = wp.List()
				for _, l := range list {
					fmt.Fprintf(w, "web\t%s\n", l)
				}
				w.Flush()

			default:
				return errors.New("unknown pack type: " + typ)
			}

			return nil
		},
	}
}
