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
	"encoding/binary"
	"errors"
	"net"
	"time"
)

// Conn
//
// Conn is the base type for all RTMP connections.
// Conn is embedded with native Go types
//    - net.Conn
//    - net.Addr
//    - url.URL
//
// Both ConnClient and ConnServer are extensions of base Conn
type Conn struct {
	net.Conn
	URLAddr

	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	rw                  *ReadWriter
	pool                *Pool
	chunks              map[uint32]ChunkStream
}

func NewConnFromNetConn(c net.Conn) *Conn {
	conn := &Conn{
		Conn:                c,
		chunkSize:           DefaultRTMPChunkSizeBytes,
		remoteChunkSize:     DefaultRTMPChunkSizeBytes,
		windowAckSize:       DefaultWindowAcknowledgementSizeBytes,
		remoteWindowAckSize: DefaultWindowAcknowledgementSizeBytes,
		pool:                NewPool(),
		rw:                  NewReadWriter(c, DefaultConnBufferSizeBytes),
		chunks:              make(map[uint32]ChunkStream),
	}
	return conn
}

func NewConnFromURLAddr(urladdr *URLAddr) (*Conn, error) {
	c, err := net.Dial(DefaultProtocol, urladdr.Host())
	if err != nil {
		return nil, err
	}
	conn := NewConnFromNetConn(c)
	conn.URLAddr = *urladdr
	return conn, nil
}

func NewConn(raw string) (*Conn, error) {
	urladdr, err := NewURLAddr(raw)
	if err != nil {
		return nil, err
	}
	return NewConnFromURLAddr(urladdr)
}

var (
	TestableEOFError = errors.New("reading bytes from client: EOF")
)

func (conn *Conn) Read(c *ChunkStream) error {

	// Read big endian bytes from the conn until we build a complete
	// chunk based on the chunk stream length and the chunk
	// headers.
	for {
		h, err := conn.rw.ReadUintBE(1)
		if err != nil {
			return TestableEOFError
		}
		format := h >> 6
		csid := h & 0x3f
		cs, ok := conn.chunks[csid]
		if !ok {
			cs = ChunkStream{}
			conn.chunks[csid] = cs
		}
		cs.tmpFormat = format
		cs.CSID = csid
		err = cs.readChunk(conn.rw, conn.remoteChunkSize, conn.pool)
		if err != nil {
			return TestableEOFError
		}

		conn.chunks[csid] = cs

		if cs.full() {
			*c = cs
			break
		}
	}

	// RTMP Can update chunk size so let's just check.
	// This is also required for our tests.
	if c.TypeID == SetChunkSizeMessageID {
		chunkSize := binary.BigEndian.Uint32(c.Data)
		//logger.Debug("  ---> in Read() chunk size set: %d", chunkSize)
		conn.remoteChunkSize = chunkSize
	} else if c.TypeID == WindowAcknowledgementSizeMessageID {
		conn.remoteWindowAckSize = binary.BigEndian.Uint32(c.Data)
	}

	// We should now have a complete chunk.

	return nil
}

func (conn *Conn) Write(c *ChunkStream) error {
	if c.TypeID == SetChunkSizeMessageID {
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
	return newChunkStream(AcknowledgementMessageID, 4, size)
}

func (conn *Conn) NewSetChunkSize(size uint32) ChunkStream {
	return newChunkStream(SetChunkSizeMessageID, 4, size)
}

func (conn *Conn) NewWindowAckSize(size uint32) ChunkStream {
	return newChunkStream(AcknowledgementMessageID, 4, size)
}

func (conn *Conn) NewSetPeerBandwidth(size uint32) ChunkStream {
	ret := newChunkStream(SetPeerBandwidthMessageID, 5, size)
	ret.Data[4] = 2
	return ret
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

func (conn *Conn) messageUserControlStreamBegin() {
	ret := conn.userControlMsg(StreamBegin, 4)
	for i := 0; i < 4; i++ {
		ret.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	conn.Write(&ret)
}

func (conn *Conn) SetRecorded() {
	ret := conn.userControlMsg(StreamIsRecorded, 4)
	for i := 0; i < 4; i++ {
		ret.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	conn.Write(&ret)
}
