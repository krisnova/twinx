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
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/kris-nova/logger"

	neturl "net/url"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/utils/pio"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/utils/pool"
)

type ChunkStream struct {
	Format    uint32
	CSID      uint32
	Timestamp uint32
	Length    uint32
	TypeID    uint32
	StreamID  uint32
	timeDelta uint32
	exted     bool
	index     uint32
	remain    uint32
	got       bool
	tmpFromat uint32
	Data      []byte
}

func (chunkStream *ChunkStream) full() bool {
	return chunkStream.got
}

func (chunkStream *ChunkStream) new(pool *pool.Pool) {
	chunkStream.got = false
	chunkStream.index = 0
	chunkStream.remain = chunkStream.Length
	chunkStream.Data = pool.Get(int(chunkStream.Length))
}

func (chunkStream *ChunkStream) writeHeader(w *ReadWriter) error {
	//Chunk Basic Header
	h := chunkStream.Format << 6
	switch {
	case chunkStream.CSID < 64:
		h |= chunkStream.CSID
		w.WriteUintBE(h, 1)
	case chunkStream.CSID-64 < 256:
		h |= 0
		w.WriteUintBE(h, 1)
		w.WriteUintLE(chunkStream.CSID-64, 1)
	case chunkStream.CSID-64 < 65536:
		h |= 1
		w.WriteUintBE(h, 1)
		w.WriteUintLE(chunkStream.CSID-64, 2)
	}
	//Chunk Message Header
	ts := chunkStream.Timestamp
	if chunkStream.Format == 3 {
		goto END
	}
	if chunkStream.Timestamp > 0xffffff {
		ts = 0xffffff
	}
	w.WriteUintBE(ts, 3)
	if chunkStream.Format == 2 {
		goto END
	}
	if chunkStream.Length > 0xffffff {
		return fmt.Errorf("length=%d", chunkStream.Length)
	}
	w.WriteUintBE(chunkStream.Length, 3)
	w.WriteUintBE(chunkStream.TypeID, 1)
	if chunkStream.Format == 1 {
		goto END
	}
	w.WriteUintLE(chunkStream.StreamID, 4)
END:
	//Extended Timestamp
	if ts >= 0xffffff {
		w.WriteUintBE(chunkStream.Timestamp, 4)
	}
	return w.WriteError()
}

func (chunkStream *ChunkStream) writeChunk(w *ReadWriter, chunkSize int) error {
	if chunkStream.TypeID == av.TAG_AUDIO {
		chunkStream.CSID = 4
	} else if chunkStream.TypeID == av.TAG_VIDEO ||
		chunkStream.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		chunkStream.TypeID == av.TAG_SCRIPTDATAAMF3 {
		chunkStream.CSID = 6
	}

	totalLen := uint32(0)
	numChunks := (chunkStream.Length / uint32(chunkSize))
	for i := uint32(0); i <= numChunks; i++ {
		if totalLen == chunkStream.Length {
			break
		}
		if i == 0 {
			chunkStream.Format = uint32(0)
		} else {
			chunkStream.Format = uint32(3)
		}
		if err := chunkStream.writeHeader(w); err != nil {
			return err
		}
		inc := uint32(chunkSize)
		start := uint32(i) * uint32(chunkSize)
		if uint32(len(chunkStream.Data))-start <= inc {
			inc = uint32(len(chunkStream.Data)) - start
		}
		totalLen += inc
		end := start + inc
		buf := chunkStream.Data[start:end]
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}

	return nil

}

