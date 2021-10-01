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
	"github.com/gwuhaolin/livego/utils/pool"
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
	TimeoutDurationSeconds time.Duration = 5 * time.Second
)

const (

	// Publish
	// RTMP Spec 7.2.2.6

	PublishCommandLive   string = "live"
	PublishCommandRecord string = "record"
	PublishCommandAppend string = "append"
)

const (

	// Control Commands
	// RTMP Spec 5.4
	// Protocol Control Messages

	CommandConnect         string = "connect"
	CommandReleaseStream   string = "releaseStream"
	CommandCreateStream    string = "createStream"
	CommandPlay            string = "play"
	CommandPublish         string = "publish"
	CommandDeleteStream    string = "deleteStream"
	CommandGetStreamLength string = "getStreamLength"
	CommandFCPublish       string = "FCPublish"
	CommandFCUnpublish     string = "FCUnpublish"

	CommandType_Result             = "_result"
	CommandType_Error              = "_error"
	CommandOnStatus                = "CommandOnStatus"
	CommandNetStreamPublishStart   = "NetStream.Publish.Start"
	CommandNetStreamPlayStart      = "NetStream.Play.Start"
	CommandNetStreamConnectSuccess = "NetConnection.Connect.Success"
	CommandOnBWDone                = "CommandOnBWDone"
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
