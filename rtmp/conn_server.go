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

type ConnServer struct {
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

func NewConnServer(conn *Conn) *ConnServer {
	return &ConnServer{
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
func (c *ConnServer) ReadPacket() (*ChunkStream, error) {
	var chunk ChunkStream
	if err := c.conn.Read(&chunk); err != nil {
		return nil, fmt.Errorf("reading chunk from client: %v", err)
	}
	return &chunk, nil
}

func (c *ConnServer) messageCommand(packet *ChunkStream) error {
	amfType := amf.AMF0
	if packet.TypeID == CommandMessageAMF3ID {
		// Arithmetic to match AMF3 encoding
		amfType = amf.AMF3
		packet.Data = packet.Data[1:]
	}
	r := bytes.NewReader(packet.Data)
	vs, err := c.decoder.DecodeBatch(r, amf.Version(amfType))
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
			if err = c.messageCommandConnect(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: connect")
			if err = c.messageCommandConnectResponse(packet); err != nil {
				return err
			}
			c.isConnected = true
		case CommandCreateStream:
			logger.Info("Command: %s", CommandCreateStream)
			if err = c.messageCommandCreateStream(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: createStream")
			if err = c.messageCommandCreateStreamResponse(packet); err != nil {
				return err
			}
			c.isConnected = true
		case CommandPublish:
			logger.Info("Command: %s", CommandPublish)
			if err = c.messageCommandPlayPublish(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: publish")
			if err = c.messageCommandPublishResponse(packet); err != nil {
				return err
			}
			c.isConnected = true
			c.isPublisher = true
		case CommandPlay:
			logger.Info("Command: %s", CommandPlay)
			if err = c.messageCommandPlayPublish(vs[1:]); err != nil {
				return err
			}
			logger.Info("   Response: play")
			if err = c.messageCommandPlayResponse(packet); err != nil {
				return err
			}
			c.done = true
			c.isPublisher = false
		default:
			logger.Critical("Unknown command: %s", vs[0].(string))
		}
	}

	return nil
}

const (
	CommandConnectWellKnownID int = 1
)

func (c *ConnServer) messageCommandConnect(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != CommandConnectWellKnownID {
				// RTMP says that the ID should be 1
				return fmt.Errorf("invalid typeID per RTMP protocol")
			}
			c.transactionID = id
		case amf.Object:
			objmap := v.(amf.Object)
			if app, ok := objmap[ConnInfoKeyApp]; ok {
				c.ConnInfo.App = app.(string)
			}
			if flashVer, ok := objmap[ConnInfoKeyFlashVer]; ok {
				c.ConnInfo.FlashVer = flashVer.(string)
			}
			if tcurl, ok := objmap[ConnInfoKeyTcURL]; ok {
				c.ConnInfo.TcUrl = tcurl.(string)
			}
			if encoding, ok := objmap[ConnInfoObjectEncoding]; ok {
				c.ConnInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
	return nil
}

func (c *ConnServer) messageCommandConnectResponse(packet *ChunkStream) error {
	respPacket := c.conn.NewWindowAckSize(DefaultWindowAcknowledgementSizeBytes)
	c.conn.Write(&respPacket)
	respPacket = c.conn.NewSetPeerBandwidth(DefaultPeerBandwidthSizeBytes)
	c.conn.Write(&respPacket)
	respPacket = c.conn.NewSetChunkSize(DefaultRTMPChunkSizeBytesLarge)
	c.conn.Write(&respPacket)

	resp := make(amf.Object)
	resp[ConnRespFMSVer] = DefaultServerFMSVersion
	resp[ConnRespCapabilities] = 31

	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamConnectSuccess
	event[ConnEventDescription] = "Connection succeeded."
	event[ConnEventObjectEncoding] = c.ConnInfo.ObjectEncoding
	return c.writeMsg(packet.CSID, packet.StreamID, CommandType_Result, c.transactionID, resp, event)
}

func (c *ConnServer) messageCommandCreateStream(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			c.transactionID = int(v.(float64))
		case amf.Object:
		}
	}
	return nil
}

func (c *ConnServer) messageCommandCreateStreamResponse(packet *ChunkStream) error {
	return c.writeMsg(packet.CSID, packet.StreamID, CommandType_Result, c.transactionID, nil, c.streamID)
}

// messageCommandPlayPublish will respond to both play and publish commands
func (c *ConnServer) messageCommandPlayPublish(vs []interface{}) error {
	for k, v := range vs {
		switch v.(type) {
		case string:
			if k == 2 {
				c.PublishInfo.Name = v.(string)
			} else if k == 3 {
				c.PublishInfo.Type = v.(string)
			}
		case float64:
			id := int(v.(float64))
			c.transactionID = id
		case amf.Object:
		}
	}

	return nil
}

func (c *ConnServer) messageCommandPublishResponse(cur *ChunkStream) error {
	c.conn.messageUserControlStreamBegin()
	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPublishStart
	event[ConnEventDescription] = "Start publishing."
	return c.writeMsg(cur.CSID, cur.StreamID, CommandType_OnStatus, 0, nil, event)
}

func (c *ConnServer) messageCommandPlayResponse(cur *ChunkStream) error {
	c.conn.SetRecorded()
	c.conn.messageUserControlStreamBegin()

	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPlayReset
	event[ConnEventDescription] = "Playing and resetting stream."
	if err := c.writeMsg(cur.CSID, cur.StreamID, CommandType_OnStatus, 0, nil, event); err != nil {
		return err
	}

	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPlayStart
	event[ConnEventDescription] = "Started playing stream."
	if err := c.writeMsg(cur.CSID, cur.StreamID, CommandType_OnStatus, 0, nil, event); err != nil {
		return err
	}

	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamDataStart
	event[ConnEventDescription] = "Started playing stream."
	if err := c.writeMsg(cur.CSID, cur.StreamID, CommandType_OnStatus, 0, nil, event); err != nil {
		return err
	}

	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPublishNotify
	event[ConnEventDescription] = "Started playing notify."
	if err := c.writeMsg(cur.CSID, cur.StreamID, CommandType_OnStatus, 0, nil, event); err != nil {
		return err
	}
	return c.conn.Flush()
}

func (c *ConnServer) IsConnected() bool {
	return c.isConnected
}

func (c *ConnServer) IsPublisher() bool {
	return c.isPublisher
}

func (c *ConnServer) Write(packet ChunkStream) error {
	if packet.TypeID == TAG_SCRIPTDATAAMF0 ||
		packet.TypeID == TAG_SCRIPTDATAAMF3 {
		var err error
		if packet.Data, err = amf.MetaDataReform(packet.Data, amf.DEL); err != nil {
			return err
		}
		packet.Length = uint32(len(packet.Data))
	}
	return c.conn.Write(&packet)
}

func (c *ConnServer) Flush() error {
	return c.conn.Flush()
}

func (c *ConnServer) Read(packet *ChunkStream) (err error) {
	return c.conn.Read(packet)
}

func (c *ConnServer) GetInfo() (app string, name string, url string) {
	app = c.ConnInfo.App
	name = c.PublishInfo.Name
	url = c.ConnInfo.TcUrl + "/" + c.PublishInfo.Name
	return
}

func (c *ConnServer) Close(err error) {
	c.conn.Close()
}

func (c *ConnServer) writeMsg(csid, streamID uint32, args ...interface{}) error {
	c.bytesw.Reset()
	for _, v := range args {
		if _, err := c.encoder.Encode(c.bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := c.bytesw.Bytes()
	packet := ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  streamID,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	c.conn.Write(&packet)
	return c.conn.Flush()
}