func (chunkStream *ChunkStream) readChunk(r *ReadWriter, chunkSize uint32, pool *pool.Pool) error {
	if chunkStream.remain != 0 && chunkStream.tmpFromat != 3 {
		return fmt.Errorf("invalid remain = %d", chunkStream.remain)
	}
	switch chunkStream.CSID {
	case 0:
		id, _ := r.ReadUintLE(1)
		chunkStream.CSID = id + 64
	case 1:
		id, _ := r.ReadUintLE(2)
		chunkStream.CSID = id + 64
	}

	switch chunkStream.tmpFromat {
	case 0:
		chunkStream.Format = chunkStream.tmpFromat
		chunkStream.Timestamp, _ = r.ReadUintBE(3)
		chunkStream.Length, _ = r.ReadUintBE(3)
		chunkStream.TypeID, _ = r.ReadUintBE(1)
		chunkStream.StreamID, _ = r.ReadUintLE(4)
		if chunkStream.Timestamp == 0xffffff {
			chunkStream.Timestamp, _ = r.ReadUintBE(4)
			chunkStream.exted = true
		} else {
			chunkStream.exted = false
		}
		chunkStream.new(pool)
	case 1:
		chunkStream.Format = chunkStream.tmpFromat
		timeStamp, _ := r.ReadUintBE(3)
		chunkStream.Length, _ = r.ReadUintBE(3)
		chunkStream.TypeID, _ = r.ReadUintBE(1)
		if timeStamp == 0xffffff {
			timeStamp, _ = r.ReadUintBE(4)
			chunkStream.exted = true
		} else {
			chunkStream.exted = false
		}
		chunkStream.timeDelta = timeStamp
		chunkStream.Timestamp += timeStamp
		chunkStream.new(pool)
	case 2:
		chunkStream.Format = chunkStream.tmpFromat
		timeStamp, _ := r.ReadUintBE(3)
		if timeStamp == 0xffffff {
			timeStamp, _ = r.ReadUintBE(4)
			chunkStream.exted = true
		} else {
			chunkStream.exted = false
		}
		chunkStream.timeDelta = timeStamp
		chunkStream.Timestamp += timeStamp
		chunkStream.new(pool)
	case 3:
		if chunkStream.remain == 0 {
			switch chunkStream.Format {
			case 0:
				if chunkStream.exted {
					timestamp, _ := r.ReadUintBE(4)
					chunkStream.Timestamp = timestamp
				}
			case 1, 2:
				var timedet uint32
				if chunkStream.exted {
					timedet, _ = r.ReadUintBE(4)
				} else {
					timedet = chunkStream.timeDelta
				}
				chunkStream.Timestamp += timedet
			}
			chunkStream.new(pool)
		} else {
			if chunkStream.exted {
				b, err := r.Peek(4)
				if err != nil {
					return err
				}
				tmpts := binary.BigEndian.Uint32(b)
				if tmpts == chunkStream.Timestamp {
					r.Discard(4)
				}
			}
		}
	default:
		return fmt.Errorf("invalid format=%d", chunkStream.Format)
	}
	size := int(chunkStream.remain)
	if size > int(chunkSize) {
		size = int(chunkSize)
	}

	buf := chunkStream.Data[chunkStream.index : chunkStream.index+uint32(size)]
	if _, err := r.Read(buf); err != nil {
		return err
	}
	chunkStream.index += uint32(size)
	chunkStream.remain -= uint32(size)
	if chunkStream.remain == 0 {
		chunkStream.got = true
	}

	return r.readError
}

const (
	_ = iota
	idSetChunkSize
	idAbortMessage
	idAck
	idUserControlMessages
	idWindowAckSize
	idSetPeerBandwidth
)

type Conn struct {
	net.Conn
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	rw                  *ReadWriter
	pool                *pool.Pool
	chunks              map[uint32]ChunkStream
}

func NewConn(c net.Conn, bufferSize int) *Conn {
	return &Conn{
		Conn:                c,
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		pool:                pool.NewPool(),
		rw:                  NewReadWriter(c, bufferSize),
		chunks:              make(map[uint32]ChunkStream),
	}
}

func (conn *Conn) Read(c *ChunkStream) error {
	for {
		h, _ := conn.rw.ReadUintBE(1)
		// if err != nil {
		// 	log.Println("read from conn error: ", err)
		// 	return err
		// }
		format := h >> 6
		csid := h & 0x3f
		cs, ok := conn.chunks[csid]
		if !ok {
			cs = ChunkStream{}
			conn.chunks[csid] = cs
		}
		cs.tmpFromat = format
		cs.CSID = csid
		err := cs.readChunk(conn.rw, conn.remoteChunkSize, conn.pool)
		if err != nil {
			return err
		}
		conn.chunks[csid] = cs
		if cs.full() {
			*c = cs
			break
		}
	}

	conn.handleControlMsg(c)

	conn.ack(c.Length)

	return nil
}

