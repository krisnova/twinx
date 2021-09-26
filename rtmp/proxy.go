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
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/gwuhaolin/livego/configure"

	"github.com/kris-nova/logger"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/protocol/rtmp/core"
)

var (
	STOP_CTRL = "RTMPRELAY_STOP"
)

type RtmpRelay struct {
	PlayUrl              string
	PublishUrl           string
	cs_chan              chan core.ChunkStream
	sndctrl_chan         chan string
	connectPlayClient    *core.ConnClient
	connectPublishClient *core.ConnClient
	startflag            bool
}

func NewRtmpRelay(playurl *string, publishurl *string) *RtmpRelay {
	return &RtmpRelay{
		PlayUrl:              *playurl,
		PublishUrl:           *publishurl,
		cs_chan:              make(chan core.ChunkStream, 500),
		sndctrl_chan:         make(chan string),
		connectPlayClient:    nil,
		connectPublishClient: nil,
		startflag:            false,
	}
}

func (r *RtmpRelay) rcvPlayChunkStream() {
	logger.Debug("rcvPlayRtmpMediaPacket connectClient.Read...")
	for {
		var rc core.ChunkStream

		if r.startflag == false {
			r.connectPlayClient.Close(nil)
			logger.Debug("rcvPlayChunkStream close: playurl=%s, publishurl=%s", r.PlayUrl, r.PublishUrl)
			break
		}
		err := r.connectPlayClient.Read(&rc)

		if err != nil && err == io.EOF {
			break
		}
		//logger.Debug("connectPlayClient.Read return rc.TypeID=%v length=%d, err=%v", rc.TypeID, len(rc.Data), err)
		switch rc.TypeID {
		case 20, 17:
			rr := bytes.NewReader(rc.Data)
			vs, err := r.connectPlayClient.DecodeBatch(rr, amf.AMF0)

			logger.Debug("rcvPlayRtmpMediaPacket: vs=%v, err=%v", vs, err)
		case 18:
			logger.Debug("rcvPlayRtmpMediaPacket: metadata....")
		case 8, 9:
			r.cs_chan <- rc
		}
	}
}

func (r *RtmpRelay) sendPublishChunkStream() {
	for {
		select {
		case rc := <-r.cs_chan:
			//logger.Debug("sendPublishChunkStream: rc.TypeID=%v length=%d", rc.TypeID, len(rc.Data))
			r.connectPublishClient.Write(rc)
		case ctrlcmd := <-r.sndctrl_chan:
			if ctrlcmd == STOP_CTRL {
				r.connectPublishClient.Close(nil)
				logger.Debug("sendPublishChunkStream close: playurl=%s, publishurl=%s", r.PlayUrl, r.PublishUrl)
				return
			}
		}
	}
}

func (r *RtmpRelay) Start() error {
	if r.startflag {
		return fmt.Errorf("The rtmprelay already started, playurl=%s, publishurl=%s\n", r.PlayUrl, r.PublishUrl)
	}

	r.connectPlayClient = core.NewConnClient()
	r.connectPublishClient = core.NewConnClient()

	logger.Debug("play server addr:%v starting....", r.PlayUrl)
	err := r.connectPlayClient.Start(r.PlayUrl, av.PLAY)
	if err != nil {
		logger.Debug("connectPlayClient.Start url=%v error", r.PlayUrl)
		return err
	}

	logger.Debug("publish server addr:%v starting....", r.PublishUrl)
	err = r.connectPublishClient.Start(r.PublishUrl, av.PUBLISH)
	if err != nil {
		logger.Debug("connectPublishClient.Start url=%v error", r.PublishUrl)
		r.connectPlayClient.Close(nil)
		return err
	}

	r.startflag = true
	go r.rcvPlayChunkStream()
	go r.sendPublishChunkStream()

	return nil
}

func (r *RtmpRelay) Stop() {
	if !r.startflag {
		logger.Debug("The rtmprelay already stoped, playurl=%s, publishurl=%s", r.PlayUrl, r.PublishUrl)
		return
	}

	r.startflag = false
	r.sndctrl_chan <- STOP_CTRL
}

type StaticPush struct {
	RtmpUrl       string
	packet_chan   chan *av.Packet
	sndctrl_chan  chan string
	connectClient *core.ConnClient
	startflag     bool
}

var G_StaticPushMap = make(map[string](*StaticPush))
var g_MapLock = new(sync.RWMutex)
var G_PushUrlList []string = nil

var (
	STATIC_RELAY_STOP_CTRL = "STATIC_RTMPRELAY_STOP"
)

