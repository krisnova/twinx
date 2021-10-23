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
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/kris-nova/logger"
)

type ServerClientType int

const (
	PlayClient    ServerClientType = 1
	PublishClient ServerClientType = 2
)

type ServerConn struct {

	// All server clients are either proxy, publish, or play clients
	clientType ServerClientType

	// conn is the base conn for all RTMP members (both clients, and servers)
	conn *Conn

	// transactionID is passed around
	// with the packets to/from a compliant RTMP member
	//
	// This can be reset to 0, and should increment at times
	transactionID int64

	// connectInfo is a well-known RTMP object that is passed
	// around with compliant RTMP members
	connectInfo *ConnectInfo

	// connectPacket is the single *ChunkStream packet
	// discovered when a client calls connect
	connectPacket *ChunkStream

	// publishInfo will be set if this connection is an RTMP publish
	// client
	publishInfo *PublishInfo

	metaData      *MetaData
	metaDataChunk *ChunkStream

	decoder *amf.Decoder
	encoder *amf.Encoder
	bytesw  *bytes.Buffer

	// server is a pointer back to the main server instance
	server *Server
}

func NewServerConn(conn *Conn) *ServerConn {
	return &ServerConn{
		conn:    conn,
		bytesw:  bytes.NewBuffer(nil),
		decoder: &amf.Decoder{},
		encoder: &amf.Encoder{},
	}
}

//func (s *ServerConn) getStreamBeginPacket() *ChunkStream {
//	return &ChunkStream{}
//}
//
//func (s *ServerConn) getMetaDataPacket() *ChunkStream {
//	return &ChunkStream{}
//}

// NextChunk will read the next packet of data from the client,
// and will attempt to respond to the packet based on it's content and
// the appropriate response per the RTMP spec.
func (s *ServerConn) NextChunk() (*ChunkStream, error) {
	var chunk ChunkStream
	if err := s.conn.Read(&chunk); err != nil {
		return nil, fmt.Errorf("reading chunk from client: %v", err)
	}
	return &chunk, nil
}

var clientLen int = 0

// RoutePackets will hang and route packets for this connection
func (s *ServerConn) RoutePackets() error {
	for {

		x, err := s.NextChunk()
		if err != nil {
			// Terminate the client!
			if err != TestableEOFError {
				return err
			}
		}
		err = s.Route(x)
		if err != nil {
			logger.Critical(err.Error())
		}

	}
	return nil
}

func (s *ServerConn) Route(x *ChunkStream) error {
	switch x.TypeID {
	case SetChunkSizeMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		chunkSize := binary.BigEndian.Uint32(x.Data)
		s.conn.remoteChunkSize = chunkSize
		logger.Debug(rtmpMessage(typeIDString(x), ack))
		s.conn.ack(x.Length)
	case AbortMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(x))
	case AcknowledgementMessageID:
		// Acks are acks we can ignore them
	case WindowAcknowledgementSizeMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		ackSize := binary.BigEndian.Uint32(x.Data)
		s.conn.remoteWindowAckSize = ackSize
		logger.Debug(rtmpMessage(typeIDString(x), ack))
		s.conn.ack(x.Length)
	case SetPeerBandwidthMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(x))
	case UserControlMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		return s.handleUserControl(x)
	case CommandMessageAMF0ID, CommandMessageAMF3ID:
		// Note: There are sub-command messages logged in the next method
		return s.handleCommand(x)
	case DataMessageAMF0ID, DataMessageAMF3ID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		return s.handleDataMessage(x)
	case SharedObjectMessageAMF0ID, SharedObjectMessageAMF3ID:
		logger.Critical("unsupported messageID: %s", typeIDString(x))
	case AudioMessageID:
		err := Multiplex(s.server.listener.URLAddr().Key()).Write(x)
		if err != nil {
			return err
		}
	case VideoMessageID:
		err := Multiplex(s.server.listener.URLAddr().Key()).Write(x)
		if err != nil {
			return err
		}
	case AggregateMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(x))
	default:
		logger.Critical("unsupported messageID: %s", typeIDString(x))

	}
	return nil
}

func (s *ServerConn) handleUserControl(x *ChunkStream) error {
	return nil
}

// Example from OBS
//2021-10-13T10:48:38-07:00 [Debug     ]    [0] (@setDataFrame)
//2021-10-13T10:48:38-07:00 [Debug     ]    [1] (onMetaData)
//2021-10-13T10:48:38-07:00 [Debug     ]    [2] (map[2.1:false 3.1:false 4.0:false 4.1:false 5.1:false 7.1:false audiochannels:2 audiocodecid:10 audiodatarate:160 audiosamplerate:48000 audiosamplesize:16 duration:0 encoder:obs-output module (libobs version 27.0.1-3) fileSize:0 framerate:30 height:720 stereo:true videocodecid:7 videodatarate:2500 width:1280])
func (s *ServerConn) handleDataMessage(x *ChunkStream) error {
	amfType := amf.AMF0
	if x.TypeID == CommandMessageAMF3ID {
		// Arithmetic to match AMF3 encoding
		amfType = amf.AMF3
		x.Data = x.Data[1:]
	}
	r := bytes.NewReader(x.Data)
	vs, err := s.LogDecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}

	// set batchedValues
	x.batchedValues = vs

	if len(x.batchedValues) != 3 {
		return fmt.Errorf("unsupported metadata packet size: 3")
	}

	// We assume field [2] is our MetaData object
	metaData, err := MetaDataMapToInstance(x.batchedValues[2])
	if err != nil {
		return fmt.Errorf("unbale to map metadata: %v", err)
	}
	s.metaData = metaData

	// Multiplex (and cache) the metadata for later
	Multiplex(s.server.listener.URLAddr().Key()).AddMetaData(x)
	err = Multiplex(s.server.listener.URLAddr().Key()).Write(x)
	if err != nil {
		return err
	}

	return nil
}

