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

package rtmp

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/kris-nova/logger"
)

// Addr is a flexible RTMP addr reference
// rtmp://host/app/key
type Addr struct {

	// raw can be any string, which we hope we can turn
	// into a valid *Addr
	raw string

	scheme string

	host string

	app string

	key string
}

func NewAddr(raw string) (*Addr, error) {
	logger.Always(raw)
	var scheme, host, app, key string

	url, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("unable to url.Parse raw rtmp string: %s", err)
	}
	if len(url.Scheme) == 4 {
		scheme = url.Scheme
	}
	path := strings.Replace(raw, fmt.Sprintf("%s://", scheme), "", 1)

	// host/app/key
	if strings.Contains(path, "/") {
		splt := strings.Split(raw, "/")
		if len(splt) == 3 {
			host = splt[0]
			app = splt[1]
			key = splt[2]
		}
		if len(splt) == 2 {
			// Assume host and app
			host = splt[0]
			app = splt[1]
		}
		if len(splt) == 1 {
			// Assume host
			host = splt[0]
		}
		if len(splt) > 3 {
			return nil, fmt.Errorf("too many slashes: %s", raw)
		}
	} else if path == "" {
		app = DefaultRTMPApp
	}
	if scheme == "" {
		scheme = DefaultScheme
	}
	if host == "" {
		host = fmt.Sprintf("%s:%s", DefaultLocalHost, DefaultLocalPort)
	}
	if app == "" {
		app = DefaultRTMPApp
	}
	if key == "" {
		key = generateKey()
	}

	return &Addr{
		raw:    raw,
		scheme: scheme,
		host:   host,
		app:    app,
		key:    key,
	}, nil
}

// Host will return a net.Listener compatible host string
//   localhost:
//   localhost:1730
//   :1730
//   :
func (a *Addr) Host() string {
	return a.host
}

// StreamURL is a resolvable stream URL that can be played, published, or relayed.
//  rtmp://localhost:1730/app/key
//
func (a *Addr) StreamURL() string {
	return fmt.Sprintf("%s://%s/%s/%s", a.scheme, a.host, a.app, a.key)
}

// generateKey will generate a random stream key
func generateKey() string {
	b := make([]byte, DefaultGenerateKeyLength)
	for i := range b {
		b[i] = letterBytePool[rand.Intn(len(letterBytePool))]
	}
	return fmt.Sprintf("%s%s", DefaultGenerateKeyPrefix, string(b))
}

func (a *Addr) Scheme() string {
	return a.scheme
}

func (a *Addr) Key() string {
	return a.key
}

func (a *Addr) App() string {
	return a.app
}
