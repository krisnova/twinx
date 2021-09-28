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

	"github.com/kris-nova/logger"

	"github.com/gwuhaolin/livego/protocol/amf"
)

var (
	STOP_CTRL = "RTMPRELAY_STOP"
)

type RtmpRelay struct {
	PlayUrl              string
	PublishUrl           string
	cs_chan              chan ChunkStream
	sndctrl_chan         chan string
	connectPlayClient    *ConnClient
	connectPublishClient *ConnClient
	startflag            bool
}

func NewRtmpRelay(playurl *string, publishurl *string) *RtmpRelay {
	return &RtmpRelay{
		PlayUrl:              *playurl,
		PublishUrl:           *publishurl,
		cs_chan:              make(chan ChunkStream, 500),
		sndctrl_chan:         make(chan string),
		connectPlayClient:    nil,
		connectPublishClient: nil,
		startflag:            false,
	}
}

func (r *RtmpRelay) rcvPlayChunkStream() {

	for {
		var rc ChunkStream

		if r.startflag == false {
			r.connectPlayClient.Close(nil)
			logger.Debug("rcvPlayChunkStream close: playurl=%s, publishurl=%s", r.PlayUrl, r.PublishUrl)
			break
		}
		err := r.connectPlayClient.Read(&rc)

		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Warning("reading chunk from proxy server: %v", err)
		}

		// Here is the type implementation.
		// This is defined https://en.wikipedia.org/wiki/Real-Time_Messaging_Protocol

		//Type 0 - Clear Stream: Sent when the connection is established and carries no further data
		//Type 1 - Clear the Buffer.
		//Type 2 - Stream Dry.
		//Type 3 - The client's buffer time. The third parameter holds the value in millisecond.
		//Type 4 - Reset a stream.
		//Type 6 - Ping the client from server. The second parameter is the current time.
		//Type 7 - Pong reply from client. The second parameter is the time when the client receives the Ping.
		//Type 8 - UDP Request.
		//Type 9 - UDP Response.
		//Type 10 - Bandwidth Limit.
		//Type 11 - Bandwidth.
		//Type 12 - Throttle Bandwidth.
		//Type 13 - Stream Created.
		//Type 14 - Stream Deleted.
		//Type 15 - Set Read Access.
		//Type 16 - Set Write Access.
		//Type 17 - Stream Meta Request.
		//Type 18 - Stream Meta Response.
		//Type 19 - Get Segment Boundary.
		//Type 20 - Set Segment Boundary.
		//Type 21 - On Disconnect.
		//Type 22 - Set Critical Link.
		//Type 23 - Disconnect.
		//Type 24 - Hash Update.
		//Type 25 - Hash Timeout.
		//Type 26 - Hash Request.
		//Type 27 - Hash Response.
		//Type 28 - Check Bandwidth.
		//Type 29 - Set Audio Sample Access.
		//Type 30 - Set Video Sample Access.
		//Type 31 - Throttle Begin.
		//Type 32 - Throttle End.
		//Type 33 - DRM Notify.
		//Type 34 - RTMFP Sync.
		//Type 35 - Query IHello.
		//Type 36 - Forward IHello.
		//Type 37 - Redirect IHello.
		//Type 38 - Notify EOF.
		//Type 39 - Proxy Continue.
		//Type 40 - Proxy Remove Upstream.
		//Type 41 - RTMFP Set Keepalives.
		//Type 46 - Segment Not Found.
		switch rc.TypeID {
		case 20, 17:
			rr := bytes.NewReader(rc.Data)
			vs, err := r.connectPlayClient.DecodeBatch(rr, amf.AMF0)
			if err != nil && err != io.EOF {
				logger.Warning("error decoding batch chunk from proxy server: %v", err)
			}
			logger.Debug("Decoding message: %v", vs)
			break
		case 18:
			logger.Debug("Chunk metadata received")
			break
		case 8, 9:
			r.cs_chan <- rc
			break
		default:
			logger.Warning("Unhandled type: %d", rc.TypeID)
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
	logger.Debug("Starting RTMP Relay")
	if r.startflag {
		return fmt.Errorf("The rtmprelay already started, playurl=%s, publishurl=%s\n", r.PlayUrl, r.PublishUrl)
	}

	r.connectPlayClient = NewConnClient()
	r.connectPublishClient = NewConnClient()

	err := r.connectPlayClient.Start(r.PlayUrl, CommandPlay)
	if err != nil {
		logger.Warning("Unable to connect [PLAY] %s %v", r.PlayUrl, err)
		return err
	}

	err = r.connectPublishClient.Start(r.PublishUrl, CommandPublish)
	if err != nil {
		logger.Warning("Unable to connect [PUBLISH] %s %v", r.PublishUrl, err)
		r.connectPlayClient.Close(nil)
		return err
	}

	r.startflag = true
	go r.rcvPlayChunkStream()
	go r.sendPublishChunkStream()

	logger.Success("Bridge Success! [%s] -> [%s]", r.PlayUrl, r.PublishUrl)

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
	packet_chan   chan *Packet
	sndctrl_chan  chan string
	connectClient *ConnClient
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
		pushurlList, ok := GetStaticPushUrlList(appname)
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
		packet_chan:   make(chan *Packet, 500),
		sndctrl_chan:  make(chan string),
		connectClient: nil,
		startflag:     false,
	}
}

func (s *StaticPush) Start() error {
	if s.startflag {
		return fmt.Errorf("StaticPush already start %s", s.RtmpUrl)
	}

	s.connectClient = NewConnClient()

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

func (s *StaticPush) WriteAvPacket(packet *Packet) {
	if !s.startflag {
		return
	}

	s.packet_chan <- packet
}

func (s *StaticPush) sendPacket(p *Packet) {
	if !s.startflag {
		return
	}
	var cs ChunkStream

	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = s.connectClient.GetStreamId()
	cs.Timestamp = p.TimeStamp
	//cs.Timestamp += v.BaseTimeStamp()

	//log.Printf("Static sendPacket: rtmpurl=%s, length=%d, streamid=%d",
	//	s.RtmpUrl, len(p.Data), cs.StreamID)
	if p.IsVideo {
		cs.TypeID = TAG_VIDEO
	} else {
		if p.IsMetadata {
			cs.TypeID = TAG_SCRIPTDATAAMF0
		} else {
			cs.TypeID = TAG_AUDIO
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