// routeCommand is a sub-router for any of the command messages.
// the server router will receive a set of requests from the client
// the commands can be unordered and of different type
//
// this is the main router for all of these commands that start out
// as an unknown interface
func (s *ServerConn) routeCommand(commandName string, x *ChunkStream) error {
	switch commandName {
	case CommandConnect:
		//logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", CommandConnect), rx))
		return s.connectRX(x)
	case CommandCreateStream:
		//logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", CommandCreateStream), rx))
		return s.createStreamRX(x)
	case CommandPublish:

		// Respond to a publish
		err := s.publishRX(x)
		if err != nil {
			return err
		}

		// Publish client
		s.clientType = PublishClient

		logger.Info(rtmpMessage("Publish Stream", stream))
	case CommandPlay:

		// Respond to a play
		err := s.playRX(x)
		if err != nil {
			return err
		}

		// Play Client
		s.clientType = PlayClient

		// Metrics
		M().Lock()
		P(s.server.listener.URLAddr().Key()).ProxyAddrTX = s.server.listener.URLAddr().SafeURL()
		P(s.server.listener.URLAddr().Key()).ProxyKeyHash = s.server.listener.URLAddr().SafeKey()
		M().Unlock()

		// Add the play client as a backend to Write() to
		Multiplex(s.server.listener.URLAddr().Key()).AddConn(s.conn)

		// Play clients get a streamBegin
		err = Multiplex(s.server.listener.URLAddr().Key()).Write(s.conn.streamBegin())
		if err != nil {
			return err
		}

		err = Multiplex(s.server.listener.URLAddr().Key()).Write(s.metaDataChunk)
		if err != nil {
			return err
		}

		logger.Info(rtmpMessage("Play Stream", stream))
	case CommandFCPublish:
		return s.oosFCPublishRX(x)
	case CommandReleaseStream:
		return s.oosReleaseStreamRX(x)
	case CommandGetStreamLength:
		return s.oosGetStreamLengthRX(x)
	default:
		return fmt.Errorf("unsupported commandName: %s", commandName)
	}
	return nil
}

//  Generate 'getStreamLength' call and send it to the server. If the server
//  knows the duration of the selected stream, it will reply with the duration
//  in seconds.
//
// For Twinx we do not need to respond.
func (s *ServerConn) oosGetStreamLengthRX(x *ChunkStream) error {
	//logger.Info(rtmpMessage(thisFunctionName(), ack))
	return nil
}

func (s *ServerConn) handleCommand(x *ChunkStream) error {
	amfType := amf.AMF0
	if x.TypeID == CommandMessageAMF3ID {
		// Arithmetic to match AMF3 encoding
		amfType = amf.AMF3
		x.Data = x.Data[1:]
	}
	r := bytes.NewReader(x.Data)

	// enable logging here (or in the logger...)
	//vs, err := s.decoder.DecodeBatch(r, amf.Version(amfType))
	vs, err := s.LogDecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}

	// set batchedValues
	x.batchedValues = vs

	// We assume the first message is the name, and in array location 0
	// Validate this before anything else.
	if len(vs) <= 1 {
		return errors.New("decoder failure: unable to decode from protocol")
	}
	commandName, ok := vs[0].(string)
	if !ok {
		return errors.New("decoder failure: unable to render command name")
	}
	return s.routeCommand(commandName, x)
}

const (
	CommandConnectWellKnownID float64 = 1
)

func (s *ServerConn) Write(x *ChunkStream) error {
	if x.TypeID == TAG_SCRIPTDATAAMF0 ||
		x.TypeID == TAG_SCRIPTDATAAMF3 {
		var err error
		if x.Data, err = amf.MetaDataReform(x.Data, amf.DEL); err != nil {
			return err
		}
		x.Length = uint32(len(x.Data))
	}
	M().Lock()
	P(s.server.listener.URLAddr().Key()).ProxyTotalBytesTX = P(s.server.listener.URLAddr().Key()).ProxyTotalBytesTX + int(x.Length)
	P(s.server.listener.URLAddr().Key()).ProxyTotalPacketsTX++
	M().Unlock()
	return s.conn.Write(x)
}

func (s *ServerConn) Flush() error {
	return s.conn.Flush()
}

func (s *ServerConn) Read(packet *ChunkStream) (err error) {
	return s.conn.Read(packet)
}

func (s *ServerConn) LogDecodeBatch(r io.Reader, ver amf.Version) ([]interface{}, error) {

	vs, err := s.decoder.DecodeBatch(r, ver)
	for k, v := range vs {
		logger.Debug("  [%+v] (%+v)", k, v)
	}
	return vs, err
}

func (s *ServerConn) Close() {
	s.conn.Close()
}

func (s *ServerConn) writeMsg(csid, streamID uint32, args ...interface{}) error {
	s.bytesw.Reset()
	for _, v := range args {
		if _, err := s.encoder.Encode(s.bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := s.bytesw.Bytes()
	packet := ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  streamID,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	s.conn.Write(&packet)
	return s.conn.Flush()
}