func (conn *Conn) Write(c *ChunkStream) error {
	if c.TypeID == idSetChunkSize {
		conn.chunkSize = binary.BigEndian.Uint32(c.Data)
	}
	return c.writeChunk(conn.rw, int(conn.chunkSize))
}

func (conn *Conn) Flush() error {
	return conn.rw.Flush()
}

func (conn *Conn) Close() error {
	return conn.Conn.Close()
}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.Conn.RemoteAddr()
}

func (conn *Conn) LocalAddr() net.Addr {
	return conn.Conn.LocalAddr()
}

func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.Conn.SetDeadline(t)
}

func (conn *Conn) NewAck(size uint32) ChunkStream {
	return initControlMsg(idAck, 4, size)
}

func (conn *Conn) NewSetChunkSize(size uint32) ChunkStream {
	return initControlMsg(idSetChunkSize, 4, size)
}

func (conn *Conn) NewWindowAckSize(size uint32) ChunkStream {
	return initControlMsg(idWindowAckSize, 4, size)
}

func (conn *Conn) NewSetPeerBandwidth(size uint32) ChunkStream {
	ret := initControlMsg(idSetPeerBandwidth, 5, size)
	ret.Data[4] = 2
	return ret
}

func (conn *Conn) handleControlMsg(c *ChunkStream) {
	if c.TypeID == idSetChunkSize {
		conn.remoteChunkSize = binary.BigEndian.Uint32(c.Data)
	} else if c.TypeID == idWindowAckSize {
		conn.remoteWindowAckSize = binary.BigEndian.Uint32(c.Data)
	}
}

func (conn *Conn) ack(size uint32) {
	conn.received += uint32(size)
	conn.ackReceived += uint32(size)
	if conn.received >= 0xf0000000 {
		conn.received = 0
	}
	if conn.ackReceived >= conn.remoteWindowAckSize {
		cs := conn.NewAck(conn.ackReceived)
		cs.writeChunk(conn.rw, int(conn.chunkSize))
		conn.ackReceived = 0
	}
}

func initControlMsg(id, size, value uint32) ChunkStream {
	ret := ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   id,
		StreamID: 0,
		Length:   size,
		Data:     make([]byte, size),
	}
	pio.PutU32BE(ret.Data[:size], value)
	return ret
}

const (
	streamBegin      uint32 = 0
	streamEOF        uint32 = 1
	streamDry        uint32 = 2
	setBufferLen     uint32 = 3
	streamIsRecorded uint32 = 4
	pingRequest      uint32 = 6
	pingResponse     uint32 = 7
)

/*
   +------------------------------+-------------------------
   |     Event Type ( 2- bytes )  | Event Data
   +------------------------------+-------------------------
   Pay load for the ‘User Control Message’.
*/
func (conn *Conn) userControlMsg(eventType, buflen uint32) ChunkStream {
	var ret ChunkStream
	buflen += 2
	ret = ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   4,
		StreamID: 1,
		Length:   buflen,
		Data:     make([]byte, buflen),
	}
	ret.Data[0] = byte(eventType >> 8 & 0xff)
	ret.Data[1] = byte(eventType & 0xff)
	return ret
}

