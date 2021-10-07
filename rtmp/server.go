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
	"encoding/binary"
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

type Server struct {
	service  *Service
	listener *Listener
}

func NewServer() *Server {
	return &Server{
		service: NewService(),
	}
}

func (s *Server) ListenAndServe(raw string) error {
	l, err := Listen(raw)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) Serve(listener net.Listener) error {
	// Translate a Go net.Listener to an RTMP net.Listener
	var concrete *Listener
	if l, ok := listener.(*Listener); !ok {
		l, err := newFromNetListener(listener)
		if err != nil {
			return fmt.Errorf("creating RTMP listener: %v", err)
		}
		concrete = l
	} else {
		concrete = l
	}
	s.listener = concrete
	urlAddr, err := NewURLAddr(s.listener.Addr().String())
	if err != nil {
		return fmt.Errorf("urlAddr: %v", err)
	}
	s.listener.addr = urlAddr
	logger.Info("Listening %s...", s.listener.addr.SafeURL())
	for {
		clientConn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("client conn accept: %v", err)
		}
		go s.handleConn(clientConn)
	}
	return nil
}

func (s *Server) handleConn(netConn net.Conn) error {

	// Translate net.Conn -> rtmp.Conn
	conn := NewConn(netConn)

	if err := conn.HandshakeServer(); err != nil {
		conn.Close()
		logger.Critical("RTMP Handshake: %v", err)
		return err
	}

	connSrv := NewConnServer(conn)
	//logger.Debug("Stream ID: %d", connSrv.streamID)
	logger.Debug("Transaction ID: %d", connSrv.transactionID)

	for {
		if connSrv.IsPublisher() {
			// Once we are connected plumb the stream through
			logger.Debug("Stream ID: %d", connSrv.streamID)
			logger.Debug("Transaction ID: %d", connSrv.transactionID)

			// **************************************
			// Hér vera drekar
			// **************************************
			//
			// So here is where I am temporarily
			// stopping my refactoring of this server
			// code.
			//
			// Ideally we do NOT have to "break" here.
			// We can clean our code up by having
			// the client responses funnel through
			// this main code point.
			//
			// The underlying implementation is how
			// we manage multiplexing onto the various
			// internal memory pools for each stream.
			//
			// Although I WANT to refactor this.
			// I will not be refactoring this right
			// now.
			//
			// **************************************
			reader := NewVirtualReader(connSrv)
			s.service.HandleReader(reader)

			// TODO: Do NOT break here
			break
		}
		x, err := connSrv.ReadPacket()
		if err != nil {
			logger.Critical("reading chunk from client: %v", err)
		}
		//logger.Debug("Message received from client: %s", typeIDString(chunk))

		switch x.TypeID {
		case SetChunkSizeMessageID:
			// 5.4.1. Set Chunk Size (1)
			logger.Info("Message: SetChunkSize")
			chunkSize := binary.BigEndian.Uint32(x.Data)
			logger.Info("   Setting remoteChunkSize: %d", chunkSize)
			conn.remoteChunkSize = chunkSize
			conn.ack(x.Length)
		case AbortMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case AcknowledgementMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case WindowAcknowledgementSizeMessageID:
			logger.Info("Message: WindowAcknowledgementSize")
			ackSize := binary.BigEndian.Uint32(x.Data)
			logger.Info("   Setting windowAcknowledgementSize: %d", ackSize)
			conn.remoteWindowAckSize = ackSize
		case SetPeerBandwidthMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case UserControlMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case CommandMessageAMF0ID, CommandMessageAMF3ID:
			// Handle the command message
			// Note: There are sub-command messages logged in the next method
			err := connSrv.messageCommand(x)
			if err != nil {
				logger.Critical("command message: %v", err)
			}
		case DataMessageAMF0ID, DataMessageAMF3ID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case SharedObjectMessageAMF0ID, SharedObjectMessageAMF3ID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case AudioMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case VideoMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		case AggregateMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(x))
		default:
			logger.Critical("unsupported messageID: %s", typeIDString(x))

		}
	}

	//writer := NewVirtualWriter(connSrv)
	//s.service.HandleWriter(writer)

	return nil
}

type GetInfo interface {
	GetInfo() (string, string, string)
}

type StreamReadWriteCloser interface {
	GetInfo
	Close()
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

type VirtualWriter struct {
	Uid    string
	closed bool
	RWBaser
	conn        StreamReadWriteCloser
	packetQueue chan *Packet
	WriteBWInfo StaticsBW
}

func NewVirtualWriter(conn StreamReadWriteCloser) *VirtualWriter {
	ret := &VirtualWriter{
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

func (v *VirtualWriter) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
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

func (v *VirtualWriter) Check() {
	var c ChunkStream
	for {
		if err := v.conn.Read(&c); err != nil {
			v.Close()
			return
		}
	}
}

func (v *VirtualWriter) DropPacket(pktQue chan *Packet, info Info) {
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
func (v *VirtualWriter) Write(p *Packet) (err error) {
	err = nil

	if v.closed {
		err = fmt.Errorf("VirtualWriter closed")
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("VirtualWriter has already been closed:%v", e)
		}
	}()
	if len(v.packetQueue) >= MaximumPacketQueueRecords-24 {
		v.DropPacket(v.packetQueue, v.Info())
	} else {
		v.packetQueue <- p
	}

	return
}

func (v *VirtualWriter) SendPacket() error {
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

func (v *VirtualWriter) Info() (ret Info) {
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

func (v *VirtualWriter) Close() {
	if !v.closed {
		close(v.packetQueue)
	}
	v.closed = true
	v.conn.Close()
}

type VirtualReader struct {
	Uid string
	RWBaser
	demuxer    *FLVDemuxer
	conn       StreamReadWriteCloser
	ReadBWInfo StaticsBW
}

func NewVirtualReader(conn StreamReadWriteCloser) *VirtualReader {
	return &VirtualReader{
		Uid:        uid.NewId(),
		conn:       conn,
		RWBaser:    NewRWBaser(time.Second * time.Duration(WriteTimeout)),
		demuxer:    NewFLVDemuxer(),
		ReadBWInfo: StaticsBW{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (v *VirtualReader) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
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

func (v *VirtualReader) Read(p *Packet) (err error) {

	defer func() {
		if r := recover(); r != nil {
			logger.Critical("Recovered panic: ", r)
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

func (v *VirtualReader) Info() (ret Info) {
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

func (v *VirtualReader) Close() {
	v.conn.Close()
}
