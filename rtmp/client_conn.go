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

	// virtualMetaData can be used to set the metadata for a client connection.
	// This will be sent during Publish()
	virtualMetaData *MetaData

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

// Publish will hang and attempt to start a Publish stream
// with a configured server.
func (cc *ClientConn) Publish() error {
	cc.method = ClientMethodPlay
	logger.Info(rtmpMessage("client.Publish", pub))
	err := cc.initialTX()
	if err != nil {
		return err
	}

	// Read until the server is connected
	for !cc.connected {
		x, err := cc.NextChunk()
		if err != nil {
			return err
		}
		err = cc.Route(x)
		if err != nil {
			return err
		}
	}

	_, err = cc.publishTX()
	if err != nil {
		return err
	}

	err = cc.sendMetaData()
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
		x, err = cc.NextChunk()
		if err != nil {
			if err != TestableEOFError {
				return err
			}
			continue
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
	case WindowAcknowledgementSizeMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.windowAckSize = ackSize
		logger.Debug(rtmpMessage(typeIDString(x), ack))
		cc.conn.ack(ackSize)
	case SetPeerBandwidthMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		ackSize := binary.BigEndian.Uint32(x.Data)
		cc.conn.ack(ackSize)
	case UserControlMessageID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		return cc.handleUserControl(x)
	case CommandMessageAMF0ID, CommandMessageAMF3ID:
		logger.Debug(rtmpMessage(typeIDString(x), rx))
		xReader := bytes.NewReader(x.Data)
		values, err := cc.LogDecodeBatch(xReader, amf.AMF0)
		if err != nil && err != io.EOF {
			return fmt.Errorf("decoding bytes from play(%s) client: %v", cc.urladdr.SafeURL(), err)
		}
		x.batchedValues = values
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
					err := cc.publishRX(x)
					if err != nil {
						return err
					}
					//logger.Info(rtmpMessage("Publish Stream", stream))
					return nil
				case CommandPlay:
					logger.Debug(rtmpMessage(fmt.Sprintf("command.%s", cc.curcmdName), rx))
					err := cc.playRX(x)
					if err != nil {
						return err
					}
					logger.Info(rtmpMessage("Play Stream", stream))
					return nil
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

func (cc *ClientConn) handleUserControl(x *ChunkStream) error {
	logger.Debug(rtmpMessage("handleUserControl", rx))
	if x.Length == UserMessagePingRequest {
		logger.Debug(rtmpMessage("handleUserControl.PingResponse", tx))
		tx := newChunkStream(UserControlMessageID, 6, UserMessagePingResponse)
		return cc.Write(tx)
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
	_, err = cc.oosReleaseStreamTX()
	if err != nil {
		return err
	}
	_, err = cc.oosFCPublishTX()
	if err != nil {
		return err
	}

	return nil
}

func (cc *ClientConn) sendMetaData() error {
	if cc.virtualMetaData == nil {
		return nil
	}
	logger.Debug(rtmpMessage("sendMetaData", tx))
	md := make(amf.Object)
	// //2021-10-13T10:48:38-07:00 [Debug     ]    [2] (map[2.1:false 3.1:false 4.0:false 4.1:false 5.1:false 7.1:false
	// audiochannels:2 audiocodecid:10 audiodatarate:160 audiosamplerate:48000 audiosamplesize:16 duration:0
	// encoder:obs-output module (libobs version 27.0.1-3) fileSize:0 framerate:30 height:720 stereo:true
	// videocodecid:7 videodatarate:2500 width:1280])
	md["2.1"] = cc.virtualMetaData.V21
	md["3.1"] = cc.virtualMetaData.V31
	md["4.0"] = cc.virtualMetaData.V40
	md["4.1"] = cc.virtualMetaData.V41
	md["5.1"] = cc.virtualMetaData.V51
	md["7.1"] = cc.virtualMetaData.V71
	md["audiochannels"] = cc.virtualMetaData.AudioChannels
	md["audiocodecid"] = cc.virtualMetaData.AudioCodecID
	md["audiodatarate"] = cc.virtualMetaData.AudioDataRate
	md["audiosamplerate"] = cc.virtualMetaData.AudioSampleRate
	md["audiosamplesize"] = cc.virtualMetaData.AudioSampleSize
	md["duration"] = cc.virtualMetaData.Duration
	md["encoder"] = cc.virtualMetaData.Encoder
	md["filesize"] = cc.virtualMetaData.FileSize
	md["framerate"] = cc.virtualMetaData.FrameRate
	md["height"] = cc.virtualMetaData.Height
	md["stero"] = cc.virtualMetaData.Stereo
	md["videocodecid"] = cc.virtualMetaData.VideoCodecID
	md["videodatarate"] = cc.virtualMetaData.VideoDataRate
	md["width"] = cc.virtualMetaData.Width
	_, err := cc.writeMsg(DataMessageAMF0ID, "@setDataFrame", "onMetaData", md)
	return err
}

// ==========================================================================================

func (cc *ClientConn) LogDecodeBatch(r io.Reader, ver amf.Version) (ret []interface{}, err error) {
	vs, err := cc.decoder.DecodeBatch(r, ver)
	for k, v := range vs {
		logger.Debug("  [%+v] (%+v)", k, v)
	}
	return vs, err
}

func (cc *ClientConn) Write(c *ChunkStream) error {
	M().Lock()
	P(cc.urladdr.SafeURL()).ProxyTotalPacketsTX++
	P(cc.urladdr.SafeURL()).ProxyTotalBytesTX = P(cc.urladdr.SafeURL()).ProxyTotalBytesTX + int(c.Length)
	M().Unlock()

	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.ADD); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return cc.conn.Write(c)
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
