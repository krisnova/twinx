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
	"fmt"
	"io"

	"github.com/gwuhaolin/livego/av"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/kris-nova/logger"
)

type ClientConn struct {
	conn       *Conn
	urladdr    *URLAddr
	method     ClientMethod
	connected  bool
	transID    int
	curcmdName string
	streamid   uint32

	encoder *amf.Encoder
	decoder *amf.Decoder
	bytesw  *bytes.Buffer
}

func NewClientConn() *ClientConn {
	return &ClientConn{
		transID: 1,
		bytesw:  bytes.NewBuffer(nil),
		encoder: &amf.Encoder{},
		decoder: &amf.Decoder{},
	}
}

func (cc *ClientConn) Dial(address string) error {
	urlAddr, err := NewURLAddr(address)
	if err != nil {
		return fmt.Errorf("client dial: %v", err)
	}
	logger.Info(rtmpMessage(fmt.Sprintf("client.Dial %s", urlAddr.Host()), conn))
	conn, err := urlAddr.NewConn()
	if err != nil {
		return fmt.Errorf("new conn from addr: %v", err)
	}
	cc.conn = conn
	cc.urladdr = urlAddr
	return nil
}

// Publish will attempt to start a Publish stream
// with a configured server.
func (cc *ClientConn) Publish() error {
	cc.method = ClientMethodPlay
	logger.Info(rtmpMessage("client.Publish", pub))
	err := cc.initialTX()
	if err != nil {
		return err
	}
	_, err = cc.publishTX()
	if err != nil {
		return err
	}

	err = cc.RoutePackets()
	if err != nil {
		logger.Critical(err.Error())
	}

	return nil
}

// Play will attempt to start a Play stream
// with a configured server.
func (cc *ClientConn) Play() error {
	cc.method = ClientMethodPlay
	logger.Info(rtmpMessage("Client Publish", play))
	err := cc.initialTX()
	if err != nil {
		return err
	}
	_, err = cc.playTX()
	if err != nil {
		return err
	}

	err = cc.RoutePackets()
	if err != nil {
		logger.Critical(err.Error())
	}

	return nil
}

func (cc *ClientConn) RoutePackets() error {
	var x *ChunkStream
	var err error
	for {
		if cc.connected {
			// **************************************
			// Hér vera drekar
			// **************************************
			//
			// We should be routing audio video packets
			// through here as well.
			//
			// For now they are handled elsewhere.
			break
		}
		x, err = cc.NextChunk()
		if err != nil {
			return err
		}
		err = cc.Route(x)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cc *ClientConn) Route(x *ChunkStream) error {
	switch x.TypeID {
	case SetChunkSizeMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		chunkSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.chunkSize = chunkSize
		cc.conn.ack(x.Length)
		logger.Debug(rtmpMessage(typeIDString(x), ack))
	case AbortMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(x))
	case AcknowledgementMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.windowAckSize = ackSize
		logger.Debug(rtmpMessage(typeIDString(x), ack))
		cc.conn.ack(ackSize)
		logger.Debug(rtmpMessage(typeIDString(x), tx))
	case WindowAcknowledgementSizeMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.windowAckSize = ackSize
		logger.Debug(rtmpMessage(typeIDString(x), ack))
		cc.conn.ack(ackSize)
		logger.Debug(rtmpMessage(typeIDString(x), tx))
	case SetPeerBandwidthMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.ack(ackSize)
		logger.Debug(rtmpMessage(typeIDString(x), tx))
	case UserControlMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.ack(ackSize)
		logger.Debug(rtmpMessage(typeIDString(x), tx))
	case CommandMessageAMF0ID, CommandMessageAMF3ID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		xReader := bytes.NewReader(x.Data)
		values, err := cc.decoder.DecodeBatch(xReader, amf.AMF0)
		if err != nil && err != io.EOF {
			return fmt.Errorf("decoding bytes from play(%s) client: %v", cc.urladdr.SafeURL(), err)
		}
		for k, v := range values {
			switch v.(type) {
			case string:
				logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", cc.curcmdName), ack))
				switch cc.curcmdName {
				case CommandConnect:
					logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", cc.curcmdName), rx))
					return cc.connectRX(x)
				case CommandCreateStream:
					logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", cc.curcmdName), rx))
					return cc.createStreamRX(x)
				case CommandPublish:
					logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", cc.curcmdName), rx))
					return cc.publishRX(x)
				case CommandPlay:
					logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", cc.curcmdName), rx))
					return cc.playRX(x)
				}
			case float64:
				switch cc.curcmdName {
				case CommandConnect, CommandCreateStream:
					id := int(v.(float64))
					if k == 1 {
						if id != cc.transID {
							return fmt.Errorf("invalid ID")
						}
					} else if k == 3 {
						logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", cc.curcmdName), ack))
						cc.streamid = uint32(id)
					}
				case CommandPublish:
					if int(v.(float64)) != 0 {
						return fmt.Errorf("invalid publish")
					}
				}
			case amf.Object:
				// Todo unmarshal this into ConnEvent
				entity := v.(amf.Object)
				switch cc.curcmdName {
				case CommandConnect:
					code, ok := entity[ConnEventCode]
					if ok && code.(string) != CommandNetStreamConnectSuccess {
						return fmt.Errorf("connect error : %v", code)
					}
				case CommandPublish:
					code, ok := entity[ConnEventCode]
					if ok && code.(string) != CommandNetStreamPublishStart {
						return fmt.Errorf("publish error: %d", code)
					}
				}
			}
		}
		return nil
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
	return nil
}

func (cc *ClientConn) initialTX() error {
	var err error
	err = cc.handshake()
	if err != nil {
		return err
	}
	_, err = cc.connectTX()
	if err != nil {
		return err
	}
	_, err = cc.createStreamTX()
	if err != nil {
		return err
	}
	return nil
}

// ==========================================================================================

func (cc *ClientConn) DecodeBatch(r io.Reader, ver amf.Version) (ret []interface{}, err error) {
	vs, err := cc.decoder.DecodeBatch(r, ver)
	return vs, err
}

func (cc *ClientConn) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.ADD); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return cc.conn.Write(&c)
}

func (cc *ClientConn) Flush() error {
	return cc.conn.Flush()
}

func (cc *ClientConn) NextChunk() (*ChunkStream, error) {
	x := ChunkStream{}
	err := cc.conn.Read(&x)
	return &x, err
}

func (cc *ClientConn) Read(c *ChunkStream) (err error) {
	return cc.conn.Read(c)
}

func (cc *ClientConn) GetStreamId() uint32 {
	return cc.streamid
}

func (cc *ClientConn) Close() {
	cc.conn.Close()
}

func (cc *ClientConn) writeMsg(args ...interface{}) (*ChunkStream, error) {
	cc.bytesw.Reset()
	for _, v := range args {
		if _, err := cc.encoder.Encode(cc.bytesw, v, amf.AMF0); err != nil {
			return nil, err
		}
	}
	msg := cc.bytesw.Bytes()
	c := ChunkStream{
		Format:    0,
		CSID:      3,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  cc.streamid,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	cc.conn.Write(&c)
	return &c, cc.conn.Flush()
}

// ==========================================================================================