func GetStaticPushList(appname string) ([]string, error) {
	if G_PushUrlList == nil {
		// Do not unmarshel the config every time, lots of reflect works -gs
		pushurlList, ok := configure.GetStaticPushUrlList(appname)
		if !ok {
			G_PushUrlList = []string{}
		} else {
			G_PushUrlList = pushurlList
		}
	}

	if len(G_PushUrlList) == 0 {
		return nil, fmt.Errorf("no static push url")
	}

	return G_PushUrlList, nil
}

func GetAndCreateStaticPushObject(rtmpurl string) *StaticPush {
	g_MapLock.RLock()
	staticpush, ok := G_StaticPushMap[rtmpurl]
	logger.Debug("GetAndCreateStaticPushObject: %s, return %v", rtmpurl, ok)
	if !ok {
		g_MapLock.RUnlock()
		newStaticpush := NewStaticPush(rtmpurl)

		g_MapLock.Lock()
		G_StaticPushMap[rtmpurl] = newStaticpush
		g_MapLock.Unlock()

		return newStaticpush
	}
	g_MapLock.RUnlock()

	return staticpush
}

func GetStaticPushObject(rtmpurl string) (*StaticPush, error) {
	g_MapLock.RLock()
	if staticpush, ok := G_StaticPushMap[rtmpurl]; ok {
		g_MapLock.RUnlock()
		return staticpush, nil
	}
	g_MapLock.RUnlock()

	return nil, fmt.Errorf("G_StaticPushMap[%s] not exist....", rtmpurl)
}

func ReleaseStaticPushObject(rtmpurl string) {
	g_MapLock.RLock()
	if _, ok := G_StaticPushMap[rtmpurl]; ok {
		g_MapLock.RUnlock()

		logger.Debug("ReleaseStaticPushObject %s ok", rtmpurl)
		g_MapLock.Lock()
		delete(G_StaticPushMap, rtmpurl)
		g_MapLock.Unlock()
	} else {
		g_MapLock.RUnlock()
		logger.Debug("ReleaseStaticPushObject: not find %s", rtmpurl)
	}
}

func NewStaticPush(rtmpurl string) *StaticPush {
	return &StaticPush{
		RtmpUrl:       rtmpurl,
		packet_chan:   make(chan *av.Packet, 500),
		sndctrl_chan:  make(chan string),
		connectClient: nil,
		startflag:     false,
	}
}

func (s *StaticPush) Start() error {
	if s.startflag {
		return fmt.Errorf("StaticPush already start %s", s.RtmpUrl)
	}

	s.connectClient = core.NewConnClient()

	logger.Debug("static publish server addr:%v starting....", s.RtmpUrl)
	err := s.connectClient.Start(s.RtmpUrl, "publish")
	if err != nil {
		logger.Debug("connectClient.Start url=%v error", s.RtmpUrl)
		return err
	}
	logger.Debug("static publish server addr:%v started, streamid=%d", s.RtmpUrl, s.connectClient.GetStreamId())
	go s.HandleAvPacket()

	s.startflag = true
	return nil
}

func (s *StaticPush) Stop() {
	if !s.startflag {
		return
	}

	logger.Debug("StaticPush Stop: %s", s.RtmpUrl)
	s.sndctrl_chan <- STATIC_RELAY_STOP_CTRL
	s.startflag = false
}

func (s *StaticPush) WriteAvPacket(packet *av.Packet) {
	if !s.startflag {
		return
	}

	s.packet_chan <- packet
}

func (s *StaticPush) sendPacket(p *av.Packet) {
	if !s.startflag {
		return
	}
	var cs core.ChunkStream

	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = s.connectClient.GetStreamId()
	cs.Timestamp = p.TimeStamp
	//cs.Timestamp += v.BaseTimeStamp()

	//log.Printf("Static sendPacket: rtmpurl=%s, length=%d, streamid=%d",
	//	s.RtmpUrl, len(p.Data), cs.StreamID)
	if p.IsVideo {
		cs.TypeID = av.TAG_VIDEO
	} else {
		if p.IsMetadata {
			cs.TypeID = av.TAG_SCRIPTDATAAMF0
		} else {
			cs.TypeID = av.TAG_AUDIO
		}
	}

	s.connectClient.Write(cs)
}

func (s *StaticPush) HandleAvPacket() {
	if !s.IsStart() {
		logger.Debug("static push %s not started", s.RtmpUrl)
		return
	}

	for {
		select {
		case packet := <-s.packet_chan:
			s.sendPacket(packet)
		case ctrlcmd := <-s.sndctrl_chan:
			if ctrlcmd == STATIC_RELAY_STOP_CTRL {
				s.connectClient.Close(nil)
				logger.Debug("Static HandleAvPacket close: publishurl=%s", s.RtmpUrl)
				return
			}
		}
	}
}

func (s *StaticPush) IsStart() bool {
	return s.startflag
}
