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
	"encoding/json"
	"fmt"

	"github.com/kris-nova/logger"

	"github.com/nicklaw5/helix"
)

// TwitchPermissions is what we ask for in oAuth
//
// These are very invasive, so please deny all!
// If you are unsure, do NOT add a permission!
//
//https://dev.twitch.tv/docs/authentication#scopes
var TwitchPermissions = []string{
	// Read email address
	"user:read:email",

	// Read follows
	//"user:read:follows",
}

const (
	// These constants are defined in the Twitch official app
	// dashboard and can be configured here:
	//	https://dev.twitch.tv/console/apps/ykiaywwbve0aa3vm15cruou06dpuct

	TWINX_PUBLIC_ID           string = "ykiaywwbve0aa3vm15cruou06dpuct"
	TWINX_PUBLIC_CALLBACK_URL string = "http://localhost:1717"
)

type InteractiveTwitchClient struct {
	AppID     string
	AppSecret string
	LoginURL  string
	Client    *helix.Client
}

type TwitchCallbackParameters struct {
	Code  []string // The response code from authenticating with Twitch
	Scope []string // scope:scope+scope:scope
	Error []string // Sometimes we can receive errors - so lets at least log them
}

func NewInteractiveTwitchClient() *InteractiveTwitchClient {
	return &InteractiveTwitchClient{
		//
	}
}

func (c *InteractiveTwitchClient) Authenticate() error {

	client, err := helix.NewClient(&helix.Options{
		ClientID:    TWINX_PUBLIC_ID,
		RedirectURI: TWINX_PUBLIC_CALLBACK_URL,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize twitch client: %v", err)
	}

	url := client.GetAuthorizationURL(&helix.AuthorizationURLParams{
		// Return an auth "Code" to use for later
		// Note: can also just use "token"
		ResponseType: "code",

		//
		Scopes: TwitchPermissions,
		//State:        "some-state",
		ForceVerify: false,
	})

	iCh, errCh := LocalhostServerGetParameters(&CallbackText{
		Title:       "Twinx - Nivenly.com",
		MainHeading: "Successfully Authenticated Twitch.tv",
		SubHeading1: "Open Source Live Streaming",
		SubHeading2: "You may now close this window.",
	})
	logger.Info("Waiting for callback response...")
	c.LoginURL = url
	err = OpenInBrowserDefault(c.LoginURL)
	if err != nil {
		return err
	}

	// After the user authenticates in their browser
	// the application will send parameters to localhost
	// Example: http://localhost:1717/?code=avsds324p93m96pv4bim4vyksrt4az&scope=user%3Aread%3Aemail+user%3Aread%3Afollows
	//    code string
	//    scope string+string

	// Hang until we get a response
	serverErr := <-errCh
	if serverErr != nil {
		return fmt.Errorf("Unable to auth application with Twitch from browser: %v", err)
	}

	jsonBytes := <-iCh
	// Start a web server and listen for twitch callback
	authResponse := TwitchCallbackParameters{}
	err = json.Unmarshal(jsonBytes, &authResponse)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response: %v", err)
	}

	if len(authResponse.Error) > 0 {
		return fmt.Errorf("error from Twitch while authenticating: %v", authResponse.Error)
	}

	// We get here we have an auth code!
	if len(authResponse.Code) != 1 {
		return fmt.Errorf("unsupported response from Twitch: returned len(code) != 1")
	}
	authenticatedCode := authResponse.Code[0]
	logger.Info("Success! Auth code from Twitch. [%v]", len(authenticatedCode))

	resp, err := client.RequestUserAccessToken(authenticatedCode)
	if err != nil {
		return fmt.Errorf("unable to request user access token: %v", err)
	}

	client.SetUserAccessToken(resp.Data.AccessToken)
	c.Client = client

	logger.Success("Successfully authenticated with Twitch!")
	return nil
}