func (conn *Conn) SetBegin() {
	ret := conn.userControlMsg(streamBegin, 4)
	for i := 0; i < 4; i++ {
		ret.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	conn.Write(&ret)
}

func (conn *Conn) SetRecorded() {
	ret := conn.userControlMsg(streamIsRecorded, 4)
	for i := 0; i < 4; i++ {
		ret.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	conn.Write(&ret)
}

var (
	respResult     = "_result"
	respError      = "_error"
	onStatus       = "onStatus"
	publishStart   = "NetStream.Publish.Start"
	playStart      = "NetStream.Play.Start"
	connectSuccess = "NetConnection.Connect.Success"
	onBWDone       = "onBWDone"
)

var (
//ErrFail = fmt.Errorf("Failed streaming!")
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
		case 20, 17:
			r := bytes.NewReader(rc.Data)
			vs, err := connClient.decoder.DecodeBatch(r, amf.AMF0)
			if err != nil && err != io.EOF {
				logger.Warning("Decoding batch: %v", err)
			}

			//logger.Debug("Reading Message From Server: %v", vs)
			for k, v := range vs {

				//
				switch v.(type) {
				case string:
					switch connClient.curcmdName {
					case cmdConnect, cmdCreateStream:
						if v.(string) != respResult {
							return fmt.Errorf("unexpected connection/creating stream result: %s", v.(string))
						}

					case cmdPublish:
						if v.(string) != onStatus {
							return fmt.Errorf("unexpected publish result: %s", v.(string))
						}
					}
				case float64:
					switch connClient.curcmdName {
					case cmdConnect, cmdCreateStream:
						id := int(v.(float64))

						if k == 1 {
							if id != connClient.transID {
								return fmt.Errorf("unexpected ID from server expected [%d] actual [%d]", id, connClient.transID)
							}
						} else if k == 3 {
							connClient.streamid = uint32(id)
						}
					case cmdPublish:
						if int(v.(float64)) != 0 {
							return fmt.Errorf("unable to publish to server")
						}
					}
				case amf.Object:
					objmap := v.(amf.Object)
					switch connClient.curcmdName {
					case cmdConnect:
						code, ok := objmap["code"]
						if ok && code.(string) != connectSuccess {
							return fmt.Errorf("unable to connect: error code: %d", code)
						}
					case cmdPublish:
						code, ok := objmap["code"]
						if ok && code.(string) != publishStart {
							return fmt.Errorf("unable to publish: error code: %d", code)
						}
					}
				}
			}

			return nil
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
	connClient.curcmdName = cmdConnect

	logger.Info("writeConnectMsg: connClient.transID=%d, event=%v", connClient.transID, event)
	if err := connClient.writeMsg(cmdConnect, connClient.transID, event); err != nil {
		return err
	}
	return connClient.readRespMsg()
}

func (connClient *ConnClient) writeCreateStreamMsg() error {
	connClient.transID++
	connClient.curcmdName = cmdCreateStream

	logger.Info("writeCreateStreamMsg: connClient.transID=%d", connClient.transID)
	if err := connClient.writeMsg(cmdCreateStream, connClient.transID, nil); err != nil {
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
	connClient.curcmdName = cmdPublish
	if err := connClient.writeMsg(cmdPublish, connClient.transID, nil, connClient.title, publishLive); err != nil {
		return err
	}
	return connClient.readRespMsg()
}

func (connClient *ConnClient) writePlayMsg() error {
	connClient.transID++
	connClient.curcmdName = cmdPlay
	logger.Info("writePlayMsg: connClient.transID=%d, cmdPlay=%v, connClient.title=%v",
		connClient.transID, cmdPlay, connClient.title)

	if err := connClient.writeMsg(cmdPlay, 0, nil, connClient.title); err != nil {
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
	//logger.Info("connection:", "local:", conn.LocalAddr(), "remote:", conn.RemoteAddr())

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

var (
	publishLive   = "live"
	publishRecord = "record"
	publishAppend = "append"
)

var (
	ErrReq = fmt.Errorf("req error")
)

var (
	cmdConnect         = "connect"
	cmdFcpublish       = "FCPublish"
	cmdReleaseStream   = "releaseStream"
	cmdCreateStream    = "createStream"
	cmdPublish         = "publish"
	cmdFCUnpublish     = "FCUnpublish"
	cmdDeleteStream    = "deleteStream"
	cmdPlay            = "play"
	cmdGetStreamLength = "getStreamLength"
)

type ConnectInfo struct {
	App            string `amf:"app" json:"app"`
	Flashver       string `amf:"flashVer" json:"flashVer"`
	SwfUrl         string `amf:"swfUrl" json:"swfUrl"`
	TcUrl          string `amf:"tcUrl" json:"tcUrl"`
	Fpad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageUrl        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

type ConnectResp struct {
	FMSVer       string `amf:"fmsVer"`
	Capabilities int    `amf:"capabilities"`
}

type ConnectEvent struct {
	Level          string `amf:"level"`
	Code           string `amf:"code"`
	Description    string `amf:"description"`
	ObjectEncoding int    `amf:"objectEncoding"`
}

type PublishInfo struct {
	Name string
	Type string
}

type ConnServer struct {
	done          bool
	streamID      int
	isPublisher   bool
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

func (connServer *ConnServer) writeMsg(csid, streamID uint32, args ...interface{}) error {
	connServer.bytesw.Reset()
	for _, v := range args {
		if _, err := connServer.encoder.Encode(connServer.bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := connServer.bytesw.Bytes()
	c := ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  streamID,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	connServer.conn.Write(&c)
	return connServer.conn.Flush()
}

func (connServer *ConnServer) connect(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != 1 {
				return ErrReq
			}
			connServer.transactionID = id
		case amf.Object:
			obimap := v.(amf.Object)
			if app, ok := obimap["app"]; ok {
				connServer.ConnInfo.App = app.(string)
			}
			if flashVer, ok := obimap["flashVer"]; ok {
				connServer.ConnInfo.Flashver = flashVer.(string)
			}
			if tcurl, ok := obimap["tcUrl"]; ok {
				connServer.ConnInfo.TcUrl = tcurl.(string)
			}
			if encoding, ok := obimap["objectEncoding"]; ok {
				connServer.ConnInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
	return nil
}

func (connServer *ConnServer) releaseStream(vs []interface{}) error {
	return nil
}

func (connServer *ConnServer) fcPublish(vs []interface{}) error {
	return nil
}

func (connServer *ConnServer) getStreamLength(cur *ChunkStream) error {
	//return connServer.writeMsg(cur.CSID, cur.StreamID,
	//	"_result", connServer.transactionID, nil, connServer.streamID)
	return nil
}

func (connServer *ConnServer) connectResp(cur *ChunkStream) error {
	c := connServer.conn.NewWindowAckSize(2500000)
	connServer.conn.Write(&c)
	c = connServer.conn.NewSetPeerBandwidth(2500000)
	connServer.conn.Write(&c)
	c = connServer.conn.NewSetChunkSize(uint32(1024))
	connServer.conn.Write(&c)

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = connServer.ConnInfo.ObjectEncoding
	return connServer.writeMsg(cur.CSID, cur.StreamID, "_result", connServer.transactionID, resp, event)
}

func (connServer *ConnServer) createStream(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			connServer.transactionID = int(v.(float64))
		case amf.Object:
		}
	}
	return nil
}

func (connServer *ConnServer) createStreamResp(cur *ChunkStream) error {
	return connServer.writeMsg(cur.CSID, cur.StreamID, "_result", connServer.transactionID, nil, connServer.streamID)
}

func (connServer *ConnServer) publishOrPlay(vs []interface{}) error {
	for k, v := range vs {
		switch v.(type) {
		case string:
			if k == 2 {
				connServer.PublishInfo.Name = v.(string)
			} else if k == 3 {
				connServer.PublishInfo.Type = v.(string)
			}
		case float64:
			id := int(v.(float64))
			connServer.transactionID = id
		case amf.Object:
		}
	}

	return nil
}

func (connServer *ConnServer) publishResp(cur *ChunkStream) error {
	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Publish.Start"
	event["description"] = "Start publishing."
	return connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event)
}

func (connServer *ConnServer) playResp(cur *ChunkStream) error {
	connServer.conn.SetRecorded()
	connServer.conn.SetBegin()

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Play.Reset"
	event["description"] = "Playing and resetting stream."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.Start"
	event["description"] = "Started playing stream."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Data.Start"
	event["description"] = "Started playing stream."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.PublishNotify"
	event["description"] = "Started playing notify."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}
	return connServer.conn.Flush()
}

func (connServer *ConnServer) handleCmdMsg(c *ChunkStream) error {
	amfType := amf.AMF0
	if c.TypeID == 17 {
		c.Data = c.Data[1:]
	}
	r := bytes.NewReader(c.Data)
	vs, err := connServer.decoder.DecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}
	// logger.Info("rtmp req: %#v", vs)
	switch vs[0].(type) {
	case string:
		switch vs[0].(string) {
		case cmdConnect:
			if err = connServer.connect(vs[1:]); err != nil {
				return err
			}
			if err = connServer.connectResp(c); err != nil {
				return err
			}
		case cmdCreateStream:
			if err = connServer.createStream(vs[1:]); err != nil {
				return err
			}
			if err = connServer.createStreamResp(c); err != nil {
				return err
			}
		case cmdPublish:
			if err = connServer.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = connServer.publishResp(c); err != nil {
				return err
			}
			connServer.done = true
			connServer.isPublisher = true
			logger.Info("Publish request complete")
		case cmdPlay:
			if err = connServer.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = connServer.playResp(c); err != nil {
				return err
			}
			connServer.done = true
			connServer.isPublisher = false
			//logger.Info("Play request")
		case cmdFcpublish:
			connServer.fcPublish(vs)
		case cmdReleaseStream:
			connServer.releaseStream(vs)
		case cmdFCUnpublish:
		case cmdDeleteStream:
		case cmdGetStreamLength:
		default:
			logger.Critical("Unknown command: %s", vs[0].(string))
		}
	}

	return nil
}

func (connServer *ConnServer) ReadMsg() error {
	var c ChunkStream
	for {
		if err := connServer.conn.Read(&c); err != nil {
			return err
		}
		switch c.TypeID {
		case 20, 17:
			if err := connServer.handleCmdMsg(&c); err != nil {
				return err
			}
		}
		if connServer.done {
			break
		}
	}
	return nil
}

func (connServer *ConnServer) IsPublisher() bool {
	return true
	return connServer.isPublisher
}

func (connServer *ConnServer) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.DEL); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return connServer.conn.Write(&c)
}

func (connServer *ConnServer) Flush() error {
	return connServer.conn.Flush()
}

func (connServer *ConnServer) Read(c *ChunkStream) (err error) {
	return connServer.conn.Read(c)
}

func (connServer *ConnServer) GetInfo() (app string, name string, url string) {
	app = connServer.ConnInfo.App
	name = connServer.PublishInfo.Name
	url = connServer.ConnInfo.TcUrl + "/" + connServer.PublishInfo.Name
	return
}

func (connServer *ConnServer) Close(err error) {
	connServer.conn.Close()
}

var (
	timeout = 5 * time.Second
)

var (
	hsClientFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsServerFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsClientPartialKey = hsClientFullKey[:30]
	hsServerPartialKey = hsServerFullKey[:36]
)

func hsMakeDigest(key []byte, src []byte, gap int) (dst []byte) {
	h := hmac.New(sha256.New, key)
	if gap <= 0 {
		h.Write(src)
	} else {
		h.Write(src[:gap])
		h.Write(src[gap+32:])
	}
	return h.Sum(nil)
}

func hsCalcDigestPos(p []byte, base int) (pos int) {
	for i := 0; i < 4; i++ {
		pos += int(p[base+i])
	}
	pos = (pos % 728) + base + 4
	return
}

func hsFindDigest(p []byte, key []byte, base int) int {
	gap := hsCalcDigestPos(p, base)
	digest := hsMakeDigest(key, p, gap)
	if bytes.Compare(p[gap:gap+32], digest) != 0 {
		return -1
	}
	return gap
}

func hsParse1(p []byte, peerkey []byte, key []byte) (ok bool, digest []byte) {
	var pos int
	if pos = hsFindDigest(p, peerkey, 772); pos == -1 {
		if pos = hsFindDigest(p, peerkey, 8); pos == -1 {
			return
		}
	}
	ok = true
	digest = hsMakeDigest(key, p[pos:pos+32], -1)
	return
}

func hsCreate01(p []byte, time uint32, ver uint32, key []byte) {
	p[0] = 3
	p1 := p[1:]
	rand.Read(p1[8:])
	pio.PutU32BE(p1[0:4], time)
	pio.PutU32BE(p1[4:8], ver)
	gap := hsCalcDigestPos(p1, 8)
	digest := hsMakeDigest(key, p1, gap)
	copy(p1[gap:], digest)
}

func hsCreate2(p []byte, key []byte) {
	rand.Read(p)
	gap := len(p) - 32
	digest := hsMakeDigest(key, p, gap)
	copy(p[gap:], digest)
}

func (conn *Conn) HandshakeClient() (err error) {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]

	C0[0] = 3
	// > C0C1
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = conn.rw.Write(C0C1); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if err = conn.rw.Flush(); err != nil {
		return
	}

	// < S0S1S2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = io.ReadFull(conn.rw, S0S1S2); err != nil {
		return
	}

	S1 := S0S1S2[1 : 1536+1]
	if ver := pio.U32BE(S1[4:8]); ver != 0 {
		C2 = S1
	} else {
		C2 = S1
	}

	// > C2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = conn.rw.Write(C2); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Time{})
	return
}

func (conn *Conn) HandshakeServer() (err error) {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S0S1 := S0S1S2[:1536+1]
	S2 := S0S1S2[1536+1:]

	// < C0C1
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = io.ReadFull(conn.rw, C0C1); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if C0[0] != 3 {
		err = fmt.Errorf("rtmp: handshake version=%d invalid", C0[0])
		return
	}

	S0[0] = 3

	clitime := pio.U32BE(C1[0:4])
	srvtime := clitime
	srvver := uint32(0x0d0e0a0d)
	cliver := pio.U32BE(C1[4:8])

	if cliver != 0 {
		var ok bool
		var digest []byte
		if ok, digest = hsParse1(C1, hsClientPartialKey, hsServerFullKey); !ok {
			err = fmt.Errorf("rtmp: handshake server: C1 invalid")
			return
		}
		hsCreate01(S0S1, srvtime, srvver, hsServerPartialKey)
		hsCreate2(S2, digest)
	} else {
		copy(S1, C2)
		copy(S2, C1)
	}

	// > S0S1S2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = conn.rw.Write(S0S1S2); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if err = conn.rw.Flush(); err != nil {
		return
	}

	// < C2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = io.ReadFull(conn.rw, C2); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Time{})
	return
}

