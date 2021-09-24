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
	"context"
	"fmt"
	"os"

	"github.com/kris-nova/twinx/activestreamer"

	"github.com/kris-nova/twinx"

	"github.com/kris-nova/logger"
	"github.com/urfave/cli/v2"
)

func main() {
	twinx.PrintBanner()
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

	// rtmpPort is the port to listen for the local RTMP server
	rtmpPort string = "1719"

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
			cli.ShowSubcommandHelp(context)
			return nil
		},
		Flags: globalFlags,
		Commands: []*cli.Command{

			// ********************************************************
			// [ RTMP ]
			// ********************************************************

			{
				Name:      "rtmp",
				Aliases:   []string{"r"},
				Usage:     "RTMP stream proxy, and forwarding.",
				UsageText: ``,
				Flags:     allFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					cli.ShowSubcommandHelp(c)
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:      "listen",
						Usage:     "Start an RTMP server locally.",
						UsageText: ``,
						Flags: allFlags([]cli.Flag{
							&cli.StringFlag{
								Name:        "port",
								Aliases:     []string{"p"},
								Value:       "1719",
								Usage:       "toggle verbose mode for logger",
								Destination: &rtmpPort,
							},
						}),
						Action: func(c *cli.Context) error {
							x, err := twinx.GetActiveStream()
							if err != nil {
								return fmt.Errorf("unable to find active running stream: %v", err)
							}
							ack, err := x.Client.StartRTMP(context.TODO(), &activestreamer.RTMP{
								Port: rtmpPort,
							}, nil)
							if err != nil {
								return fmt.Errorf("unable to start RTMP: %v", err)
							}
							if ack.Success {
								logger.Always("Success!")
								return nil
							}
							return fmt.Errorf("error RTMP: %s", ack.Message)
						},
					},
					{
						Name:      "stop",
						Usage:     "Stop the RTMP server.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							return nil
						},
					},
					{
						Name:      "proxy",
						Usage:     "An advanced RTMP proxy that can be used in parallel. Example: Proxy to both Twitch and YouTube.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							return nil
						},
					},
				},
			},

			// ********************************************************
			// [ OBS ]
			// ********************************************************

			{
				Name:      "obs",
				Aliases:   []string{"o"},
				Usage:     "The OBS subresource. Used to control OBS.",
				UsageText: ``,
				Flags:     allFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					cli.ShowSubcommandHelp(c)
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:      "start",
						Usage:     "Start an OBS Stream ",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							return nil
						},
					},
				},
			},

			// ********************************************************
			// [ Twitch ]
			// ********************************************************

			{
				Name:      "twitch",
				Aliases:   []string{"t"},
				Usage:     "The Twitch subresource. Used to control Twitch.",
				UsageText: ``,
				Flags:     allFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					cli.ShowSubcommandHelp(c)
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:      "update",
						Usage:     "Send the current StreamMeta to Twitch to use.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							return nil
						},
					},
				},
			},

			// ********************************************************
			// [ YouTube ]
			// ********************************************************

			{
				Name:      "youtube",
				Aliases:   []string{"yt"},
				Usage:     "The YouTube subresource. Used to control YouTube.",
				UsageText: ``,
				Flags:     allFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					cli.ShowSubcommandHelp(c)
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:      "update",
						Usage:     "Send the current StreamMeta to YouTube to use.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							return nil
						},
					},
				},
			},

			// ********************************************************
			// [ Stream ]
			// ********************************************************

			{
				Name:      "stream",
				Aliases:   []string{"s"},
				Usage:     "The stream subresource. Used to manage streams at runtime.",
				UsageText: ``,
				Flags:     allFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					cli.ShowSubcommandHelp(c)
					return nil
				},
				Subcommands: []*cli.Command{
					// Stream Start
					{
						Name:      "start",
						Usage:     "Start a new stream. Only one stream can be ran at a time.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {

							args := c.Args()
							if args.Len() != 2 {
								return fmt.Errorf("usage: twinx stream start <title> <description>")
							}
							title := args.Get(0)
							description := args.Get(1)

							x, err := twinx.NewActiveStream(title, description)
							if err != nil {
								return fmt.Errorf("unable to start new active stream: %v", err)
							}
							logger.Info("Child ActiveStreamer PID %d", x.PID)
							err = x.Assure()
							if err != nil {
								return fmt.Errorf("unable to connect to gRPC server over unix domain socket: %v", err)
							}
							logger.Always("Success!")
							return nil
						},
					},

					// Stream assure
					{
						Name:      "assure",
						Usage:     "Connect to an existing active stream, and validate the connection.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							x, err := twinx.GetActiveStream()
							if err != nil {
								return fmt.Errorf("unable to find active running stream: %v", err)
							}
							err = x.Assure()
							if err != nil {
								return fmt.Errorf("unable to connect to gRPC server over unix domain socket: %v", err)
							}
							logger.Always("Connected to active stream!")
							return nil
						},
					},

					// Stream Stop
					{
						Name:      "stop",
						Usage:     "Stop any running stream.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							x, err := twinx.GetActiveStream()
							if err != nil {
								return fmt.Errorf("unable to find active running stream: %v", err)
							}
							err = twinx.StopActiveStream(x)
							if err != nil {
								return fmt.Errorf("unable to stop active stream. consider twinx stream kill: %v", err)
							}
							logger.Always("SIGHUP sent...")
							return nil
						},
					},

					// Stream Kill
					{
						Name:      "kill",
						Usage:     "Kill any existing stream. Forcefully.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							x, err := twinx.GetActiveStream()
							if err != nil {
								return fmt.Errorf("unable to find active running stream: %v", err)
							}
							err = twinx.KillActiveStream(x)
							if err != nil {
								return fmt.Errorf("unable to force kill active stream: %v", err)
							}
							logger.Always("SIGKILL sent...")
							return nil
						},
					},

					// Stream Clean
					{
						Name:      "clean",
						Usage:     "Clean any existing stream files. Forcefully.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							logger.Success("Remove %s", twinx.ActiveStreamPID)
							return os.Remove(twinx.ActiveStreamPID)
						},
					},

					// Stream Info
					{
						Name:      "info",
						Usage:     "Print stream metrics and data.",
						UsageText: ``,
						Flags:     allFlags([]cli.Flag{}),
						Action: func(c *cli.Context) error {
							x, err := twinx.GetActiveStream()
							if err != nil {
								return fmt.Errorf("unable to find active running stream: %v", err)
							}
							ch := x.InfoChannel()
							for {
								logger.Always(<-ch)
							}
							return nil
						},
					},
				},
			},
			// ********************************************************
			// [ ActiveStreamer ]
			// ********************************************************

			{
				Name:      "activestreamer",
				Aliases:   []string{"x"},
				Usage:     "Run a new active streamer process in the foreground. Expert usage only. ⚠",
				UsageText: ``,
				Flags:     allFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					// Default verbose for daemon
					logger.BitwiseLevel = logger.LogEverything
					stream := twinx.NewStream()

					// This should log and exit cleanly.
					return stream.Run()
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
