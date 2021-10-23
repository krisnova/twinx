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
	listener *Listener

	// *** Clients ***
	//
	// We have 3 types of clients for the server. All clients
	// need to be one of these or the other. (proxy, play, publish)
	//
	//   [ Play] <--- (1234) --- [ Server ]
	// [ Publish ] -- (1234) --> [ Server ]
	// [ Publish ] -- (1234) --> [ Server ] -- (5678) --> [ Proxy Publish ]

	// playClients are clients connected to the server, that have been registered
	// as play clients
	playClients map[string]*ServerConn

	// publishClients are clients connected to the server, that have been registered
	// as publish clients
	publishClients map[string]*ServerConn

	// proxyPublishClients are clients connected to the server that will be used
	// to proxy the RTMP as a new publish client on a remote backend.
	//
	// These are known as "push" clients in the Nginx module.
	proxyPublishClients map[string]*ClientConn
}

func NewServer() *Server {
	return &Server{
		proxyPublishClients: make(map[string]*ClientConn),
		playClients:         make(map[string]*ServerConn),
		publishClients:      make(map[string]*ServerConn),
	}
}

// Proxy will configure forward addresses for the RTMP server.
//
// Proxy can be called before or after Serve()
// and the backend server will be smart enough to sync clients.
func (s *Server) Proxy(raw string) error {
	forwardClient := NewClient()
	err := forwardClient.Dial(raw)
	if err != nil {
		return err
	}
	return s.ProxyClient(forwardClient.conn)
}

// ProxyClient will add clients to this server.
//
// Client forwarding is handled at the server level.
// We trust each subsequent stream to update to the configured
// clients as they are added.
func (s *Server) ProxyClient(f *ClientConn) error {

	// New clients will always be publishers.
	go func() {
		// Set Default OBS for testing
		//logger.Warning("DEBUG sending VirtualOBSMetaData")
		//f.virtualMetaData = VirtualOBSOutputClientMetadata()

		err := f.Publish()
		if err != nil {
			logger.Critical(err.Error())
		}
	}()

	logger.Info(rtmpMessage(fmt.Sprintf("server.AddClient(%s)", f.urladdr.SafeURL()), ack))
	s.proxyPublishClients[f.urladdr.SafeURL()] = f
	return nil
}

func (s *Server) PublishClient(f *ServerConn) {
	s.publishClients[s.listener.URLAddr().SafeURL()] = f
}

func (s *Server) PlayClient(f *ServerConn) {
	s.playClients[s.listener.URLAddr().SafeURL()] = f
}

func (s *Server) ListenAndServe(raw string) error {
	l, err := Listen(raw)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

// Serve
//
// A blocking method that will listen for new connections
// and create subsequent go routines for new clients as they
// connect.
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
	logger.Info(rtmpMessage("server.Serve", serve))

	// At this point we should have a full RTMP listener, with a
	// complete Key.

	// Metrics Point
	M().Lock()
	M().ServerAddrRX = concrete.URLAddr().SafeURL()
	M().ServerKeyHash = concrete.URLAddr().SafeKey()
	M().Unlock()

	for {
		clientConn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("client conn accept: %v", err)
		}
		go func() {
			err := s.handleConn(clientConn, concrete.URLAddr())
			if err != nil {
				logger.Critical("dropped client: %v", err)
			}
		}()
	}
	return nil
}

func (s *Server) handleConn(netConn net.Conn, urladdr *URLAddr) error {
	logger.Info(rtmpMessage(fmt.Sprintf("server.Accept client %s", netConn.RemoteAddr()), new))

	// Base connection
	conn := NewConn(netConn)
	conn.URLAddr = *urladdr

	// Server Connection
	// This is a bit weird naming, but each of these connections
	// are the accepted client to the server.
	client := NewServerConn(conn)
	client.conn = conn

	// Point all clients back to the main server
	client.server = s

	// Handshakes
	var err error
	err = client.handshake()
	if err != nil {
		return nil
	}

	go func() {
		err = client.RoutePackets()
		if err != nil {
			logger.Critical(err.Error())
			client.Close()
		}
	}()
	for {
		if client.clientType == PlayClient {
			s.PlayClient(client)
			return nil
		}
		if client.clientType == PublishClient {
			s.PublishClient(client)
			return nil
		}
	}

}
