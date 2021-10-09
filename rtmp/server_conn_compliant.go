// Copyright © 2021 Kris Nóva <kris@nivenly.com>
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
	"fmt"
	"io"
	"time"

	"github.com/gwuhaolin/livego/utils/pio"
	"github.com/kris-nova/logger"
)

func (s *ServerConn) handshake() error {
	var err error
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
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = io.ReadFull(s.conn.rw, C0C1); err != nil {
		return err
	}
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if C0[0] != 3 {
		return fmt.Errorf("rtmp: handshake version=%d invalid", C0[0])
	}

	S0[0] = 3

	clitime := pio.U32BE(C1[0:4])
	srvtime := clitime
	srvver := uint32(0x0d0e0a0d)
	cliver := pio.U32BE(C1[4:8])

	if cliver != 0 {
		var ok bool
		var digest []byte
		if ok, digest = hsParse1(C1, HandshakeClientPartial30, HandshakeServerKey); !ok {
			err = fmt.Errorf("rtmp: handshake server: C1 invalid")
			return err
		}
		hsCreate01(S0S1, srvtime, srvver, HandshakeServerPartial36)
		hsCreate2(S2, digest)
	} else {
		copy(S1, C2)
		copy(S2, C1)
	}

	// > S0S1S2
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = s.conn.rw.Write(S0S1S2); err != nil {
		return err
	}
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if err = s.conn.rw.Flush(); err != nil {
		return err
	}

	// < C2
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = io.ReadFull(s.conn.rw, C2); err != nil {
		return err
	}
	s.conn.SetDeadline(time.Time{})
	logger.Debug(rtmpMessage(thisFunctionName(), hs))
	return nil
}

func (s *ServerConn) connectRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return nil
}

func (s *ServerConn) connectTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) createStreamRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return nil
}

func (s *ServerConn) createStreamTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) playRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), ack))
	return nil
}

func (s *ServerConn) playTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) play2RX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) play2TX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) deleteStreamRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) deleteStreamTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) receiveAudioRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) receiveAudioTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) receiveVideoRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) receiveVideoTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) publishRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), ack))
	return nil
}

func (s *ServerConn) publishTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) seekRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) seekTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) pauseRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) pauseTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}
