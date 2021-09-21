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

package twinx

type Launcher struct {
	dryRun      bool
	Title       string
	Description string

	TwitchClient  *InteractiveTwitchClient
	OBSClient     *OBSClient
	YouTubeClient *YouTubeClient
}

func NewLauncher(title, description string) *Launcher {
	return &Launcher{
		Title:       title,
		Description: description,
	}
}

func (l *Launcher) SetDryRun(dryRun bool) {
	l.dryRun = dryRun
}

func (l *Launcher) Start() error {
	var err error

	// Initialize OBS
	l.OBSClient = NewOBSClient()
	err = l.OBSClient.Authenticate()
	if err != nil {
		return err
	}

	// Initialize Twitch
	l.TwitchClient = NewInteractiveTwitchClient()
	err = l.TwitchClient.Authenticate()
	if err != nil {
		return err
	}

	// Initialize YouTube
	l.YouTubeClient = NewYouTubeClient()
	err = l.YouTubeClient.Authenticate()
	if err != nil {
		return err
	}

	return nil
}
