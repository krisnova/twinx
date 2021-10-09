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
	"time"

	"github.com/kris-nova/logger"
)

// Service is the main RTMP active service.
//
// Multiple streams can be multiplexed over a single active service.
// We track each of the subordinate streams here.
type Service struct {

	// key is the stream key
	// this is auth material, please keep it safe
	key string

	// mux is the concurrent safe hashmap used to track
	// various subordinate streams.
	mux *sync.Map
}

// NewService will start a new RTMP service which streams can be added to later.
func NewService(key string) *Service {
	svc := &Service{
		key: key,
		mux: &sync.Map{},
	}
	go svc.CheckAlive()
	return svc
}

func (svc *Service) HandleReader(UID string, r ReadCloser) {
	logger.Warning(rtmpMessage("service.HandleReader", err))
	var stream *Stream
	i, ok := svc.mux.Load(svc.key)
	if stream, ok = i.(*Stream); ok {
		stream.TransStop()
		id := stream.ID()
		if id != "" && id != UID {
			ns := NewStream()
			stream.Copy(ns)
			stream = ns
			svc.mux.Store(svc.key, ns)
		}
	} else {
		stream = NewStream()
		svc.mux.Store(svc.key, stream)
	}

	stream.AddReader(r)
	go stream.TransactionStart()
}

func (svc *Service) HandleWriter(w WriteCloser) {
	var s *Stream
	item, ok := svc.mux.Load(svc.key)
	if !ok {
		s = NewStream()
		svc.mux.Store(svc.key, s)
	} else {
		s = item.(*Stream)
		s.AddWriter(w)
	}
}

func (svc *Service) GetStreams() *sync.Map {
	return svc.mux
}

func (svc *Service) CheckAlive() {
	for {
		<-time.After(5 * time.Second)
		svc.mux.Range(func(key, val interface{}) bool {
			v := val.(*Stream)
			if v.CheckAlive() == 0 {
				svc.mux.Delete(key)
			}
			return true
		})
	}
}
