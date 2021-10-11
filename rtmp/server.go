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
	"net"

	"github.com/kris-nova/logger"
)

const (
	// MaximumPacketQueueRecords
	//
	// Similar to how HTTP-based transmission protocols like
	// HLS and DASH behave, RTMP too, breaks a multimedia
	// stream into fragments that are usually:
	//  - 64 bytes for audio
	//  - 128 bytes for video
	//
	// The size of the fragments can be negotiated between the client and the server.
	MaximumPacketQueueRecords int = 1024

	// SaveStaticMediaTimeIntervalMilliseconds
	//
	// The time delay (in MS) between saving static media
	SaveStaticMediaTimeIntervalMilliseconds int64 = 5000

	ReadTimeout  int = 10
	WriteTimeout int = 10
)

type Server struct {
	muxdem   *SafeMuxDemuxService
	listener *Listener
	conn     *ServerConn
}

func NewServer() *Server {
	return &Server{
		muxdem: NewMuxDemService(),
	}
}

func (s *Server) ListenAndServe(raw string) error {
	l, err := Listen(raw)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) Serve(listener net.Listener) error {
	// Translate a Go net.Listener to an RTMP net.Listener
	var concrete *Listener
	if l, ok := listener.(*Listener); !ok {
		l, err := newFromNetListener(listener)
		if err != nil {
			return fmt.Errorf("creating RTMP listener: %v", err)
		}
		concrete = l
	} else {
		concrete = l
	}
	s.listener = concrete
	urlAddr, err := NewURLAddr(s.listener.Addr().String())
	if err != nil {
		return fmt.Errorf("urlAddr: %v", err)
	}
	s.listener.addr = urlAddr
	logger.Info(rtmpMessage("server.Serve", serve))
	for {
		clientConn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("client conn accept: %v", err)
		}
		go func() {
			err := s.handleConn(clientConn)
			if err != nil {
				logger.Critical("dropped client: %v", err)
			}
		}()
	}
	return nil
}

func (s *Server) handleConn(netConn net.Conn) error {
	logger.Info(rtmpMessage(fmt.Sprintf("server.Accept client %s", netConn.RemoteAddr()), new))

	// Base connection
	conn := NewConn(netConn)

	// Server Connection
	connSrv := NewServerConn(conn)
	s.conn = connSrv
	s.conn.conn = conn

	// Set up multiplexing
	stream, err := s.muxdem.GetStream(s.listener.URLAddr().Key())
	if err != nil {
		return err
	}

	// Map the stream to the conn
	s.conn.stream = stream

	// Handshakes
	err = s.conn.handshake()
	if err != nil {
		return nil
	}

	return s.conn.RoutePackets()
}
