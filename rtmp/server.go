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
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/kris-nova/logger"

	"github.com/gwuhaolin/livego/utils/uid"
)

const (
	// MaximumPacketQueueRecords
	//
	// Similar to how HTTP-based transmission protocols like
	// HLS and DASH behave, RTMP too, breaks a multimedia
	// stream into fragments that are usually:
	//  - 64 bytes for audio
	//  - 128 bytes for video
	//
	// The size of the fragments can be negotiated between the client and the server.
	MaximumPacketQueueRecords int = 1024

	// SaveStaticMediaTimeIntervalMilliseconds
	//
	// The time delay (in MS) between saving static media
	SaveStaticMediaTimeIntervalMilliseconds int64 = 5000

	ReadTimeout  int = 10
	WriteTimeout int = 10
)

type Client struct {
	handler Handler
	getter  GetWriter
}

func NewRtmpClient(h Handler, getter GetWriter) *Client {
	return &Client{
		handler: h,
		getter:  getter,
	}
}

func (c *Client) Dial(url string, method string) error {
	connClient := NewConnClient()
	if err := connClient.Start(url, method); err != nil {
		return err
	}
	if method == CommandPublish {
		writer := NewVirWriter(connClient)
		logger.Info("client Dial call NewVirWriter url=%s, method=%s", url, method)
		c.handler.HandleWriter(writer)
	} else if method == CommandPlay {
		reader := NewVirReader(connClient)
		logger.Info("client Dial call NewVirReader url=%s, method=%s", url, method)
		c.handler.HandleReader(reader)
		if c.getter != nil {
			writer := c.getter.GetWriter(reader.Info())
			c.handler.HandleWriter(writer)
		}
	}
	return nil
}

func (c *Client) GetHandle() Handler {
	return c.handler
}

type Server struct {
	handler Handler
	getter  GetWriter
}

func NewRtmpServer(h Handler, getter GetWriter) *Server {
	return &Server{
		handler: h,
		getter:  getter,
	}
}

func (s *Server) Serve(listener net.Listener) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Critical("rtmp serve panic: ", r)
		}
	}()

	for {
		var netconn net.Conn
		netconn, err = listener.Accept()
		if err != nil {
			return
		}
		conn := NewConn(netconn, 4*1024)
		logger.Info("Client connected!")
		logger.Info("Remote : %s", conn.RemoteAddr().String())
		logger.Info("Local  : %s", conn.LocalAddr().String())
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *Conn) error {
	if err := conn.HandshakeServer(); err != nil {
		conn.Close()
		logger.Critical("RTMP Handshake: %v", err)
		return err
	}

	connServer := NewConnServer(conn)

	if err := connServer.ReadMsg(); err != nil {
		conn.Close()
		logger.Critical("RTMP Read Message: %v", err)
		return err
	}

	appname, name, _ := connServer.GetInfo()
	logger.Debug("Client %s handshake success", appname)

	//if Config.GetBool("rtmp_noauth") {
	// Default rtmp_noauth
	key, err := RoomKeys.GetKey(name)
	if err != nil {
		err := fmt.Errorf("Cannot create key err=%s", err.Error())
		conn.Close()
		logger.Critical("GetKey err: ", err)
		return err
	}
	name = key
	//}
	channel, err := RoomKeys.GetChannel(name)
	if err != nil {
		err := fmt.Errorf("invalid key err=%s", err.Error())
		conn.Close()
		logger.Critical("CheckKey err: ", err)
		return err
	}
	connServer.PublishInfo.Name = channel

	reader := NewVirReader(connServer)
	s.handler.HandleReader(reader)
	logger.Info("New publisher: %s", reader.Info().URL)

	if s.getter != nil {
		writeType := reflect.TypeOf(s.getter)
		logger.Info("Setting writeType: %v", writeType)
		writer := s.getter.GetWriter(reader.Info())
		s.handler.HandleWriter(writer)
	}
	//flvWriter := new(FlvDvr)
	//s.handler.HandleWriter(flvWriter.GetWriter(reader.Info()))

	return nil
}

type GetInFo interface {
	GetInfo() (string, string, string)
}

type StreamReadWriteCloser interface {
	GetInFo
	Close(error)
	Write(ChunkStream) error
	Read(c *ChunkStream) error
}

type StaticsBW struct {
	StreamId               uint32
	VideoDatainBytes       uint64
	LastVideoDatainBytes   uint64
	VideoSpeedInBytesperMS uint64

	AudioDatainBytes       uint64
	LastAudioDatainBytes   uint64
	AudioSpeedInBytesperMS uint64

	LastTimestamp int64
}

