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
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/gwuhaolin/livego/utils/pio"

	"github.com/gwuhaolin/livego/av"
)

var (
	HandshakeClientKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	HandshakeServerKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}

	HandshakeClientPartial30 []byte = HandshakeClientKey[:30]
	HandshakeServerPartial36 []byte = HandshakeServerKey[:36]
)

const (

	// TimeoutDurationSeconds is the timeout used for all
	// connection timeouts.
	TimeoutDurationSeconds time.Duration = 5 * time.Second

	// Chunk Size
	// 5.4.1 Set Chunk Size
	// The maximum chunk size defaults to 128 bytes, but the client or the
	// server can change this value, and updates its peer using this
	// message.

	DefaultRTMPChunkSizeBytes             uint32 = 128
	DefaultRTMPChunkSizeBytesLarge        uint32 = DefaultRTMPChunkSizeBytes * 64
	DefaultWindowAcknowledgementSizeBytes uint32 = 2500000
	DefaultPeerBandwidthSizeBytes         uint32 = 2500000
	DefaultMaximumPoolSizeBytes           int    = 1024 * 1024 * 512
	DefaultConnBufferSizeBytes            int    = 1024 * 1024 * 512
	DefaultServerFMSVersion               string = "FMS/3,0,1,123"

	ClientMethodPlay    ClientMethod = "play"
	ClientMethodPublish ClientMethod = "publish"

	// Publish
	// RTMP Spec 7.2.2.6

	PublishCommandLive   string = "live"
	PublishCommandRecord string = "record"
	PublishCommandAppend string = "append"

	// Control Commands
	// RTMP Spec 5.4
	// Protocol Control Messages

	// 7.1.1. Command Message (20, 17)
	//
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

	CommandConnect         string = "connect"
	CommandReleaseStream   string = "releaseStream"
	CommandCreateStream    string = "createStream"
	CommandPlay            string = "play"
	CommandPublish         string = "publish"
	CommandDeleteStream    string = "deleteStream"
	CommandGetStreamLength string = "getStreamLength"
	CommandFCPublish       string = "FCPublish"
	CommandFCUnpublish     string = "FCUnpublish"

	// 7.1.1. Command Message (20, 17)
	//
	// Command message responses.
	//
	// These are used to inform the sender about the status of the requested commands.

	CommandType_Result  = "_result"
	CommandType_Error   = "_error"
	CommandTypeOnStatus = "onStatus"

	// 7.2.2. NetStream Commands
	//
	// The NetStream defines the channel through which the streaming audio,
	// video, and data messages can flow over the NetConnection that
	// connects the client to the server. A NetConnection object can
	// support multiple NetStreams for multiple data streams.

	CommandNetStreamPublishStart   = "NetStream.Publish.Start"
	CommandNetStreamPublishNotify  = "NetStream.Publish.Notify"
	CommandNetStreamPlayStart      = "NetStream.Play.Start"
	CommandNetStreamPlayReset      = "NetStream.Play.Reset"
	CommandNetStreamDataStart      = "NetStream.Data.Start"
	CommandNetStreamConnectSuccess = "NetConnection.Connect.Success"
	CommandOnBWDone                = "CommandOnBWDone"

	StreamBegin      uint32 = 0
	StreamEOF        uint32 = 1
	StreamDry        uint32 = 2
	SetBufferLen     uint32 = 3
	StreamIsRecorded uint32 = 4
	PingRequest      uint32 = 6
	PingResponse     uint32 = 7
)

type ClientMethod string

