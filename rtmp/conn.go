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
	"net"
	"time"

	"github.com/kris-nova/logger"

	"github.com/gwuhaolin/livego/utils/pool"
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
		h, err := conn.rw.ReadUintBE(1)
		if err != nil {
			logger.Warning("Reading chunk from client: %v", err)
			return err
		}
		logger.Debug("2")
		format := h >> 6
		csid := h & 0x3f
		cs, ok := conn.chunks[csid]
		if !ok {
			cs = ChunkStream{}
			conn.chunks[csid] = cs
		}
		cs.tmpFromat = format
		cs.CSID = csid
		logger.Debug("3")

		err = cs.readChunk(conn.rw, conn.remoteChunkSize, conn.pool)
		if err != nil {
			return err
		}
		logger.Debug("4")

		conn.chunks[csid] = cs
		if cs.full() {
			*c = cs
			break
		}
	}

	logger.Debug("5")

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
	logger.Info("conn: handleControlMsg")
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
