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
	"os"

	"github.com/kris-nova/twinx"

	"github.com/kris-nova/logger"

	"github.com/urfave/cli/v2"
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
		Flags:   []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Start a server that can accept client (play/publish) streams.",
				Action: func(c *cli.Context) error {
					return RunServer()
				},
			},
			{
				Name:    "client",
				Aliases: []string{"c"},
				Usage:   "Start a client that can send client (publish) streams.",
				Action: func(c *cli.Context) error {
					return RunClient()
				},
			},
			{
				Name:    "proxy",
				Aliases: []string{"p"},
				Usage:   "Start a proxy server that can accept client (publish) streams and proxy to remote (play) streams.",
				Action: func(c *cli.Context) error {
					return RunProxy()
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Critical(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func RunServer() error {
	return nil
}

func RunClient() error {
	return nil
}

func RunProxy() error {
	return nil
}
