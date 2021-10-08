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

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/kris-nova/logger"
)

type ConnClient struct {
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

func NewConnClient() *ConnClient {
	return &ConnClient{
		transID: 1,
		bytesw:  bytes.NewBuffer(nil),
		encoder: &amf.Encoder{},
		decoder: &amf.Decoder{},
	}
}

func (cc *ConnClient) Dial(address string) error {
	urlAddr, err := NewURLAddr(address)
	if err != nil {
		return fmt.Errorf("client dial: %v", err)
	}
	logger.Info("ConnClient.Dial %s", urlAddr.String())
	conn, err := urlAddr.NewConn()
	if err != nil {
		return fmt.Errorf("new conn from addr: %v", err)
	}
	cc.conn = conn
	cc.urladdr = urlAddr
	return nil
}

func (cc *ConnClient) Publish() error {
	cc.method = ClientMethodPublish
	logger.Info("Client: Publish")
	err := cc.connect()
	if err != nil {
		return err
	}
	if err := cc.writePublishMsg(); err != nil {
		return err
	}
	for {
		cc.readRespMsg()
	}
	return nil
}

func (cc *ConnClient) Play() error {
	cc.method = ClientMethodPlay
	logger.Info("Client: Play")
	err := cc.connect()
	if err != nil {
		return err
	}
	if err := cc.writePlayMsg(); err != nil {
		return err
	}
	for {
		cc.readRespMsg()
	}
	return nil
}

func (cc *ConnClient) connect() error {
	if cc.connected {
		return errors.New("already connected")
	}
	if err := cc.conn.HandshakeClient(); err != nil {
		return err
	}
	if err := cc.writeConnectMsg(); err != nil {
		return err
	}
	if err := cc.writeCreateStreamMsg(); err != nil {
		return err
	}
	cc.connected = true
	return nil
}

func (cc *ConnClient) DecodeBatch(r io.Reader, ver amf.Version) (ret []interface{}, err error) {
	vs, err := cc.decoder.DecodeBatch(r, ver)
	return vs, err
}

func (cc *ConnClient) readRespMsg() error {

	var x ChunkStream

	if err := cc.conn.Read(&x); err != nil {
		return err
	}

	switch x.TypeID {
	case SetChunkSizeMessageID:
		logger.Debug(typeIDString(&x))
		chunkSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.chunkSize = chunkSize
		cc.conn.ack(x.Length)
	case AbortMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(&x))
	case AcknowledgementMessageID:
		logger.Debug(typeIDString(&x))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.remoteWindowAckSize = ackSize
		cc.conn.ack(ackSize)
	case WindowAcknowledgementSizeMessageID:
		logger.Debug(typeIDString(&x))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.remoteWindowAckSize = ackSize
		cc.conn.ack(ackSize)
	case SetPeerBandwidthMessageID:
		logger.Debug(typeIDString(&x))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.ack(ackSize)
	case UserControlMessageID:
		logger.Debug(typeIDString(&x))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.ack(ackSize)
	case CommandMessageAMF0ID, CommandMessageAMF3ID:
		logger.Debug(typeIDString(&x))
		xReader := bytes.NewReader(x.Data)
		values, err := cc.decoder.DecodeBatch(xReader, amf.AMF0)
		if err != nil && err != io.EOF {
			return fmt.Errorf("decoding bytes from play(%s) client: %v", cc.urladdr.SafeURL(), err)
		}
		for k, v := range values {
			switch v.(type) {
			case string:
				switch cc.curcmdName {
				case CommandConnect, CommandCreateStream:
				case CommandPublish:
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
						return fmt.Errorf("unable to connect: error code: %v", code)
					}
				case CommandPublish:
					code, ok := entity[ConnEventCode]
					if ok && code.(string) != CommandNetStreamPublishStart {
						return fmt.Errorf("unable to publish: error code: %d", code)
					}
				}
			}
		}

	case DataMessageAMF0ID, DataMessageAMF3ID:
		logger.Critical("unsupported messageID: %s", typeIDString(&x))
	case SharedObjectMessageAMF0ID, SharedObjectMessageAMF3ID:
		logger.Critical("unsupported messageID: %s", typeIDString(&x))
	case AudioMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(&x))
	case VideoMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(&x))
	case AggregateMessageID:
		logger.Critical("unsupported messageID: %s", typeIDString(&x))
	default:
		logger.Critical("unsupported messageID: %s", typeIDString(&x))
	}
	return nil
}

func (cc *ConnClient) writeMsg(args ...interface{}) error {
	cc.bytesw.Reset()
	for _, v := range args {
		if _, err := cc.encoder.Encode(cc.bytesw, v, amf.AMF0); err != nil {
			return err
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
	return cc.conn.Flush()
}

func (cc *ConnClient) writeConnectMsg() error {
	event := make(amf.Object)
	event[ConnInfoKeyApp] = cc.urladdr.App()
	event[ConnInfoKeyType] = "nonprivate"
	event[ConnInfoKeyFlashVer] = DefaultServerFMSVersion
	event[ConnInfoKeyTcURL] = cc.urladdr.StreamURL()
	cc.curcmdName = CommandConnect

	if err := cc.writeMsg(CommandConnect, cc.transID, event); err != nil {
		return err
	}
	return cc.readRespMsg()
}

func (cc *ConnClient) writeCreateStreamMsg() error {
	cc.transID++
	cc.curcmdName = CommandCreateStream

	if err := cc.writeMsg(CommandCreateStream, cc.transID, nil); err != nil {
		return err
	}

	err := cc.readRespMsg()
	if err != nil {
		return err
	}
	return nil

}

func (cc *ConnClient) writePublishMsg() error {
	cc.transID++
	cc.curcmdName = CommandPublish
	if err := cc.writeMsg(CommandPublish, cc.transID, nil, cc.urladdr.Key(), PublishCommandLive); err != nil {
		return err
	}
	return cc.readRespMsg()
}

func (cc *ConnClient) writePlayMsg() error {
	cc.transID++
	cc.curcmdName = CommandPlay

	if err := cc.writeMsg(CommandPlay, 0, nil, cc.urladdr.Key()); err != nil {
		return err
	}
	return cc.readRespMsg()
}

func (cc *ConnClient) Write(c ChunkStream) error {
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

func (cc *ConnClient) Flush() error {
	return cc.conn.Flush()
}

func (cc *ConnClient) Read(c *ChunkStream) (err error) {
	return cc.conn.Read(c)
}

func (cc *ConnClient) GetStreamId() uint32 {
	return cc.streamid
}

func (cc *ConnClient) Close() {
	cc.conn.Close()
}
