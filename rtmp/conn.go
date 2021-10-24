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
// Both ClientConn and ServerConn are extensions of base Conn
type Conn struct {
	net.Conn
	URLAddr
	chunkSize uint32
	//remoteChunkSize     uint32
	windowAckSize uint32
	//remoteWindowAckSize uint32
	received    uint32
	ackReceived uint32
	rw          *ReadWriter
	pool        *Pool
	chunks      map[uint32]ChunkStream
}

func NewConn(c net.Conn) *Conn {
	conn := &Conn{
		Conn:      c,
		chunkSize: DefaultRTMPChunkSizeBytes,
		//remoteChunkSize:     DefaultRTMPChunkSizeBytes,
		windowAckSize: DefaultWindowAcknowledgementSizeBytes,
		//remoteWindowAckSize: DefaultWindowAcknowledgementSizeBytes,
		pool:   NewPool(),
		rw:     NewReadWriter(c, DefaultConnBufferSizeBytes),
		chunks: make(map[uint32]ChunkStream),
	}
	return conn
}

var (
	WellKnownClosedClientError = errors.New("closed 🛑 client: EOF")
)

func (conn *Conn) Read(c *ChunkStream) error {

	// Read big endian bytes from the conn until we build a complete
	// chunk based on the chunk stream length and the chunk
	// headers.
	for {
		h, err := conn.rw.ReadUintBE(1)
		if err != nil {
			return WellKnownClosedClientError
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
		err = cs.readChunk(conn.rw, conn.chunkSize, conn.pool)
		if err != nil {
			return WellKnownClosedClientError
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
		conn.chunkSize = chunkSize
	} else if c.TypeID == WindowAcknowledgementSizeMessageID {
		conn.windowAckSize = binary.BigEndian.Uint32(c.Data)
	}

	// We should now have a complete chunk.

	//go conn.ack(c.Length)

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

func (conn *Conn) Close() {
	conn.Conn.Close()
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

// ==================================================================================================

func (conn *Conn) newChunkStreamWindowAcknowledgementMessage(size uint32) *ChunkStream {
	return newChunkStream(AcknowledgementMessageID, 4, size)
}

func (conn *Conn) newChunkStreamSetPeerBandwidth(size uint32) *ChunkStream {
	x := newChunkStream(SetPeerBandwidthMessageID, 5, size)
	// The Limit Type is one of the following values:
	//
	//   0 - Hard:  The peer SHOULD limit its output bandwidth to the
	//      indicated window size.
	//   1 - Soft:  The peer SHOULD limit its output bandwidth to the the
	//      window indicated in this message or the limit already in effect,
	//      whichever is smaller.
	//   2 - Dynamic:  If the previous Limit Type was Hard, treat this message
	//      as though it was marked Hard, otherwise ignore this message.
	x.Data[4] = 2
	return x
}

func (conn *Conn) newChunkStreamSetChunkSize(size uint32) *ChunkStream {
	return newChunkStream(SetChunkSizeMessageID, 4, size)
}

func (conn *Conn) newChunkStreamAck(size uint32) *ChunkStream {
	return newChunkStream(AcknowledgementMessageID, 4, size)
}

func (conn *Conn) ack(size uint32) error {
	conn.received += uint32(size)
	conn.ackReceived += uint32(size)
	if conn.received >= 0xf0000000 {
		conn.received = 0
	}
	if conn.ackReceived >= conn.windowAckSize {
		cs := conn.newChunkStreamAck(conn.ackReceived)
		err := cs.writeChunk(conn.rw, int(conn.chunkSize))
		conn.ackReceived = 0
		if err != nil {
			return err
		}
	}
	return nil
}

func (conn *Conn) setRecorded() error {
	ret := conn.userControlMsg(StreamIsRecorded, 4)
	for i := 0; i < 4; i++ {
		ret.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	return conn.Write(&ret)
}

func (conn *Conn) streamBegin() *ChunkStream {
	x := conn.userControlMsg(StreamBegin, 4)
	for i := 0; i < 4; i++ {
		x.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	return &x
}
