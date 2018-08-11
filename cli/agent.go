/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package cli

import (
	"log"
	"net"
	"os"

	"github.com/sniperkit/snk.fork.thrap"
	"github.com/sniperkit/snk.fork.thrap/config"
	"github.com/sniperkit/snk.fork.thrap/consts"
	"github.com/sniperkit/snk.fork.thrap/core"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
	"google.golang.org/grpc"
	"gopkg.in/urfave/cli.v2"
)

func commandAgent() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Run a server agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "bind-addr",
				Usage: "bind address",
				Value: "0.0.0.0:10000",
			},
			&cli.StringFlag{
				Name:  "data-dir",
				Usage: "Data directory",
				Value: consts.DefaultDataDir,
			},
			// &cli.StringFlag{
			// 	Name:  "adv-addr",
			// 	Usage: "advertise address",
			// },
		},
		Action: func(ctx *cli.Context) error {
			conf := &core.Config{
				DataDir: ctx.String("data-dir"),
				Logger:  log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds),
			}

			pconf, err := config.ReadProjectConfig(".")
			if err == nil {
				conf.ThrapConfig = pconf
			}

			core, err := core.NewCore(conf)
			if err != nil {
				return err
			}

			srv := grpc.NewServer()
			svc := thrap.NewService(core, conf.Logger)
			thrapb.RegisterThrapServer(srv, svc)

			baddr := ctx.String("bind-addr")
			lis, err := net.Listen("tcp", baddr)
			if err != nil {
				return err
			}
			conf.Logger.Println("Starting server:", lis.Addr().String())

			err = srv.Serve(lis)
			return err
		},
	}
}