type VirWriter struct {
	Uid    string
	closed bool
	RWBaser
	conn        StreamReadWriteCloser
	packetQueue chan *Packet
	WriteBWInfo StaticsBW
}

func NewVirWriter(conn StreamReadWriteCloser) *VirWriter {
	ret := &VirWriter{
		Uid:         uid.NewId(),
		conn:        conn,
		RWBaser:     NewRWBaser(time.Second * time.Duration(WriteTimeout)),
		packetQueue: make(chan *Packet, MaximumPacketQueueRecords),
		WriteBWInfo: StaticsBW{0, 0, 0, 0, 0, 0, 0, 0},
	}

	go ret.Check()
	go func() {
		err := ret.SendPacket()
		if err != nil {
			logger.Debug("Dropped packet(s). Possible closed connection: %v", err)
		}
	}()
	return ret
}

func (v *VirWriter) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
	nowInMS := int64(time.Now().UnixNano() / 1e6)

	v.WriteBWInfo.StreamId = streamid
	if isVideoFlag {
		v.WriteBWInfo.VideoDatainBytes = v.WriteBWInfo.VideoDatainBytes + length
	} else {
		v.WriteBWInfo.AudioDatainBytes = v.WriteBWInfo.AudioDatainBytes + length
	}

	if v.WriteBWInfo.LastTimestamp == 0 {
		v.WriteBWInfo.LastTimestamp = nowInMS
	} else if (nowInMS - v.WriteBWInfo.LastTimestamp) >= SaveStaticMediaTimeIntervalMilliseconds {
		diffTimestamp := (nowInMS - v.WriteBWInfo.LastTimestamp) / 1000

		v.WriteBWInfo.VideoSpeedInBytesperMS = (v.WriteBWInfo.VideoDatainBytes - v.WriteBWInfo.LastVideoDatainBytes) * 8 / uint64(diffTimestamp) / 1000
		v.WriteBWInfo.AudioSpeedInBytesperMS = (v.WriteBWInfo.AudioDatainBytes - v.WriteBWInfo.LastAudioDatainBytes) * 8 / uint64(diffTimestamp) / 1000

		v.WriteBWInfo.LastVideoDatainBytes = v.WriteBWInfo.VideoDatainBytes
		v.WriteBWInfo.LastAudioDatainBytes = v.WriteBWInfo.AudioDatainBytes
		v.WriteBWInfo.LastTimestamp = nowInMS
	}
}

func (v *VirWriter) Check() {
	var c ChunkStream
	for {
		if err := v.conn.Read(&c); err != nil {
			v.Close(err)
			return
		}
	}
}

func (v *VirWriter) DropPacket(pktQue chan *Packet, info Info) {
	logger.Critical("packet queue max [%+v]", info)
	for i := 0; i < MaximumPacketQueueRecords-84; i++ {
		tmpPkt, ok := <-pktQue
		// try to don't drop audio
		if ok && tmpPkt.IsAudio {
			if len(pktQue) > MaximumPacketQueueRecords-2 {
				logger.Info("drop audio pkt")
				<-pktQue
			} else {
				pktQue <- tmpPkt
			}

		}

		if ok && tmpPkt.IsVideo {
			videoPkt, ok := tmpPkt.Header.(VideoPacketHeader)
			// dont't drop sps config and dont't drop key frame
			if ok && (videoPkt.IsSeq() || videoPkt.IsKeyFrame()) {
				pktQue <- tmpPkt
			}
			if len(pktQue) > MaximumPacketQueueRecords-10 {
				logger.Info("drop video pkt")
				<-pktQue
			}
		}

	}
	logger.Info("packet queue len: ", len(pktQue))
}

//
func (v *VirWriter) Write(p *Packet) (err error) {
	err = nil

	if v.closed {
		err = fmt.Errorf("VirWriter closed")
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("VirWriter has already been closed:%v", e)
		}
	}()
	if len(v.packetQueue) >= MaximumPacketQueueRecords-24 {
		v.DropPacket(v.packetQueue, v.Info())
	} else {
		v.packetQueue <- p
	}

	return
}

func (v *VirWriter) SendPacket() error {
	Flush := reflect.ValueOf(v.conn).MethodByName("Flush")
	var cs ChunkStream
	for {
		p, ok := <-v.packetQueue
		if ok {
			cs.Data = p.Data
			cs.Length = uint32(len(p.Data))
			cs.StreamID = p.StreamID
			cs.Timestamp = p.TimeStamp
			cs.Timestamp += v.BaseTimeStamp()

			if p.IsVideo {
				cs.TypeID = TAG_VIDEO
			} else {
				if p.IsMetadata {
					cs.TypeID = TAG_SCRIPTDATAAMF0
				} else {
					cs.TypeID = TAG_AUDIO
				}
			}

			v.SaveStatics(p.StreamID, uint64(cs.Length), p.IsVideo)
			v.SetPreTime()
			v.RecTimeStamp(cs.Timestamp, cs.TypeID)
			err := v.conn.Write(cs)
			if err != nil {
				v.closed = true
				return err
			}
			Flush.Call(nil)
		} else {
			return fmt.Errorf("closed")
		}

	}
}

