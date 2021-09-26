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
	"net/url"
	"strings"
)

const (
	RTMPLocalAddress = "localhost"

	RTMPProtocol string = "tcp"

	// RTMPBufferSizeOBSDefaultBytes is the default output buffer size used by OBS.
	// This should be used for the simplest and smoothest use with OBS.
	// This can be adjusted (and so should OBS!) if you are sure what you
	// are doing, and have system resources to support your change.
	RTMPBufferSizeOBSDefaultBytes int64 = 2500

	// RTMPBufferSizeNovaDefaultBytes is my personal default buffer size for my
	// streams. I run Arch btw.
	RTMPBufferSizeNovaDefaultBytes int64 = 256

	RTMPPrefix = "rtmp://"
)

type RTMPAddr struct {
	Raw string
	URL *url.URL
}

func RTMPNewAddr(raw string) (*RTMPAddr, error) {
	url, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("unable to url.Parse raw rtmp string: %s", err)
	}
	return &RTMPAddr{
		Raw: raw,
		URL: url,
	}, nil
}

func (r *RTMPAddr) Full() string {
	if !strings.HasPrefix(r.Raw, RTMPPrefix) {
		return fmt.Sprintf("%s%s", RTMPPrefix, r.Raw)
	}
	return r.Raw
}

func (r *RTMPAddr) Server() string {
	return r.URL.Host
}
