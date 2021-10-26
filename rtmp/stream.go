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
	dropped   int
}

var mx = map[string]*Stream{}

// Multiplex onto key
//
// All bytes written to this key (the base key)
// will be multiplexed onto any configured proxy clients.
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

func (s *Stream) AddMetaData(x *ChunkStream) error {
	s.metaData = x
	err := s.Write(s.metaData)
	if err != nil {
		return err
	}

	logger.Debug(rtmpMessage("Multiplex: StreamBegin", tx))
	for _, conn := range s.conns {
		err := conn.Write(conn.streamBegin())
		if err != nil {
			return err
		}
	}
	return nil
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
	if c.Key() == "" {
		return fmt.Errorf("empty conn key, unable to multiplex")
	}
	if c.SafeKey() == "" {
		return fmt.Errorf("unable to find safe key to hash metrics")
	}
	p := P(c.SafeURL())
	p.ProxyKeyHash = c.SafeKey()
	s.conns[c.SafeURL()] = c

	// All new conns need metadata right away

	// write the chunk size of the client to the server
	if s.chunkSize == 0 {
		return fmt.Errorf("invalid chunk size: %d", s.chunkSize)
	}
	logger.Debug(rtmpMessage(fmt.Sprintf("SetChunkSize: %d", s.chunkSize), tx))
	err := s.Write(c.newChunkStreamSetChunkSize(s.chunkSize))
	if err != nil {
		return err
	}
	return nil
}

// [ Write ]
//
// The almighty Write() method.
//
// This method is effectively a single threaded
// Write() method.
//
// If this blocks. All corresponding *Conn objects
// will also block.
func (s *Stream) Write(x *ChunkStream) error {
	if x == nil {
		fmt.Errorf("nil packet in multiplex writter")
	}

	if x == nil {
		return nil
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	packetWrite := false

	for _, c := range s.conns {
		if c == nil {
			continue
		}
		M().Lock()
		p := P(c.SafeURL())
		p.ProxyTotalBytesTX = p.ProxyTotalBytesTX + int(x.Length)
		p.ProxyTotalPacketsTX++
		M().Unlock()
		err := c.Write(x)
		if err != nil {
			s.conns[c.SafeURL()] = nil
			return err
		}

		// If we are dropping packets
		// it is going to be here.

		//err = c.Flush()
		//if err != nil {
		//	return err
		//}
		packetWrite = true
	}
	if !packetWrite {
		s.dropped++
	}

	return nil
}