func (v *VirWriter) Info() (ret Info) {
	ret.UID = v.Uid
	_, _, URL := v.conn.GetInfo()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		logger.Warning("parsing URL: %v", err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	ret.Inter = true
	return
}

func (v *VirWriter) Close(err error) {
	logger.Warning("Client connection closed: %v", err)
	if !v.closed {
		close(v.packetQueue)
	}
	v.closed = true
	v.conn.Close(err)
}

type VirReader struct {
	Uid string
	RWBaser
	demuxer    *FLVDemuxer
	conn       StreamReadWriteCloser
	ReadBWInfo StaticsBW
}

func NewVirReader(conn StreamReadWriteCloser) *VirReader {
	return &VirReader{
		Uid:        uid.NewId(),
		conn:       conn,
		RWBaser:    NewRWBaser(time.Second * time.Duration(WriteTimeout)),
		demuxer:    NewFLVDemuxer(),
		ReadBWInfo: StaticsBW{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (v *VirReader) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
	nowInMS := int64(time.Now().UnixNano() / 1e6)

	v.ReadBWInfo.StreamId = streamid
	if isVideoFlag {
		v.ReadBWInfo.VideoDatainBytes = v.ReadBWInfo.VideoDatainBytes + length
	} else {
		v.ReadBWInfo.AudioDatainBytes = v.ReadBWInfo.AudioDatainBytes + length
	}

	if v.ReadBWInfo.LastTimestamp == 0 {
		v.ReadBWInfo.LastTimestamp = nowInMS
	} else if (nowInMS - v.ReadBWInfo.LastTimestamp) >= SaveStaticMediaTimeIntervalMilliseconds {
		diffTimestamp := (nowInMS - v.ReadBWInfo.LastTimestamp) / 1000

		//log.Printf("now=%d, last=%d, diff=%d", nowInMS, v.ReadBWInfo.LastTimestamp, diffTimestamp)
		v.ReadBWInfo.VideoSpeedInBytesperMS = (v.ReadBWInfo.VideoDatainBytes - v.ReadBWInfo.LastVideoDatainBytes) * 8 / uint64(diffTimestamp) / 1000
		v.ReadBWInfo.AudioSpeedInBytesperMS = (v.ReadBWInfo.AudioDatainBytes - v.ReadBWInfo.LastAudioDatainBytes) * 8 / uint64(diffTimestamp) / 1000

		v.ReadBWInfo.LastVideoDatainBytes = v.ReadBWInfo.VideoDatainBytes
		v.ReadBWInfo.LastAudioDatainBytes = v.ReadBWInfo.AudioDatainBytes
		v.ReadBWInfo.LastTimestamp = nowInMS
	}
}

func (v *VirReader) Read(p *Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warning("rtmp read packet panic: ", r)
		}
	}()

	v.SetPreTime()
	var cs ChunkStream
	for {
		err = v.conn.Read(&cs)
		if err != nil {
			return err
		}
		if cs.TypeID == TAG_AUDIO ||
			cs.TypeID == TAG_VIDEO ||
			cs.TypeID == TAG_SCRIPTDATAAMF0 ||
			cs.TypeID == TAG_SCRIPTDATAAMF3 {
			break
		}
	}

	p.IsAudio = cs.TypeID == TAG_AUDIO
	p.IsVideo = cs.TypeID == TAG_VIDEO
	p.IsMetadata = cs.TypeID == TAG_SCRIPTDATAAMF0 || cs.TypeID == TAG_SCRIPTDATAAMF3
	p.StreamID = cs.StreamID
	p.Data = cs.Data
	p.TimeStamp = cs.Timestamp

	v.SaveStatics(p.StreamID, uint64(len(p.Data)), p.IsVideo)
	v.demuxer.DemuxH(p)
	return err
}

func (v *VirReader) Info() (ret Info) {
	ret.UID = v.Uid
	_, _, URL := v.conn.GetInfo()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		logger.Warning("parsing URL: %v", err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	return
}

func (v *VirReader) Close(err error) {
	logger.Warning("Connection closed: %v", err)
	v.conn.Close(err)
}
