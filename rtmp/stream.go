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
	"sync"

	"github.com/kris-nova/logger"
)

type Stream struct {
	uid     string
	key     string
	isStart bool
	cache   *Cache
	r       ReadCloser
	ws      *sync.Map
}

func NewStream() *Stream {
	return &Stream{
		cache: NewCache(),
		ws:    &sync.Map{},
	}
}

func (s *Stream) ID() string {
	return s.uid
}

func (s *Stream) GetReader() ReadCloser {
	return s.r
}

func (s *Stream) GetWs() *sync.Map {
	return s.ws
}

func (s *Stream) Copy(dst *Stream) {
	dst.uid = s.uid
	dst.key = s.key
	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		s.ws.Delete(key)
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
		return true
	})
}

func (s *Stream) AddReader(r ReadCloser) {
	s.r = r
}

func (s *Stream) AddWriter(w WriteCloser) {
	pw := &PackWriterCloser{w: w}
	s.ws.Store(s.uid, pw)
}

func (s *Stream) TransactionStart() {
	logger.Info("Starting stream transaction")
	s.isStart = true
	var p Packet
	for {
		if !s.isStart {
			s.closeInter()
			return
		}
		err := s.r.Read(&p)
		if err != nil {
			s.closeInter()
			s.isStart = false
			return
		}

		s.cache.Write(p)

		s.ws.Range(func(key, val interface{}) bool {
			v := val.(*PackWriterCloser)
			if !v.init {
				if err = s.cache.Send(v.w); err != nil {
					s.ws.Delete(key)
					return true
				}
				v.init = true
			} else {
				newPacket := p
				if err = v.w.Write(&newPacket); err != nil {
					s.ws.Delete(key)
				}
			}
			return true
		})
	}
}

func (s *Stream) TransStop() {
	if s.isStart && s.r != nil {
		s.r.Close()
	}
	s.isStart = false
}

func (s *Stream) CheckAlive() (n int) {
	if s.r != nil && s.isStart {
		if s.r.Alive() {
			n++
		} else {
			// Read timeout
			s.r.Close()
		}
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			//Alive from RWBaser, check last frame now - timestamp, if > timeout then Remove it
			if !v.w.Alive() {
				// Write Timeout
				s.ws.Delete(key)
				v.w.Close()
				return true
			}
			n++
		}
		return true
	})

	return
}

func (s *Stream) closeInter() {
	if s.r != nil {
		//s.StopStaticPush()
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			v.w.Close()
			//if v.w.Info().IsInterval() {
			//	s.ws.Delete(key)
			//	logger.Info("[%v] player closed and remove\n", v.w.Info())
			//}
		}
		return true
	})
}
