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

package rtmp

import (
	"fmt"
	"net"
)

type Listener struct {
	net.Listener
	addr *URLAddr
}

func newFromNetListener(l net.Listener) (*Listener, error) {
	netAddr := l.Addr()
	urlAddr, err := NewURLAddr(netAddr.String())
	if err != nil {
		return nil, fmt.Errorf("urlAddr: %v", err)
	}
	return &Listener{
		Listener: l,
		addr:     urlAddr,
	}, nil
}

func Listen(network string, address string) (*Listener, error) {
	addr, err := NewURLAddr(address)
	if err != nil {
		return nil, fmt.Errorf("rtmp URL addr: %v", err)
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, fmt.Errorf("rtmp listen: %v", err)
	}
	return &Listener{
		Listener: listener,
		addr:     addr,
	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	return l.Listener.Accept()
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.addr
}

func (l *Listener) URLAddr() *URLAddr {
	return l.addr
}
