//
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
// ============================================================================
//
//  ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗
//  ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝
//     ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝
//     ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗
//     ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗
//     ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
//
// ============================================================================

package main

import (
	"os"

	"github.com/kris-nova/logger"
	"github.com/urfave/cli/v2"
)

func main() {
	err := RunWithOptions(instanceOptions)
	if err != nil {
		logger.Critical("%v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

type RuntimeOptions struct {
}

var instanceOptions = &RuntimeOptions{}

func RunWithOptions(opt *RuntimeOptions) error {

	var verbose bool

	// cli assumes "-v" for version.
	// override that here
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "Print the version",
	}

	// ********************************************************
	// [ Twinx Application ]
	// ********************************************************

	app := &cli.App{
		Name:      "twinx",
		HelpName:  "A twitch focused command line tool for producing, archiving and managing live stream content. Built for Linux.",
		Usage:     "Framework for developing video production automation tasks.",
		UsageText: ``,
		Action: func(context *cli.Context) error {
			// TODO
			cli.ShowSubcommandHelp(context)
			return nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Value:       false,
				Usage:       "toggle verbose mode for logger",
				Destination: &verbose,
			},
		},
		Commands: []*cli.Command{

			// ********************************************************
			// [ Stream ]
			// ********************************************************

			{
				Name:      "stream",
				Aliases:   []string{"i"},
				Usage:     "Start a new stream!",
				UsageText: "twinx stream 'Content of your stream'",
				Action: func(c *cli.Context) error {
					logger.Always("Starting stream...")
					return nil
				},
			},
		},
	}

	return app.Run(os.Args)
}
