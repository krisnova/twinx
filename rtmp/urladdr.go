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
	"net"
	"net/url"
	"strconv"
	"strings"
)

// URLAddr is a flexible RTMP address member that resembles url.URL.
type URLAddr struct {
	url.URL
	net.Addr

	// raw can be any string, which we hope we can turn
	// into a valid *Addr
	raw string

	// scheme should always be DefaultScheme "rtmp://"
	scheme string

	// host is the host:port combination for the server
	// host should be valid with net.Listen() and net.Dial()
	host string

	// app is the first parameter to the RTMP URL
	// such as rtmp://host:port/app/key
	app string

	// key is the 2nd and final parameter to the RTMP URL
	// such as rtmp://host:port/app/key
	key string
}

func NewURLAddr(raw string) (*URLAddr, error) {
	var scheme, host, app, key string

	url, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("unable to url.Parse raw rtmp string: %s", err)
	}
	if len(url.Scheme) == 4 {
		scheme = url.Scheme
	}

	path := strings.Replace(raw, fmt.Sprintf("%s://", scheme), "", 1)

	if strings.Contains(path, "/") {
		splt := strings.Split(path, "/")
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
	} else if strings.Contains(path, ":") {
		splt := strings.Split(path, ":")
		if len(splt) == 2 {
			if len(splt[0]) == 0 {
				splt[0] = DefaultLocalHost
			}
			if len(splt[1]) == 0 {
				splt[0] = DefaultLocalPort
			}
			host = fmt.Sprintf("%s:%s", splt[0], splt[1])
		}
	}
	if scheme == "" {
		scheme = DefaultScheme
	}
	if host == "" {
		// Check for host/port
		host = fmt.Sprintf("%s:%s", DefaultLocalHost, DefaultLocalPort)
	}
	if app == "" {
		app = DefaultRTMPApp
	}
	if key == "" {
		key = generateKey()
	}

	a := &URLAddr{
		raw:    raw,
		scheme: scheme,
		host:   host,
		app:    app,
		key:    key,
	}

	// Grab the port
	rawHost, port, err := net.SplitHostPort(a.host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			port = DefaultLocalPort
		} else {
			return nil, fmt.Errorf("split host port: %v", err)
		}
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("convert port: %v", err)
	}
	if rawHost == DefaultLocalHost {
		ip := net.ParseIP(DefaultLo)
		a.Addr = &net.TCPAddr{
			IP:   ip,
			Port: portInt,
			Zone: "", // Only support IPv4 for now
		}
		a.URL = *url
		return a, nil
	}
	ips, err := net.LookupIP(a.host)
	if err != nil {
		return nil, fmt.Errorf("dns lookup: %s", a.host)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("dns lookup failure: no records: %s", a.SafeURL())
	}
	var ip net.IP
	if len(ips) == 1 {
		// This is the ideal state
		ip = ips[0]
	}
	if len(ips) > 1 {
		// For now just use the first one found
		ip = ips[0]
	}
	a.Addr = &net.TCPAddr{
		IP:   ip,
		Port: portInt,
		Zone: "", // Only support IPv4 for now
	}
	a.URL = *url
	return a, nil
}

// Host will return a net.Listener compatible host string as verbosely as possible.
// Given inputs such as:
//   localhost:
//   localhost:1935
//   :1935
//   :
// We should see
//   localhost:1935
func (a *URLAddr) Host() string {
	return a.host
}

// SafeURL will log the StreamURL() without the key.
//  rtmp://localhost:1935/app/[obfuscated]
func (a *URLAddr) SafeURL() string {
	return fmt.Sprintf("%s://%s/%s", a.scheme, a.host, a.app)
}

// StreamURL is a resolvable stream URL that can be played, published, or proxied.
//  rtmp://localhost:1935/app/key
func (a *URLAddr) StreamURL() string {
	return fmt.Sprintf("%s://%s/%s/%s", a.scheme, a.host, a.app, a.key)
}

// generateKey will generate a random stream key
func generateKey() string {
	b := make([]byte, DefaultGenerateKeyLength)
	for i := range b {
		b[i] = StreamKeyRandomBytePool[rand.Intn(len(StreamKeyRandomBytePool))]
	}
	return fmt.Sprintf("%s%s", DefaultGenerateKeyPrefix, string(b))
}

// Scheme should always return DefaultScheme "rtmp://"
func (a *URLAddr) Scheme() string {
	return a.scheme
}

// Key should return the stream key for this instance of *rtmp.Addr
// All instances will generate a key if one is not provided.
func (a *URLAddr) Key() string {
	return a.key
}

// App will return the first parameter of the path.
// Such as rtmp://host:port/app/key
func (a *URLAddr) App() string {
	return a.app
}

func (a *URLAddr) Network() string {
	return a.Addr.Network()
}

func (a *URLAddr) String() string {
	return a.Addr.String()
}

func (a *URLAddr) NewNetConn() (net.Conn, error) {
	return net.Dial(DefaultProtocol, a.StreamURL())
}

func (a *URLAddr) NewConn() (*Conn, error) {
	netConn, err := a.NewNetConn()
	if err != nil {
		return nil, err
	}
	conn := NewConn(netConn)
	return conn, nil
}
