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

import "sync"

type Stream struct {
	URLAddr
	key   string
	conns map[string]*Conn
	mtx   sync.Mutex
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
	//go s.stream()
	return s
}

//func (s *Stream) stream() {
//
//}

func (s *Stream) AddConn(c *Conn) error {
	s.conns[c.SafeURL()] = c
	return nil
}

func (s *Stream) Write(x *ChunkStream) error {
	if x == nil {
		return nil
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, c := range s.conns {
		err := c.Write(x)
		if err != nil {
			return err
		}
		err = c.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}
