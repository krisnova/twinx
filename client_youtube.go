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

	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	ENVAR_YouTubeAPIKey = "TWINX_YOUTUBE_API_KEY"
)

type YouTubeClient struct {
	Client *youtube.Service
}

func NewYouTubeClient() *YouTubeClient {
	return &YouTubeClient{
		//
	}
}

func (c *YouTubeClient) Authenticate() error {
	client, err := youtube.NewService(context.TODO(), option.WithAPIKey(os.Getenv(ENVAR_YouTubeAPIKey)))
	if err != nil {
		return fmt.Errorf("unable to authenticate with YouTube: %v", err)
	}
	c.Client = client
	logger.Success("Successfully authenticated with YouTube!")
	return nil
}
