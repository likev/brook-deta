// Copyright (c) 2016-present Cloud <cloud@txthinking.com>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of version 3 of the GNU General Public
// License as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
//	"context"
//	"crypto/x509"
	"errors"
	"fmt"
//	"io/fs"
//	"io/ioutil"
	"log"
//	"net"
	"os"
//	"os/exec"
	"os/signal"
	"path/filepath"
//	"runtime"
	"strings"
//	"sync"
	"syscall"
//	"time"

//	"net/http"
//	"net/url"

	"github.com/txthinking/brook"
	"github.com/txthinking/brook/plugins/block"
	"github.com/txthinking/brook/plugins/pprof"
	"github.com/txthinking/brook/plugins/socks5dial"
//	"github.com/txthinking/brook/plugins/thedns"
//	"github.com/txthinking/brook/plugins/tproxy"
	"github.com/txthinking/runnergroup"
	"github.com/txthinking/socks5"
	"github.com/urfave/cli/v2"
)

func main() {
	g := runnergroup.New()
	app := cli.NewApp()
	app.Name = "Brook"
	app.Version = "20230218"
	app.Usage = "A cross-platform network tool designed for developers"
	app.Authors = []*cli.Author{
		{
			Name:  "Cloud",
			Email: "cloud@txthinking.com",
		},
	}
	app.Copyright = "https://github.com/txthinking/brook"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "pprof",
			Usage: "go http pprof listen addr, such as :6060",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.String("pprof") != "" {
			p, err := pprof.NewPprof(c.String("pprof"))
			if err != nil {
				return err
			}
			g.Add(&runnergroup.Runner{
				Start: func() error {
					return p.ListenAndServe()
				},
				Stop: func() error {
					return p.Shutdown()
				},
			})
		}
		return nil
	}
	app.Commands = []*cli.Command{
		&cli.Command{
			Name:  "wsserver",
			Usage: "Run as brook wsserver, both TCP and UDP, it will start a standard http server and websocket server",
			BashComplete: func(c *cli.Context) {
				l := c.Command.VisibleFlags()
				for _, v := range l {
					fmt.Println("--" + v.Names()[0])
				}
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "listen",
					Aliases: []string{"l"},
					Usage:   "Listen address, like: ':80'",
					Value: ":" + os.Getenv("PORT"),
				},
				&cli.StringFlag{
					Name:    "password",
					Aliases: []string{"p"},
					Usage:   "Server password",
					Value: os.Getenv("SECRET_WS"),
				},
				&cli.StringFlag{
					Name:  "path",
					Usage: "URL path",
					Value: "/ws",
				},
				&cli.BoolFlag{
					Name:  "withoutBrookProtocol",
					Usage: "The data will not be encrypted with brook protocol",
				},
				&cli.IntFlag{
					Name:  "tcpTimeout",
					Value: 0,
					Usage: "time (s)",
				},
				&cli.IntFlag{
					Name:  "udpTimeout",
					Value: 60,
					Usage: "time (s)",
				},
				&cli.StringFlag{
					Name:  "blockDomainList",
					Usage: "One domain per line, suffix match mode. https://, http:// or local file absolute path. Like: https://txthinking.github.io/bypass/example_domain.txt",
				},
				&cli.StringFlag{
					Name:  "blockCIDR4List",
					Usage: "One CIDR per line, https://, http:// or local file absolute path, like: https://txthinking.github.io/bypass/example_cidr4.txt",
				},
				&cli.StringFlag{
					Name:  "blockCIDR6List",
					Usage: "One CIDR per line, https://, http:// or local file absolute path, like: https://txthinking.github.io/bypass/example_cidr6.txt",
				},
				&cli.StringSliceFlag{
					Name:  "blockGeoIP",
					Usage: "Block IP by Geo country code, such as US",
				},
				&cli.Int64Flag{
					Name:  "updateListInterval",
					Usage: "Update list interval, second. default 0, only read one time on start",
				},
				&cli.StringFlag{
					Name:  "toSocks5",
					Usage: "Forward to socks5 server, requires your socks5 supports standard socks5 TCP and UDP, such as 1.2.3.4:1080",
				},
				&cli.StringFlag{
					Name:  "toSocks5Username",
					Usage: "Forward to socks5 server, username",
				},
				&cli.StringFlag{
					Name:  "toSocks5Password",
					Usage: "Forward to socks5 server, password",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("listen") == "" || c.String("password") == "" {
					return cli.ShowSubcommandHelp(c)
				}
				if c.String("blockDomainList") != "" && !strings.HasPrefix(c.String("blockDomainList"), "http://") && !strings.HasPrefix(c.String("blockDomainList"), "https://") && !filepath.IsAbs(c.String("blockDomainList")) {
					return errors.New("--blockDomainList must be with absolute path")
				}
				if c.String("blockCIDR4List") != "" && !strings.HasPrefix(c.String("blockCIDR4List"), "http://") && !strings.HasPrefix(c.String("blockCIDR4List"), "https://") && !filepath.IsAbs(c.String("blockCIDR4List")) {
					return errors.New("--blockCIDR4List must be with absolute path")
				}
				if c.String("blockCIDR6List") != "" && !strings.HasPrefix(c.String("blockCIDR6List"), "http://") && !strings.HasPrefix(c.String("blockCIDR6List"), "https://") && !filepath.IsAbs(c.String("blockCIDR6List")) {
					return errors.New("--blockCIDR6List must be with absolute path")
				}
				if c.String("blockDomainList") != "" || c.String("blockCIDR4List") != "" || c.String("blockCIDR6List") != "" || len(c.StringSlice("blockGeoIP")) != 0 {
					p, err := block.NewBlock(c.String("blockDomainList"), c.String("blockCIDR4List"), c.String("blockCIDR6List"), c.StringSlice("blockGeoIP"), c.Int("updateListInterval"))
					if err != nil {
						return err
					}
					p.TouchBrook()
					if c.Int("updateListInterval") != 0 {
						g.Add(&runnergroup.Runner{
							Start: func() error {
								p.Update()
								return nil
							},
							Stop: func() error {
								p.Stop()
								return nil
							},
						})
					}
				}
				if c.String("toSocks5") != "" {
					p, err := socks5dial.NewSocks5Dial(c.String("toSocks5"), c.String("toSocks5Username"), c.String("toSocks5Password"), c.Int("tcpTimeout"), c.Int("udpTimeout"))
					if err != nil {
						return err
					}
					p.TouchBrook()
				}
				s, err := brook.NewWSServer(c.String("listen"), c.String("password"), "", c.String("path"), c.Int("tcpTimeout"), c.Int("udpTimeout"), c.Bool("withoutBrookProtocol"))
				if err != nil {
					return err
				}
				g.Add(&runnergroup.Runner{
					Start: func() error {
						return s.ListenAndServe()
					},
					Stop: func() error {
						return s.Shutdown()
					},
				})
				return nil
			},
		},
	}
	if os.Getenv("SOCKS5_DEBUG") != "" {
		socks5.Debug = true
	}

	//os.Args
	osArgs := []string{"./brook", "wsserver"}
	if err := app.Run(osArgs); err != nil {
		log.Println(err)
		return
	}
	if len(g.Runners) == 0 {
		return
	}
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		g.Done()
	}()
	log.Println(g.Wait())
}
