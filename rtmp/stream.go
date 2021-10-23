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

	"github.com/kris-nova/logger"
)

type Stream struct {
	URLAddr
	key       string
	chunkSize uint32
	conns     map[string]*Conn
	mtx       sync.Mutex
	metaData  *ChunkStream
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
	// Hacky cache
	mx[key] = s
	return s
}

func (s *Stream) SetChunkSize(chunkSize uint32) {
	logger.Debug(rtmpMessage(fmt.Sprintf("SetChunkSize: %d", chunkSize), ack))
	s.chunkSize = chunkSize
}

func (s *Stream) AddMetaData(x *ChunkStream) {
	s.metaData = x
}

func (s *Stream) GetMetaData() *ChunkStream {
	if s.metaData == nil {
		panic("nil metadata for play client")
	}
	return s.metaData
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

	// Implement connectTX()

	// write the chunk size of the client to the server
	if s.chunkSize == 0 {
		return fmt.Errorf("invalid chunk size: %d", s.chunkSize)
	}
	logger.Debug(rtmpMessage(fmt.Sprintf("SetChunkSize: %d", s.chunkSize), tx))
	err := s.Write(c.newChunkStreamSetChunkSize(s.chunkSize))
	if err != nil {
		return err
	}

	err = s.Write(c.streamBegin())
	if err != nil {
		return err
	}

	return s.Write(s.metaData)
}

func (s *Stream) Write(x *ChunkStream) error {
	if x == nil {
		fmt.Errorf("nil packet in multiplex writter")
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
