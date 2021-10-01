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
	done       bool
	transID    int
	url        string
	tcurl      string
	app        string
	title      string
	query      string
	curcmdName string
	streamid   uint32
	conn       *Conn
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

func (connClient *ConnClient) DecodeBatch(r io.Reader, ver amf.Version) (ret []interface{}, err error) {
	vs, err := connClient.decoder.DecodeBatch(r, ver)
	return vs, err
}

// readRespMsg will parse a message from a server
func (connClient *ConnClient) readRespMsg() error {
	var err error
	var rc ChunkStream
	for {
		if err = connClient.conn.Read(&rc); err != nil {
			return fmt.Errorf("error reading message from server: %v", err)
		}
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading message from server [non-EOF]: %v", err)
		}
		switch rc.TypeID {

		//Command Message (20, 17)
		// Command messages carry the AMF-encoded commands between the client
		// and the server. These messages have been assigned message type value
		// of 20 for AMF0 encoding and message type value of 17 for AMF3
		// encoding. These messages are sent to perform some operations like
		// connect, createStream, publish, play, pause on the peer. Command
		// messages like onstatus, result etc. are used to inform the sender
		// about the status of the requested commands. A command message
		// consists of command name, transaction ID, and command object that
		// contains related parameters. A client or a server can request Remote
		// Procedure Calls (RPC) over streams that are communicated using the
		// command messages to the peer.
		case 20, 17:
			r := bytes.NewReader(rc.Data)
			rspMsgMap, err := connClient.decoder.DecodeBatch(r, amf.AMF0)
			if err != nil && err != io.EOF {
				logger.Warning("Decoding batch: %v", err)
			}

			logger.Debug("Reading raw message from server: %v", rspMsgMap)

			for k, v := range rspMsgMap {
				logger.Debug("respMap %v: %v", k, v)

				switch v.(type) {
				case string:
					logger.Warning("Unimplemented type in readRespMsg")
				case float64:

				case amf.Object:
					objmap := v.(amf.Object)
					switch connClient.curcmdName {
					case CommandConnect:
						code, ok := objmap["code"]
						if ok && code.(string) != CommandNetStreamConnectSuccess {
							return fmt.Errorf("unable to connect: error code: %d", code)
						}
					case CommandPublish:
						code, ok := objmap["code"]
						if ok && code.(string) != CommandNetStreamPublishStart {
							return fmt.Errorf("unable to publish: error code: %d", code)
						}
					}
				}
			}
			break
		//Data Message (18, 15)
		// The client or the server sends this message to send Metadata or any
		// user data to the peer. Metadata includes details about the
		// data(audio, video etc.) like creation time, duration, theme and so
		// on. These messages have been assigned message type value of 18 for
		// AMF0 and message type value of 15 for AMF3.
		case 18, 15:
			break
		//7.1.3. Shared Object Message (19, 16)
		case 19, 16:
			break
		//7.1.4. Audio Message (8)
		case 8:
			break
		//7.1.5. Video Message (9)
		case 9:
			break
		}
	}
}

