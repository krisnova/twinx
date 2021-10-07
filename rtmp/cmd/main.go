// Copyright © 2021 Kris Nóva <kris@nivenly.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// ────────────────────────────────────────────────────────────────────────────
//
//  ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗
//  ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝
//     ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝
//     ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗
//     ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗
//     ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
//
// ────────────────────────────────────────────────────────────────────────────

package main

import (
	"errors"
	"os"

	"github.com/kris-nova/twinx/rtmp"

	"github.com/kris-nova/twinx"

	"github.com/kris-nova/logger"

	"github.com/urfave/cli/v2"
)

var (

	// clientPlay can be opted in to a client.Play() instead of default client.Publish()
	clientPlay bool = false

	// verbose enables log verbosity
	verbose bool = true

	globalFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Aliases:     []string{"v"},
			Usage:       "toggle verbose mode for logger",
			Destination: &verbose,
		},
	}
)

func main() {
	twinx.PrintBanner()

	// cli assumes "-v" for version.
	// override that here
	cli.VersionFlag = &cli.StringFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "Print the version",
	}

	app := &cli.App{
		Name:  "twinx-rtmp",
		Usage: "Simple, fast client, server, and proxy.",
		Action: func(c *cli.Context) error {
			cli.ShowSubcommandHelp(c)
			return nil
		},
		Version: twinx.CompileTimeVersion,
		Flags:   globalFlags,
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Start a server that can accept client (play/publish) streams.",
				Flags:   globalFlags,
				Action: func(c *cli.Context) error {
					args := c.Args()
					if args.Len() != 1 {
						return errors.New("usage: twinx-rtmp server <bind-addr>")
					}
					raw := args.First()
					return RunServer(raw)
				},
			},
			{
				Name:    "client",
				Aliases: []string{"c"},
				Usage:   "Start a client that can send client (publish) streams.",
				Flags: append([]cli.Flag{
					// Default publish (This is what OBS does)
					&cli.BoolFlag{
						Name:        "play",
						Destination: &clientPlay,
					},
				}, globalFlags...),
				Action: func(c *cli.Context) error {
					args := c.Args()
					if args.Len() != 1 {
						return errors.New("usage: twinx-rtmp server <bind-addr>")
					}
					raw := args.First()
					if clientPlay {
						return RunClientPlay(raw)
					}
					return RunClientPublish(raw)
				},
			},
			{
				Name:    "proxy",
				Aliases: []string{"p"},
				Usage:   "Start a proxy server that can accept client (publish) streams and proxy to remote (play) streams.",
				Flags:   globalFlags,
				Action: func(c *cli.Context) error {
					args := c.Args()
					if args.Len() != 2 {
						return errors.New("usage: twinx-rtmp proxy <play-addr> <publish-addr>")
					}
					play := args.Get(0)
					publish := args.Get(1)
					return RunProxy(play, publish)
				},
			},
		},
	}

	if verbose {
		logger.BitwiseLevel = logger.LogEverything
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Critical(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func RunServer(raw string) error {
	rtmpServer := rtmp.NewServer()
	rtmpListener, err := rtmp.Listen(raw)
	if err != nil {
		return err
	}
	return rtmpServer.Serve(rtmpListener)
}

func RunClientPlay(raw string) error {
	rtmpClient := rtmp.NewClient()
	err := rtmpClient.Dial(raw)
	if err != nil {
		return err
	}
	return rtmpClient.Play()
}

func RunClientPublish(raw string) error {
	rtmpClient := rtmp.NewClient()
	err := rtmpClient.Dial(raw)
	if err != nil {
		return err
	}
	return rtmpClient.Publish()
}

func RunProxy(play, publish string) error {
	playClient := rtmp.NewClient()
	err := playClient.Dial(play)
	if err != nil {
		return err
	}

	publishClient := rtmp.NewClient()
	err = publishClient.Dial(publish)
	if err != nil {
		return err
	}
	rtmpProxy := rtmp.NewRTMPProxy(playClient.Client(), publishClient.Client())
	return rtmpProxy.Start()
}
