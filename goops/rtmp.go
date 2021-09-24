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

package goops

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kris-nova/logger"
)

const (

	// RTMPProtocol is a well-known TCP protocol.
	RTMPProtocol string = "tcp"

	// BufferSizeOBSDefaultBytes is the default output buffer size used by OBS.
	// This should be used for the simplest and smoothest use with OBS.
	// This can be adjusted (and so should OBS!) if you are sure what you
	// are doing, and have system resources to support your change.
	BufferSizeOBSDefaultBytes int64 = 2500

	// BufferSizeNovaDefaultBytes is my personal default buffer size for my
	// streams. I run Arch btw.
	BufferSizeNovaDefaultBytes int64 = 256
)

type Service struct {
	listenHost   string
	listenPort64 int64
	listenPort   int
	listener     net.Listener
	proxyMutex   sync.Mutex
	proxies      []*Proxy
	bufferSize   int
}

type Proxy struct {
	conn net.Conn
}

func NewService(host string, port int, bufferSize int) *Service {
	return &Service{
		listenHost:   host,
		listenPort:   port,
		listenPort64: int64(port),
		proxies:      []*Proxy{},
		bufferSize:   bufferSize,
	}
}

func (g *Service) Listen() {
	//logger.BitwiseLevel = logger.LogEverything
	listener, err := net.Listen(RTMPProtocol, g.ListenAddr())
	if err != nil {
		logger.Critical("unable to start RTMP server: %v", err)
		time.Sleep(time.Second * 2)
		logger.Info("Restarting RTMP server recursively...")
		g.Listen()
	}
	logger.Info("Started RTMP server [%s:%d] BufferSize %d bytes", g.listenHost, g.listenPort, g.bufferSize)
	g.listener = listener
	defer g.listener.Close()
	for {
		conn, err := g.listener.Accept()
		if err != nil {
			logger.Warning("unable to accept new connection: %v", err)
			continue
		}
		logger.Info("*****************************************************")
		logger.Info("client connected: %s", conn.RemoteAddr().String())
		logger.Info("*****************************************************")
		go g.manageConn(conn)
	}
}

var (
	connectedCount int = 0
)

// manageConn is the internal concurrent system that we launch for every
// connection.
//
// This is the system that will route the bytes for our configured proxies.
// This system will also respect the associated mutex with the proxies.
func (g *Service) manageConn(conn net.Conn) {
	connectedCount++
	logger.Debug("total managed connections: %d", connectedCount)
	logger.Debug("managing local connection: %s", conn.LocalAddr().String())
	logger.Debug("buffer size: %d bytes", g.bufferSize)
	buffer := make([]byte, g.bufferSize)
	for {

		// Read the bytes into the configured buffer.
		// If there are no proxies configured, then the buffer
		// will just rewrite over itself.
		//
		// It is important to note we deliberately ignore errors
		// here. Our intention is for this connection to favor speed
		// over resiliency.
		conn.Read(buffer)
		g.proxyMutex.Lock()
		for _, proxy := range g.proxies {
			proxy.conn.Write(buffer)
		}
		g.proxyMutex.Unlock()
	}
}

// AddProxyDestination will attempt to add a proxy destination as seemlessly as possible.
func (g *Service) AddProxyDestination(host string, port int) error {
	g.proxyMutex.Lock()
	defer g.proxyMutex.Unlock()
	conn, err := net.Dial(RTMPProtocol, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("unable to connect to foreign RTMP server: %s", err)
	}
	logger.Info("connected to foreign proxy server: %s:%d", host, port)

	g.proxies = append(g.proxies, &Proxy{
		conn: conn,
	})
	return nil
}

func (g *Service) ListenAddr() string {
	return fmt.Sprintf("%s:%d", g.listenHost, g.listenPort)
}