func (connClient *ConnClient) writeMsg(args ...interface{}) error {
	connClient.bytesw.Reset()
	for _, v := range args {
		if _, err := connClient.encoder.Encode(connClient.bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := connClient.bytesw.Bytes()
	c := ChunkStream{
		Format:    0,
		CSID:      3,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  connClient.streamid,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	connClient.conn.Write(&c)
	return connClient.conn.Flush()
}

func (connClient *ConnClient) writeConnectMsg() error {
	event := make(amf.Object)
	event["app"] = connClient.app
	event["type"] = "nonprivate"
	event["flashVer"] = "FMS.3.1"
	event["tcUrl"] = connClient.tcurl
	connClient.curcmdName = CommandConnect

	logger.Info("writeConnectMsg: connClient.transID=%d, event=%v", connClient.transID, event)
	if err := connClient.writeMsg(CommandConnect, connClient.transID, event); err != nil {
		return err
	}
	return connClient.readRespMsg()
}

func (connClient *ConnClient) writeCreateStreamMsg() error {
	connClient.transID++
	connClient.curcmdName = CommandCreateStream

	logger.Info("writeCreateStreamMsg: connClient.transID=%d", connClient.transID)
	if err := connClient.writeMsg(CommandCreateStream, connClient.transID, nil); err != nil {
		return err
	}

	err := connClient.readRespMsg()
	if err == nil {
		return nil
	}

	logger.Info("writeCreateStreamMsg readRespMsg err=%v", err)
	return err

}

func (connClient *ConnClient) writePublishMsg() error {
	connClient.transID++
	connClient.curcmdName = CommandPublish
	if err := connClient.writeMsg(CommandPublish, connClient.transID, nil, connClient.title, PublishCommandLive); err != nil {
		return err
	}
	return connClient.readRespMsg()
}

func (connClient *ConnClient) writePlayMsg() error {
	connClient.transID++
	connClient.curcmdName = CommandPlay
	logger.Info("writePlayMsg: connClient.transID=%d, CommandPlay=%v, connClient.title=%v",
		connClient.transID, CommandPlay, connClient.title)

	if err := connClient.writeMsg(CommandPlay, 0, nil, connClient.title); err != nil {
		return err
	}
	return connClient.readRespMsg()
}

func (connClient *ConnClient) Start(url string, method string) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	connClient.url = url
	path := strings.TrimLeft(u.Path, "/")
	ps := strings.SplitN(path, "/", 2)
	if len(ps) != 2 {
		return fmt.Errorf("u path err: %s", path)
	}
	connClient.app = ps[0]
	connClient.title = ps[1]
	connClient.query = u.RawQuery
	connClient.tcurl = "rtmp://" + u.Host + "/" + connClient.app
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
		logger.Warning("look up host IP: %v", err)
		return err
	}
	remoteIP = ips[rand.Intn(len(ips))].String()
	if strings.Index(remoteIP, ":") == -1 {
		remoteIP += port
	}

	local, err := net.ResolveTCPAddr("tcp", localIP)
	if err != nil {
		logger.Warning("Proxy (local) resolve TCP addr: %v", err)
		return err
	}
	remote, err := net.ResolveTCPAddr("tcp", remoteIP)
	if err != nil {
		logger.Warning("Proxy (remote) resolve TCP addr: %v", err)
		return err
	}
	conn, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		logger.Critical("Bridging proxy connection from local -> remote %v", err)
		return err
	}

	//logger.Info("Connection")
	logger.Info("connection:", "local:", conn.LocalAddr(), "remote:", conn.RemoteAddr())

	connClient.conn = NewConn(conn, 4*1024)

	if err := connClient.conn.HandshakeClient(); err != nil {
		logger.Warning("[RTMP] Handshake", err)
		return err
	}
	logger.Debug("[RTMP] Handshake")

	if err := connClient.writeConnectMsg(); err != nil {
		logger.Warning("[RTMP] Connecting", err)
		return err
	}
	logger.Debug("[RTMP] Connecting")

	if err := connClient.writeCreateStreamMsg(); err != nil {
		logger.Warning("[RTMP] Creating Stream", err)
		return err
	}
	logger.Debug("[RTMP] Creating Stream")

	logger.Info("Method control: %s %s %s", method, av.PUBLISH, av.PLAY)
	if method == av.PUBLISH {
		if err := connClient.writePublishMsg(); err != nil {
			return err
		}
	} else if method == av.PLAY {
		if err := connClient.writePlayMsg(); err != nil {
			return err
		}
	}

	return nil
}

func (connClient *ConnClient) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.ADD); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return connClient.conn.Write(&c)
}

func (connClient *ConnClient) Flush() error {
	return connClient.conn.Flush()
}

func (connClient *ConnClient) Read(c *ChunkStream) (err error) {
	return connClient.conn.Read(c)
}

func (connClient *ConnClient) GetInfo() (app string, name string, url string) {
	app = connClient.app
	name = connClient.title
	url = connClient.url
	return
}

func (connClient *ConnClient) GetStreamId() uint32 {
	return connClient.streamid
}

func (connClient *ConnClient) Close(err error) {
	connClient.conn.Close()
}