// ChunkStream
//
// 5.1.
//
// The format of a message that can be split into chunks to support
// multiplexing depends on a higher level protocol. The message format
// SHOULD however contain the following fields which are necessary for
// creating the chunks.
//
type ChunkStream struct {
	// Timestamp of the message. This field can transport 4
	// bytes.
	Timestamp uint32

	// Length of the message payload. If the message header cannot
	// be elided, it should be included in the length. This field
	// occupies 3 bytes in the chunk header.
	Length uint32

	// A range of type IDs are reserved for protocol control
	// messages. These messages which propagate information are handled
	// by both RTMP Chunk Stream protocol and the higher-level protocol.
	// All other type IDs are available for use by the higher-level
	// protocol, and treated as opaque values by RTMP Chunk Stream. In
	// fact, nothing in RTMP Chunk Stream requires these values to be
	// used as a type; all (non-protocol) messages could be of the same
	// type, or the application could use this field to distinguish
	// simultaneous tracks rather than types. This field occupies 1 byte
	// in the chunk header.
	TypeID uint32

	// The message stream ID can be any arbitrary value.
	// Different message streams multiplexed onto the same chunk stream
	// are demultiplexed based on their message stream IDs. Beyond that,
	// as far as RTMP Chunk Stream is concerned, this is an opaque value.
	// This field occupies 4 bytes in the chunk header in little endian
	// format.
	StreamID uint32

	// Data is the set of bytes in the Chunk. The chunk payload.
	Data []byte

	Format    uint32
	CSID      uint32
	timeDelta uint32
	exited    bool
	index     uint32
	remain    uint32
	got       bool
	tmpFormat uint32
}

func (chunkStream *ChunkStream) full() bool {
	return chunkStream.got
}

