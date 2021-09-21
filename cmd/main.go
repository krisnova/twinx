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
	"fmt"
	"os"

	"github.com/kris-nova/twinx"

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

// Global Flags
var (

	// verbose sets log verbosity
	verbose bool

	// dryRun will run the command without calling the services
	dryRun bool

	globalFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Aliases:     []string{"v"},
			Value:       false,
			Usage:       "toggle verbose mode for logger",
			Destination: &verbose,
		},
	}
)

func RunWithOptions(opt *RuntimeOptions) error {

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
		Name: "twinx",
		//HelpName:  "A twitch focused command line tool for producing, archiving and managing live stream content. Built for Linux.",
		Usage:     "Framework for developing video production automation tasks.",
		UsageText: ``,
		Version:   twinx.Version,
		Action: func(context *cli.Context) error {
			twinx.PrintBanner()
			cli.ShowSubcommandHelp(context)
			return nil
		},
		Flags: globalFlags,
		Commands: []*cli.Command{

			// ********************************************************
			// [ Stream ]
			// ********************************************************

			{
				Name:    "stream",
				Aliases: []string{"s"},
				Usage:   "Start a new stream!",
				UsageText: `
twinx stream <title> <description>
twinx stream "Working on Twinx" "A command line tool for live streaming"`,
				CustomHelpTemplate: fmt.Sprintf("%s%s", twinx.Banner(), DefaultSubCommandHelpTemplate),
				Flags: allFlags([]cli.Flag{
					&cli.BoolFlag{
						Name:        "dryrun",
						Aliases:     []string{"d"},
						Value:       false,
						Usage:       "toggle dryrun mode",
						Destination: &dryRun,
					},
				}),
				Action: func(c *cli.Context) error {
					allInit()
					args := c.Args()
					if args.Len() != 2 {
						cli.ShowCommandHelp(c, "stream")
						return nil
					}
					title := args.Get(0)
					description := args.Get(1)
					if dryRun {
						logger.Info("DRYRUN MODE ENABLED")
					}
					logger.Info("TITLE:       %s", title)
					logger.Info("DESCRIPTION: %s", description)
					logger.Always("Starting stream...")
					launcher := twinx.NewLauncher(title, description)
					launcher.SetDryRun(dryRun)
					return launcher.Start()
				},
			},
		},
	}

	return app.Run(os.Args)
}

func allInit() {
	if verbose {
		logger.BitwiseLevel = logger.LogEverything
		logger.Info("VERBOSE MODE ENABLED")
	} else {
		logger.BitwiseLevel = logger.LogAlways | logger.LogCritical | logger.LogDeprecated | logger.LogSuccess | logger.LogWarning
	}
}

func allFlags(flags []cli.Flag) []cli.Flag {
	return append(globalFlags, flags...)
}

// DefaultSubCommandHelpTemplate is taken from https://github.com/urfave/cli/blob/master/template.go
const DefaultSubCommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}
USAGE:
   {{if .UsageText}}{{.UsageText | nindent 3 | trim}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Description}}
DESCRIPTION:
   {{.Description | nindent 3 | trim}}{{end}}
COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`
