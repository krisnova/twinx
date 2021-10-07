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
	"math/rand"
	"net"
	neturl "net/url"
	"strings"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/kris-nova/logger"
)

type ConnClient struct {
	conn    *Conn
	urladdr *URLAddr

	done       bool
	transID    int
	url        string
	tcurl      string
	app        string
	title      string
	query      string
	curcmdName string
	streamid   uint32
	encoder    *amf.Encoder
	decoder    *amf.Decoder
	bytesw     *bytes.Buffer
}

func NewConnClient() *ConnClient {
	return &ConnClient{
		transID: 1,
		bytesw:  bytes.NewBuffer(nil),
		encoder: &amf.Encoder{},
		decoder: &amf.Decoder{},
	}
}

func (cc *ConnClient) DecodeBatch(r io.Reader, ver amf.Version) (ret []interface{}, err error) {
	vs, err := cc.decoder.DecodeBatch(r, ver)
	return vs, err
}

func (cc *ConnClient) readRespMsg() error {

	var x ChunkStream
	for {
		if err := cc.conn.Read(&x); err != nil {
			return err
		}

		switch x.TypeID {
		case SetChunkSizeMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case AbortMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case AcknowledgementMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case WindowAcknowledgementSizeMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case SetPeerBandwidthMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case UserControlMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case CommandMessageAMF0ID, CommandMessageAMF3ID:
			xReader := bytes.NewReader(x.Data)
			values, err := cc.decoder.DecodeBatch(xReader, amf.AMF0)
			if err != nil && err != io.EOF {
				return fmt.Errorf("decoding bytes from play(%s) client: %v", cc.urladdr.SafeURL(), err)
			}
			for _, v := range values {
				switch v.(type) {
				case string:
				case float64:
				case amf.Object:
					// Todo unmarshal this into ConnEvent
					entity := v.(amf.Object)
					switch cc.curcmdName {
					case CommandConnect:
						code, ok := entity[ConnEventCode]
						if ok && code.(string) != CommandNetStreamConnectSuccess {
							return fmt.Errorf("unable to connect: error code: %d", code)
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
	}
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
	event["app"] = cc.app
	event["type"] = "nonprivate"
	event["flashVer"] = "FMS.3.1"
	event["tcUrl"] = cc.tcurl
	cc.curcmdName = CommandConnect

	//logger.Info("writeConnectMsg: cc.transID=%d, event=%v", cc.transID, event)
	if err := cc.writeMsg(CommandConnect, cc.transID, event); err != nil {
		return err
	}
	return cc.readRespMsg()
}

func (cc *ConnClient) writeCreateStreamMsg() error {
	cc.transID++
	cc.curcmdName = CommandCreateStream

	//logger.Info("writeCreateStreamMsg: cc.transID=%d", cc.transID)
	if err := cc.writeMsg(CommandCreateStream, cc.transID, nil); err != nil {
		return err
	}

	err := cc.readRespMsg()
	if err == nil {
		return nil
	}

	//logger.Info("writeCreateStreamMsg readRespMsg err=%v", err)
	return err

}

func (cc *ConnClient) writePublishMsg() error {
	cc.transID++
	cc.curcmdName = CommandPublish
	if err := cc.writeMsg(CommandPublish, cc.transID, nil, cc.title, PublishCommandLive); err != nil {
		return err
	}
	return cc.readRespMsg()
}

func (cc *ConnClient) writePlayMsg() error {
	cc.transID++
	cc.curcmdName = CommandPlay

	if err := cc.writeMsg(CommandPlay, 0, nil, cc.title); err != nil {
		return err
	}
	return cc.readRespMsg()
}

func (cc *ConnClient) Start(url string, method string) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	cc.url = url
	path := strings.TrimLeft(u.Path, "/")
	ps := strings.SplitN(path, "/", 2)
	if len(ps) != 2 {
		return fmt.Errorf("u path err: %s", path)
	}
	cc.app = ps[0]
	cc.title = ps[1]
	cc.query = u.RawQuery
	cc.tcurl = "rtmp://" + u.Host + "/" + cc.app
	port := ":1935"
	host := u.Host
	localIP := ":0"
	var remoteIP string
	if strings.Index(host, ":") != -1 {
		host, port, err = net.SplitHostPort(host)
		if err != nil {
			return err
		}
		port = ":" + port
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		//logger.Warning("look up host IP: %v", err)
		return err
	}
	remoteIP = ips[rand.Intn(len(ips))].String()
	if strings.Index(remoteIP, ":") == -1 {
		remoteIP += port
	}

	local, err := net.ResolveTCPAddr("tcp", localIP)
	if err != nil {
		//logger.Warning("Proxy (local) resolve TCP addr: %v", err)
		return err
	}
	remote, err := net.ResolveTCPAddr("tcp", remoteIP)
	if err != nil {
		//logger.Warning("Proxy (remote) resolve TCP addr: %v", err)
		return err
	}
	conn, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		//logger.Critical("Bridging proxy connection from local -> remote %v", err)
		return err
	}

	////logger.Info("Connection")
	//logger.Info("connection:", "local:", conn.LocalAddr(), "remote:", conn.RemoteAddr())

	cc.conn = NewConn(conn, 4*1024)

	if err := cc.conn.HandshakeClient(); err != nil {
		//logger.Warning("[RTMP] Handshake", err)
		return err
	}
	//logger.Debug("[RTMP] Handshake")

	if err := cc.writeConnectMsg(); err != nil {
		//logger.Warning("[RTMP] Connecting", err)
		return err
	}
	//logger.Debug("[RTMP] Connecting")

	if err := cc.writeCreateStreamMsg(); err != nil {
		//logger.Warning("[RTMP] Creating Stream", err)
		return err
	}
	//logger.Debug("[RTMP] Creating Stream")

	//logger.Info("Method control: %s %s %s", method, av.PUBLISH, av.PLAY)
	if method == av.PUBLISH {
		if err := cc.writePublishMsg(); err != nil {
			return err
		}
	} else if method == av.PLAY {
		if err := cc.writePlayMsg(); err != nil {
			return err
		}
	}

	return nil
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

func (cc *ConnClient) GetInfo() (app string, name string, url string) {
	app = cc.app
	name = cc.title
	url = cc.url
	return
}

func (cc *ConnClient) GetStreamId() uint32 {
	return cc.streamid
}

func (cc *ConnClient) Close() {
	cc.conn.Close()
}