type ReadWriter struct {
	*bufio.ReadWriter
	readError  error
	writeError error
}

func NewReadWriter(rw io.ReadWriter, bufSize int) *ReadWriter {
	return &ReadWriter{
		ReadWriter: bufio.NewReadWriter(bufio.NewReaderSize(rw, bufSize), bufio.NewWriterSize(rw, bufSize)),
	}
}

func (rw *ReadWriter) Read(p []byte) (int, error) {
	if rw.readError != nil {
		return 0, rw.readError
	}
	n, err := io.ReadAtLeast(rw.ReadWriter, p, len(p))
	rw.readError = err
	return n, err
}

func (rw *ReadWriter) ReadError() error {
	return rw.readError
}

func (rw *ReadWriter) ReadUintBE(n int) (uint32, error) {
	if rw.readError != nil {
		return 0, rw.readError
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		b, err := rw.ReadByte()
		if err != nil {
			rw.readError = err
			return 0, err
		}
		ret = ret<<8 + uint32(b)
	}
	return ret, nil
}

func (rw *ReadWriter) ReadUintLE(n int) (uint32, error) {
	if rw.readError != nil {
		return 0, rw.readError
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		b, err := rw.ReadByte()
		if err != nil {
			rw.readError = err
			return 0, err
		}
		ret += uint32(b) << uint32(i*8)
	}
	return ret, nil
}

func (rw *ReadWriter) Flush() error {
	if rw.writeError != nil {
		return rw.writeError
	}

	if rw.ReadWriter.Writer.Buffered() == 0 {
		return nil
	}
	return rw.ReadWriter.Flush()
}

func (rw *ReadWriter) Write(p []byte) (int, error) {
	if rw.writeError != nil {
		return 0, rw.writeError
	}
	return rw.ReadWriter.Write(p)
}

func (rw *ReadWriter) WriteError() error {
	return rw.writeError
}

func (rw *ReadWriter) WriteUintBE(v uint32, n int) error {
	if rw.writeError != nil {
		return rw.writeError
	}
	for i := 0; i < n; i++ {
		b := byte(v>>uint32((n-i-1)<<3)) & 0xff
		if err := rw.WriteByte(b); err != nil {
			rw.writeError = err
			return err
		}
	}
	return nil
}

func (rw *ReadWriter) WriteUintLE(v uint32, n int) error {
	if rw.writeError != nil {
		return rw.writeError
	}
	for i := 0; i < n; i++ {
		b := byte(v) & 0xff
		if err := rw.WriteByte(b); err != nil {
			rw.writeError = err
			return err
		}
		v = v >> 8
	}
	return nil
}
