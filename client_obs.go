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

import (
	"fmt"
	"os"

	"github.com/kris-nova/logger"

	obsws "github.com/christopher-dG/go-obs-websocket"
)

const (
	ENVAR_OBSPort     = "TWINX_OBS_PORT"
	ENVAR_OBSPassword = "TWINX_OBS_PASSWORD"
	ENVAR_OBSHost     = "TWINX_OBS_HOST"
)

type OBSClient struct {
	Client *obsws.Client
}

func NewOBSClient() *OBSClient {
	return &OBSClient{
		//
	}
}

func (c *OBSClient) Authenticate() error {
	client := obsws.Client{
		Host:     os.Getenv(ENVAR_OBSHost),
		Port:     GetenvInt(ENVAR_OBSPort),
		Password: os.Getenv(ENVAR_OBSPassword),
	}
	err := client.Connect()
	if err != nil {
		return fmt.Errorf("unable to connect to OBS websocket. Change settings OBS > tools > WebSockets Server Settings > Password/Port/Host. Also make sure https://github.com/Palakis/obs-websocket is installed. AUR: https://aur.archlinux.org/packages/obs-websocket")
	}
	c.Client = &client
	logger.Success("Successfully authenticated with OBS!")
	return nil
}
