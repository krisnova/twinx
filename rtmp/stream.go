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
	"strings"
	"sync"

	"github.com/kris-nova/logger"
)

type Stream struct {
	isStart bool
	cache   *Cache
	r       ReadCloser
	ws      *sync.Map
	info    Info
}

func NewStream() *Stream {
	return &Stream{
		cache: NewCache(),
		ws:    &sync.Map{},
	}
}

func (s *Stream) ID() string {
	if s.r != nil {
		return s.r.Info().UID
	}
	return ""
}

func (s *Stream) GetReader() ReadCloser {
	return s.r
}

func (s *Stream) GetWs() *sync.Map {
	return s.ws
}

func (s *Stream) Copy(dst *Stream) {
	dst.info = s.info
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
	go s.TransStart()
}

func (s *Stream) AddWriter(w WriteCloser) {
	info := w.Info()
	pw := &PackWriterCloser{w: w}
	s.ws.Store(info.UID, pw)
}

func (s *Stream) StartStaticPush() {
	key := s.info.Key

	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return
	}

	streamname := key[index+1:]
	appname := dscr[0]

	//logger.Info("StartStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		logger.Info("Static Push: %v", err)
		return
	}

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		//logger.Info("StartStaticPush: static pushurl=%s", pushurl)

		staticpushObj := GetAndCreateStaticPushObject(pushurl)
		if staticpushObj != nil {
			if err := staticpushObj.Start(); err != nil {
				logger.Info("StartStaticPush: staticpushObj.Start %s error= v", pushurl, err)
			} else {
				logger.Info("StartStaticPush: staticpushObj.Start %s ok", pushurl)
			}
		} else {
			logger.Info("StartStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
}

func (s *Stream) StopStaticPush() {
	key := s.info.Key

	logger.Info("StopStaticPush......%s", key)
	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return
	}

	streamname := key[index+1:]
	appname := dscr[0]

	logger.Info("StopStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		logger.Info("StopStaticPush: GetStaticPushList error=%v", err)
		return
	}

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		logger.Info("StopStaticPush: static pushurl=%s", pushurl)

		staticpushObj, err := GetStaticPushObject(pushurl)
		if (staticpushObj != nil) && (err == nil) {
			staticpushObj.Stop()
			ReleaseStaticPushObject(pushurl)
			logger.Info("StopStaticPush: staticpushObj.Stop %s ", pushurl)
		} else {
			logger.Info("StopStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
}

func (s *Stream) IsSendStaticPush() bool {
	key := s.info.Key

	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return false
	}

	appname := dscr[0]

	//logger.Info("SendStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		//logger.Info("SendStaticPush: GetStaticPushList error=%v", err)
		return false
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return false
	}

	streamname := key[index+1:]

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		//logger.Info("SendStaticPush: static pushurl=%s", pushurl)

		staticpushObj, err := GetStaticPushObject(pushurl)
		if (staticpushObj != nil) && (err == nil) {
			return true
			//staticpushObj.WriteAvPacket(&packet)
			//logger.Info("SendStaticPush: WriteAvPacket %s ", pushurl)
		} else {
			logger.Info("SendStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
	return false
}

func (s *Stream) SendStaticPush(packet Packet) {
	key := s.info.Key

	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return
	}

	streamname := key[index+1:]
	appname := dscr[0]

	//logger.Info("SendStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		//logger.Info("SendStaticPush: GetStaticPushList error=%v", err)
		return
	}

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		//logger.Info("SendStaticPush: static pushurl=%s", pushurl)

		staticpushObj, err := GetStaticPushObject(pushurl)
		if (staticpushObj != nil) && (err == nil) {
			staticpushObj.WriteAvPacket(&packet)
			//logger.Info("SendStaticPush: WriteAvPacket %s ", pushurl)
		} else {
			logger.Info("SendStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
}

func (s *Stream) TransStart() {
	logger.Info("Starting stream transaction")
	s.isStart = true
	var p Packet

	//s.StartStaticPush()

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

		if s.IsSendStaticPush() {
			s.SendStaticPush(p)
		}

		s.cache.Write(p)

		s.ws.Range(func(key, val interface{}) bool {
			v := val.(*PackWriterCloser)
			if !v.init {
				//logger.Info("cache.send: %v", v.w.Info())
				if err = s.cache.Send(v.w); err != nil {
					logger.Info("[%s] send cache packet error: %v, remove", v.w.Info(), err)
					s.ws.Delete(key)
					return true
				}
				v.init = true
			} else {
				newPacket := p
				//writeType := reflect.TypeOf(v.w)
				//logger.Info("w.Write: type=%v, %v", writeType, v.w.Info())
				if err = v.w.Write(&newPacket); err != nil {
					logger.Info("[%s] write packet error: %v, remove", v.w.Info(), err)
					s.ws.Delete(key)
				}
			}
			return true
		})
	}
}

func (s *Stream) TransStop() {
	logger.Info("TransStop: %s", s.info.Key)

	if s.isStart && s.r != nil {
		s.r.Close(fmt.Errorf("stopping existing stream"))
	}

	s.isStart = false
}

func (s *Stream) CheckAlive() (n int) {
	if s.r != nil && s.isStart {
		if s.r.Alive() {
			n++
		} else {
			s.r.Close(fmt.Errorf("read timeout"))
		}
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			//Alive from RWBaser, check last frame now - timestamp, if > timeout then Remove it
			if !v.w.Alive() {
				logger.Info("write timeout remove")
				s.ws.Delete(key)
				v.w.Close(fmt.Errorf("write timeout"))
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
		logger.Warning("Publisher closed: %s", s.r.Info().UID)
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			v.w.Close(fmt.Errorf("closed"))
			if v.w.Info().IsInterval() {
				s.ws.Delete(key)
				logger.Info("[%v] player closed and remove\n", v.w.Info())
			}
		}
		return true
	})
}