func (chunkStream *ChunkStream) new(pool *Pool) {
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

func (chunkStream *ChunkStream) readChunk(r *ReadWriter, chunkSize uint32, pool *Pool) error {
	if chunkStream.remain != 0 && chunkStream.tmpFormat != 3 {
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

	switch chunkStream.tmpFormat {
	case 0:
		chunkStream.Format = chunkStream.tmpFormat
		chunkStream.Timestamp, _ = r.ReadUintBE(3)
		chunkStream.Length, _ = r.ReadUintBE(3)
		chunkStream.TypeID, _ = r.ReadUintBE(1)
		chunkStream.StreamID, _ = r.ReadUintLE(4)
		if chunkStream.Timestamp == 0xffffff {
			chunkStream.Timestamp, _ = r.ReadUintBE(4)
			chunkStream.exited = true
		} else {
			chunkStream.exited = false
		}
		chunkStream.new(pool)
	case 1:
		chunkStream.Format = chunkStream.tmpFormat
		timeStamp, _ := r.ReadUintBE(3)
		chunkStream.Length, _ = r.ReadUintBE(3)
		chunkStream.TypeID, _ = r.ReadUintBE(1)
		if timeStamp == 0xffffff {
			timeStamp, _ = r.ReadUintBE(4)
			chunkStream.exited = true
		} else {
			chunkStream.exited = false
		}
		chunkStream.timeDelta = timeStamp
		chunkStream.Timestamp += timeStamp
		chunkStream.new(pool)
	case 2:
		chunkStream.Format = chunkStream.tmpFormat
		timeStamp, _ := r.ReadUintBE(3)
		if timeStamp == 0xffffff {
			timeStamp, _ = r.ReadUintBE(4)
			chunkStream.exited = true
		} else {
			chunkStream.exited = false
		}
		chunkStream.timeDelta = timeStamp
		chunkStream.Timestamp += timeStamp
		chunkStream.new(pool)
	case 3:
		if chunkStream.remain == 0 {
			switch chunkStream.Format {
			case 0:
				if chunkStream.exited {
					timestamp, _ := r.ReadUintBE(4)
					chunkStream.Timestamp = timestamp
				}
			case 1, 2:
				var timedet uint32
				if chunkStream.exited {
					timedet, _ = r.ReadUintBE(4)
				} else {
					timedet = chunkStream.timeDelta
				}
				chunkStream.Timestamp += timedet
			}
			chunkStream.new(pool)
		} else {
			if chunkStream.exited {
				b, err := r.Peek(4)
				if err != nil {
					return TestableEOFError
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

func newChunkStream(typeID, length, payload uint32) ChunkStream {
	ret := ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   typeID,
		StreamID: 0,
		Length:   length,
		Data:     make([]byte, length),
	}
	pio.PutU32BE(ret.Data[:length], payload)
	return ret
}

// ConnectInfo is the RTMP spec's parameters of the key value pairs passed
// during a connect command message
//
// The following is the description of the name-value pairs used in Command
//                      Object of the connect command.
//   +-----------+--------+-----------------------------+----------------+
//   | Property  |  Type  |        Description          | Example Value  |
//   +-----------+--------+-----------------------------+----------------+
//   |   app     | String | The Server application name |    testapp     |
//   |           |        | the client is connected to. |                |
//   +-----------+--------+-----------------------------+----------------+
//   | flashver  | String | Flash Player version. It is |    FMSc/1.0    |
//   |           |        | the same string as returned |                |
//   |           |        | by the ApplicationScript    |                |
//   |           |        | getversion () function.     |                |
//   +-----------+--------+-----------------------------+----------------+
//   |  swfUrl   | String | URL of the source SWF file  | file://C:/     |
//   |           |        | making the connection.      | FlvPlayer.swf  |
//   +-----------+--------+-----------------------------+----------------+
//   |  tcUrl    | String | URL of the Server.          | rtmp://local   |
//   |           |        | It has the following format.| host:1935/test |
//   |           |        | protocol://servername:port/ | app/instance1  |
//   |           |        | appName/appInstance         |                |
//   +-----------+--------+-----------------------------+----------------+
//   |  fpad     | Boolean| True if proxy is being used.| true or false  |
//   +-----------+--------+-----------------------------+----------------+
//   |audioCodecs| Number | Indicates what audio codecs | SUPPORT_SND    |
//   |           |        | the client supports.        | _MP3           |
//   +-----------+--------+-----------------------------+----------------+
//   |videoCodecs| Number | Indicates what video codecs | SUPPORT_VID    |
//   |           |        | are supported.              | _SORENSON      |
//   +-----------+--------+-----------------------------+----------------+
//   |videoFunct-| Number | Indicates what special video| SUPPORT_VID    |
//   |ion        |        | functions are supported.    | _CLIENT_SEEK   |
//   +-----------+--------+-----------------------------+----------------+
//   |  pageUrl  | String | URL of the web page from    | http://        |
//   |           |        | where the SWF file was      | somehost/      |
//   |           |        | loaded.                     | sample.html    |
//   +-----------+--------+-----------------------------+----------------+
//   | object    | Number | AMF encoding method.        |     AMF3       |
//   | Encoding  |        |                             |                |
//   +-----------+--------+-----------------------------+----------------+
type ConnectInfo struct {
	App            string `amf:"app" json:"app"`
	FlashVer       string `amf:"flashVer" json:"flashVer"`
	SwfUrl         string `amf:"swfUrl" json:"swfUrl"`
	TcUrl          string `amf:"tcUrl" json:"tcUrl"`
	Fpad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageUrl        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

const (
	ConnInfoKeyApp         string = "app"
	ConnInfoKeyType        string = "type"
	ConnInfoKeyTcURL       string = "tcUrl"
	ConnInfoKeyFlashVer    string = "flashVer"
	ConnInfoObjectEncoding string = "objectEncoding"
)

type ConnResp struct {
	FMSVer       string `amf:"fmsVer"`
	Capabilities int    `amf:"capabilities"`
}

const (
	ConnRespFMSVer       string = "fmsVer"
	ConnRespCapabilities string = "capabilities"
)

type ConnEvent struct {
	Level          string `amf:"level"`
	Code           string `amf:"code"`
	Description    string `amf:"description"`
	ObjectEncoding int    `amf:"objectEncoding"`
}

const (
	ConnEventLevel          string = "level"
	ConnEventCode           string = "code"
	ConnEventDescription    string = "description"
	ConnEventObjectEncoding string = "objectEncoding"
	ConnEventStatus         string = "status"
)

type PublishInfo struct {
	Name string
	Type string
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

// MessageID is the main "Message stream ID" for each packet
// sent over the RTMP protocol
//
// 3. Definitions
// Message stream ID: Each message has an ID associated with it to
// identify the message stream in which it is flowing.
type MessageID uint32

const (

	// Protocol Control Messages
	// 5.4

	// 5.4.1 Set Chunk Size

	SetChunkSizeMessage   string = "setChunkSize"
	SetChunkSizeMessageID uint32 = 1

	// 5.4.2 Abort Messages

	AbortMessage   string = "abort"
	AbortMessageID uint32 = 2

	// 5.4.3. Acknowledgement

	AcknowledgementMessage   string = "acknowledgement"
	AcknowledgementMessageID uint32 = 3

	// 5.4.4. Window Acknowledgement Size

	WindowAcknowledgementSizeMessage   string = "windowAcknowledgementSize"
	WindowAcknowledgementSizeMessageID uint32 = 5

	// 5.4.5. Set Peer Bandwidth

	SetPeerBandwidthMessage   string = "setPeerBandwidth"
	SetPeerBandwidthMessageID uint32 = 6

	// 6.2. User Control Messages

	UserControlMessage   string = "userControl"
	UserControlMessageID uint32 = 4

	// RTMP Message Types
	// 7 and 7.1

	// 7.1.1 Command Message

	CommandMessage       string = "command"
	CommandMessageAMF3ID uint32 = 17
	CommandMessageAMF0ID uint32 = 20

	// 7.1.2 Data Message

	DataMessage       string = "data"
	DataMessageAMF3ID uint32 = 15
	DataMessageAMF0ID uint32 = 18

	// 7.1.3 Shared Object Message

	SharedObjectMessage       string = "sharedObject"
	SharedObjectMessageAMF3ID uint32 = 16
	SharedObjectMessageAMF0ID uint32 = 19

	// 7.1.4 Audio Message

	AudioMessage   string = "audio"
	AudioMessageID uint32 = 8

	// 7.1.5 Video Message

	VideoMessage   string = "video"
	VideoMessageID uint32 = 9

	// 7.1.6 Aggregate Message

	AggregateMessage   string = "aggregate"
	AggregateMessageID uint32 = 22

	// UnknownMessageID should never happen, but we default
	// all unknown message type IDs to this string
	UnknownMessageID = "UNKNOWN"
)

// chunkTypeLabel will return the label for the type of chunk based on it's type ID.
func typeIDString(chunk *ChunkStream) string {
	return typeIDStringUint32(chunk.TypeID)
}

// chunkTypeIDLabel will return the label for the type ID of a given RTMP chunk.
func typeIDStringUint32(id uint32) string {
	switch id {
	case SetChunkSizeMessageID:
		return SetChunkSizeMessage
	case AbortMessageID:
		return AbortMessage
	case AcknowledgementMessageID:
		return AcknowledgementMessage
	case WindowAcknowledgementSizeMessageID:
		return WindowAcknowledgementSizeMessage
	case SetPeerBandwidthMessageID:
		return SetPeerBandwidthMessage
	case UserControlMessageID:
		return UserControlMessage
	case CommandMessageAMF0ID, CommandMessageAMF3ID:
		return CommandMessage
	case DataMessageAMF0ID, DataMessageAMF3ID:
		return DataMessage
	case SharedObjectMessageAMF0ID, SharedObjectMessageAMF3ID:
		return SharedObjectMessage
	case AudioMessageID:
		return AudioMessage
	case VideoMessageID:
		return VideoMessage
	case AggregateMessageID:
		return AggregateMessage
	default:
		return UnknownMessageID
	}
	return UnknownMessageID
}
