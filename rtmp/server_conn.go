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

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/kris-nova/logger"
)

type ServerConn struct {
	done          bool
	streamID      int
	isPublisher   bool
	isConnected   bool
	conn          *Conn
	transactionID int
	ConnInfo      ConnectInfo
	PublishInfo   PublishInfo
	decoder       *amf.Decoder
	encoder       *amf.Encoder
	bytesw        *bytes.Buffer
}

func NewServerConn(conn *Conn) *ServerConn {
	return &ServerConn{
		conn:     conn,
		streamID: 1,
		bytesw:   bytes.NewBuffer(nil),
		decoder:  &amf.Decoder{},
		encoder:  &amf.Encoder{},
	}
}

// ReadPacket will read the next packet of data from the client,
// and will attempt to respond to the packet based on it's content and
// the appropriate response per the RTMP spec.
func (s *ServerConn) ReadPacket() (*ChunkStream, error) {
	var chunk ChunkStream
	if err := s.conn.Read(&chunk); err != nil {
		return nil, fmt.Errorf("reading chunk from client: %v", err)
	}
	return &chunk, nil
}

func (s *ServerConn) messageCommand(packet *ChunkStream) error {
	amfType := amf.AMF0
	if packet.TypeID == CommandMessageAMF3ID {
		// Arithmetic to match AMF3 encoding
		amfType = amf.AMF3
		packet.Data = packet.Data[1:]
	}
	r := bytes.NewReader(packet.Data)
	vs, err := s.decoder.DecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}

	//logger.Debug("   Raw Command Message from Client: %#v", vs)
	// []interface {}{"connect", 1, amf.Object{"app":"twinx", "flashVer":"FMLE/3.0 (compatible; FMSc/1.0)", "swfUrl":"rtmp://localhost:1935/twinx", "tcUrl":"rtmp://localhost:1935/twinx", "type":"nonprivate"}}
	// []interface {}{"releaseStream", 2, interface {}(nil), "1234"}
	// []interface {}{"FCPublish", 3, interface {}(nil), "1234"}
	// []interface {}{"createStream", 4, interface {}(nil)}
	// []interface {}{"publish", 5, interface {}(nil), "1234", "live"}
	switch vs[0].(type) {
	case string:
		switch vs[0].(string) {
		case CommandConnect:
			logger.Info("Command: %s ", CommandConnect)
			if err = s.messageCommandConnect(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: connect")
			if err = s.messageCommandConnectResponse(packet); err != nil {
				return err
			}
			s.isConnected = true
		case CommandCreateStream:
			logger.Info("Command: %s", CommandCreateStream)
			if err = s.messageCommandCreateStream(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: createStream")
			if err = s.messageCommandCreateStreamResponse(packet); err != nil {
				return err
			}
			s.isConnected = true
		case CommandPublish:
			logger.Info("Command: %s", CommandPublish)
			if err = s.messageCommandPlayPublish(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: publish")
			if err = s.messageCommandPublishResponse(packet); err != nil {
				return err
			}
			s.isConnected = true
			s.isPublisher = true
		case CommandPlay:
			logger.Info("Command: %s", CommandPlay)
			if err = s.messageCommandPlayPublish(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: play")
			if err = s.messageCommandPlayResponse(packet); err != nil {
				return err
			}
			s.done = true
			s.isPublisher = false
		default:
			logger.Critical("Unknown command: %s", vs[0].(string))
		}
	}

	return nil
}

const (
	CommandConnectWellKnownID int = 1
)

func (s *ServerConn) messageCommandConnect(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != CommandConnectWellKnownID {
				// RTMP says that the ID should be 1
				return fmt.Errorf("invalid typeID per RTMP protocol")
			}
			s.transactionID = id
		case amf.Object:
			objmap := v.(amf.Object)
			if app, ok := objmap[ConnInfoKeyApp]; ok {
				s.ConnInfo.App = app.(string)
			}
			if flashVer, ok := objmap[ConnInfoKeyFlashVer]; ok {
				s.ConnInfo.FlashVer = flashVer.(string)
			}
			if tcurl, ok := objmap[ConnInfoKeyTcURL]; ok {
				s.ConnInfo.TcUrl = tcurl.(string)
			}
			if encoding, ok := objmap[ConnInfoObjectEncoding]; ok {
				s.ConnInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
	return nil
}

func (s *ServerConn) messageCommandConnectResponse(packet *ChunkStream) error {
	respPacket := s.conn.NewWindowAckSize(DefaultWindowAcknowledgementSizeBytes)
	s.conn.Write(&respPacket)
	respPacket = s.conn.NewSetPeerBandwidth(DefaultPeerBandwidthSizeBytes)
	s.conn.Write(&respPacket)
	respPacket = s.conn.NewSetChunkSize(DefaultRTMPChunkSizeBytesLarge)
	s.conn.Write(&respPacket)

	resp := make(amf.Object)
	resp[ConnRespFMSVer] = DefaultServerFMSVersion
	resp[ConnRespCapabilities] = 31

	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamConnectSuccess
	event[ConnEventDescription] = "Connection succeeded."
	event[ConnEventObjectEncoding] = s.ConnInfo.ObjectEncoding
	return s.writeMsg(packet.CSID, packet.StreamID, CommandType_Result, s.transactionID, resp, event)
}

func (s *ServerConn) messageCommandCreateStream(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			s.transactionID = int(v.(float64))
		case amf.Object:
		}
	}
	return nil
}

func (s *ServerConn) messageCommandCreateStreamResponse(packet *ChunkStream) error {
	return s.writeMsg(packet.CSID, packet.StreamID, CommandType_Result, s.transactionID, nil, s.streamID)
}

// messageCommandPlayPublish will respond to both play and publish commands
func (s *ServerConn) messageCommandPlayPublish(vs []interface{}) error {
	for k, v := range vs {
		switch v.(type) {
		case string:
			if k == 2 {
				s.PublishInfo.Name = v.(string)
			} else if k == 3 {
				s.PublishInfo.Type = v.(string)
			}
		case float64:
			id := int(v.(float64))
			s.transactionID = id
		case amf.Object:
		}
	}

	return nil
}

func (s *ServerConn) messageCommandPublishResponse(cur *ChunkStream) error {
	s.conn.messageUserControlStreamBegin()
	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPublishStart
	event[ConnEventDescription] = "Start publishing."
	return s.writeMsg(cur.CSID, cur.StreamID, CommandTypeOnStatus, 0, nil, event)
}

func (s *ServerConn) messageCommandPlayResponse(cur *ChunkStream) error {
	s.conn.SetRecorded()
	s.conn.messageUserControlStreamBegin()

	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPlayReset
	event[ConnEventDescription] = "Playing and resetting stream."
	if err := s.writeMsg(cur.CSID, cur.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
		return err
	}

	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPlayStart
	event[ConnEventDescription] = "Started playing stream."
	if err := s.writeMsg(cur.CSID, cur.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
		return err
	}

	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamDataStart
	event[ConnEventDescription] = "Started playing stream."
	if err := s.writeMsg(cur.CSID, cur.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
		return err
	}

	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPublishNotify
	event[ConnEventDescription] = "Started playing notify."
	if err := s.writeMsg(cur.CSID, cur.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
		return err
	}
	return s.conn.Flush()
}

func (s *ServerConn) IsConnected() bool {
	return s.isConnected
}

func (s *ServerConn) IsPublisher() bool {
	return s.isPublisher
}

func (s *ServerConn) Write(packet ChunkStream) error {
	if packet.TypeID == TAG_SCRIPTDATAAMF0 ||
		packet.TypeID == TAG_SCRIPTDATAAMF3 {
		var err error
		if packet.Data, err = amf.MetaDataReform(packet.Data, amf.DEL); err != nil {
			return err
		}
		packet.Length = uint32(len(packet.Data))
	}
	return s.conn.Write(&packet)
}

func (s *ServerConn) Flush() error {
	return s.conn.Flush()
}

func (s *ServerConn) Read(packet *ChunkStream) (err error) {
	return s.conn.Read(packet)
}

func (s *ServerConn) GetInfo() (app string, name string, url string) {
	app = s.ConnInfo.App
	name = s.PublishInfo.Name
	url = s.ConnInfo.TcUrl + "/" + s.PublishInfo.Name
	return
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
