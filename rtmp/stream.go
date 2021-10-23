// Copyright © 2021 Kris Nóva <kris@nivenly.com>
// Copyright (c) 2017 吴浩麟
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
	"sync"
)

type Stream struct {
	URLAddr
	key      string
	conns    map[string]*Conn
	mtx      sync.Mutex
	metaData *ChunkStream
}

var mx = map[string]*Stream{}

func Multiplex(key string) *Stream {
	s, ok := mx[key]
	if ok {
		return s
	}
	mx[key] = NewStream(key)
	return mx[key]
}

func NewStream(key string) *Stream {
	s := &Stream{
		key:   key,
		conns: make(map[string]*Conn),
	}
	return s
}

func (s *Stream) AddMetaData(x *ChunkStream) {
	s.metaData = x
}

func (s *Stream) RemoveConn(c *Conn) {
	// Setting to nil is safe, we check for nil and bypass
	// later in the Write()
	s.conns[c.SafeURL()] = nil
}

func (s *Stream) AddConn(c *Conn) error {
	if c.URLAddr.key == "" {
		return fmt.Errorf("empty conn key, unable to multiplex")
	}
	if c.URLAddr.SafeKey() == "" {
		return fmt.Errorf("unable to find safe key to hash metrics")
	}
	s.conns[c.SafeURL()] = c
	// All new conns need metadata right away
	return s.Write(s.metaData)
}

func (s *Stream) Write(x *ChunkStream) error {
	if x == nil {
		return nil
	}

	if x == nil {
		return nil
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, c := range s.conns {
		if c == nil {
			continue
		}
		M().Lock()
		P(c.SafeURL()).ProxyKeyHash = c.SafeKey()
		P(c.SafeURL()).ProxyTotalBytesTX = P(c.SafeURL()).ProxyTotalBytesTX + int(x.Length)
		P(c.SafeURL()).ProxyTotalPacketsTX++
		M().Unlock()
		err := c.Write(x)
		if err != nil {
			s.conns[c.SafeURL()] = nil
			return err
		}
		err = c.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}
