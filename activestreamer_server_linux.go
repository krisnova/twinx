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

//
// This is the "server".
// This is the code that an active streamer runs.

//go:build linux
// +build linux

package twinx

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gwuhaolin/livego/av"

	"github.com/sirupsen/logrus"

	"github.com/gwuhaolin/livego/protocol/rtmp/rtmprelay"

	"github.com/gwuhaolin/livego/protocol/rtmp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kris-nova/twinx/activestreamer"

	"google.golang.org/grpc"

	"github.com/kris-nova/logger"
)

const (
	ActiveStreamPIDWriteMode os.FileMode = 0600
	ActiveStreamSocket                   = "/var/run/twinx.sock"
	ActiveStreamRTMPHost                 = "localhost"
)

type Stream struct {
	Shutdown        chan bool
	IsManagedDaemon bool
	Server          *grpc.Server
	LogrusBuffer    *bytes.Buffer
}

func NewStream() *Stream {
	return &Stream{
		Shutdown: make(chan bool, 1),
	}
}

// Run will run the stream until a client tells it to stop.
func (s *Stream) Run() error {
	if Exists(ActiveStreamPID) {
		return fmt.Errorf("existing PID file %s", ActiveStreamPID)
	}

	// Setup the signal handler in Run()
	s.SigHandler()

	// Setup the gRPC server
	go func() {
		err := s.ServerGRPC()
		if err != nil {
			logger.Critical("Unable to start gRPC server! %v", err)
			s.Shutdown <- true
		}
	}()

	// Setup the log buffer for logrus
	s.LogrusBuffer = &bytes.Buffer{}
	logger.Info("Logrus level: TraceLevel")
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(s.LogrusBuffer)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableColors:          true,
		DisableLevelTruncation: false,
		//CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		//	filename := path.Base(f.File)
		//	return fmt.Sprintf("Buffered Logrus Log: %s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
		//},
	})
	logrus.Info("Ensure level: TraceLevel")
	defer logrus.Exit(0)

	// Do not handle error. If it cannot be removed just exit and let the user
	// figure out what to do.
	defer os.Remove(ActiveStreamPID)
	pidInt := os.Getpid()
	pidStr := fmt.Sprintf("%d", pidInt)
	err := ioutil.WriteFile(ActiveStreamPID, []byte(pidStr), ActiveStreamPIDWriteMode)
	if err != nil {
		return fmt.Errorf("unable to write PID file: %v", err)
	}

	logger.Info("Streaming...")
	if s.Server != nil {
		info := s.Server.GetServiceInfo()
		for name, service := range info {
			logger.Info("%s %v", name, service.Metadata)
		}
	}
	for {
		select {
		case <-s.Shutdown:
			s.Server.GracefulStop()
			os.Remove(ActiveStreamSocket)
			os.Remove(ActiveStreamPID)
			logger.Always("Graceful shutdown...")
			return nil
		default:
			break
		}

		flushedLogsRaw, err := ioutil.ReadAll(s.LogrusBuffer)
		if err != nil {
			logger.Critical("logrus buffer read: %v", err)
		}
		if len(flushedLogsRaw) > 0 {
			logger.Warning(string(flushedLogsRaw))
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}

func (s *Stream) SigHandler() {
	sigCh := make(chan os.Signal, 2)

	// Register signals for the signal handler
	// os.Interrupt is ^C
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, os.Interrupt)
	go func() {
		sig := <-sigCh
		logger.Always("Shutting down...")
		switch sig {
		case syscall.SIGHUP:
			logger.Always("SIGHUP")
			s.Shutdown <- true
		case syscall.SIGINT:
			logger.Always("SIGINT")
			// Check parent PID
			ppid := os.Getppid()
			logger.Always("%d", ppid)

			// ppid == 1 The daemon was started by root in true daemon mode
			// ppid != 1 The deamon was started in foreground mode
			if ppid != 1 {
				s.Shutdown <- true
			} else {
				logger.Always("Daemon launched successfully!")
				s.IsManagedDaemon = true
			}
		case syscall.SIGTERM:
			logger.Always("SIGTERM")
			s.Shutdown <- true
		case syscall.SIGKILL:
			logger.Always("SIGKILL")
			s.Shutdown <- true
		case syscall.SIGQUIT:
			logger.Always("SIGQUIT")
			s.Shutdown <- true
		default:
			logger.Always("os.Interrupt() DEFAULT")
			logger.Always("Caught Signal!")
			s.Shutdown <- true
		}
	}()
}

func (s *Stream) ServerGRPC() error {
	if Exists(ActiveStreamSocket) {
		return fmt.Errorf("grpc stream socket exists %s", ActiveStreamSocket)
	}

	conn, err := net.Listen("unix", ActiveStreamSocket)
	if err != nil {
		return fmt.Errorf("unable to open unix domain socket: %v", err)
	}
	server := grpc.NewServer()
	activestreamer.RegisterActiveStreamerServer(server, NewActiveStreamerServer())
	//log.Printf("server listening at %v", lis.Addr())
	logger.Info("ActiveStreamer listening: %v", conn.Addr())
	s.Server = server
	if err := server.Serve(conn); err != nil {
		return fmt.Errorf("unable to start server on unix domain socket: %v", err)
	}
	return nil
}

func S(s string) *string {
	return SPointer(s)
}

func SPointer(s string) *string {
	return &s
}

type ActiveStreamerServer struct {
	activestreamer.UnimplementedActiveStreamerServer
	Local    *activestreamer.RTMPHost
	Remotes  map[string]*activestreamer.RTMPHost
	Handler  av.Handler
	Listener net.Listener
}

func NewActiveStreamerServer() *ActiveStreamerServer {
	return &ActiveStreamerServer{
		Remotes: make(map[string]*activestreamer.RTMPHost),
	}
}

func (a *ActiveStreamerServer) RTMPStartRelay(ctx context.Context, r *activestreamer.RTMPHost) (*activestreamer.Ack, error) {

	addr, err := RTMPNewAddr(r.Addr)
	if err != nil {
		return &activestreamer.Ack{
			Success: false,
			Message: S("invalid RTMP addr"),
		}, fmt.Errorf("invalid RTPM addr: %v", err)
	}

	logger.Debug("Starting RTMP Relay Addr       %s", r.Addr)
	logger.Debug("Starting RTMP Relay BufferSize %d", r.BufferSize)

	// Ensure no host has been started
	if a.Local != nil {
		return &activestreamer.Ack{
			Success: false,
			Message: S("unable to start rtmp, already running"),
		}, fmt.Errorf("unable to start rtmp, already running")
	}

	// RTMPStream is a set of rtmp.Stream
	stream := rtmp.NewRtmpStream()
	server := rtmp.NewRtmpServer(stream, nil)

	logger.Debug("net.listen TCP %s", addr.Server())
	listener, err := net.Listen(RTMPProtocol, addr.Server())
	if err != nil {
		return &activestreamer.Ack{
			Success: false,
			Message: S(err.Error()),
		}, err
	}

	// Cache the local server
	logger.Debug("Caching local RTMP server")
	a.Local = r
	a.Listener = listener

	// Run the server in a go routine
	go func() {
		logger.Info("Starting local RTMP server %s", r.Addr)
		err = server.Serve(listener)
		if err != nil {
			logger.Critical(err.Error())
		}
	}()

	// This is called in New()
	// go handler.CheckAlive()
	a.Handler = stream

	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
func (a *ActiveStreamerServer) RTMPStopRelay(context.Context, *activestreamer.Null) (*activestreamer.Ack, error) {

	// Ensure no host has been started
	if a.Local == nil {
		return &activestreamer.Ack{
			Success: false,
			Message: S("unable to stop rtmp, not running"),
		}, nil
	}

	err := a.Listener.Close()
	if err != nil {
		return &activestreamer.Ack{
			Success: false,
			Message: S(fmt.Sprintf("closing RTMP server: %v", err)),
		}, err
	}
	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
func (a *ActiveStreamerServer) RTMPForward(ctx context.Context, r *activestreamer.RTMPHost) (*activestreamer.Ack, error) {

	addr, err := RTMPNewAddr(r.Addr)
	if err != nil {
		return &activestreamer.Ack{
			Success: false,
			Message: S("invalid RTMP addr"),
		}, fmt.Errorf("invalid RTPM addr: %v", err)
	}

	//logger.Debug("Starting RTMP Relay Forward Addr       %s", r.Addr)
	//logger.Debug("Starting RTMP Relay Forward BufferSize %d", r.BufferSize)

	// Ensure no host has been started
	if a.Local == nil {
		return &activestreamer.Ack{
			Success: false,
			Message: S("unable to start rtmp relay, local server not running"),
		}, fmt.Errorf("unable to start rtmp relay, local server notrunning")
	}

	logger.Debug("Starting RTMP relay %s -> %s", a.Local.Addr, r.Addr)
	relay := rtmprelay.NewRtmpRelay(S(a.Local.Addr), S(addr.Full()))

	// Cache
	a.Remotes[r.Addr] = r

	go func() {
		err := relay.Start()
		if err != nil {

			logger.Critical("Error forwarding RTMP. Raw: %v", err)
			logger.Critical("Check the forward address.")

			// Note: The backend library is written by what I assume is an ESL engineer.
			// The "u path err:" message is here: https://github.com/gwuhaolin/livego/blob/master/protocol/rtmp/core/conn_client.go#L220
			//
			// u, err := neturl.Parse(url)
			//	if err != nil {
			//		return err
			//	}
			//	connClient.url = url
			//	path := strings.TrimLeft(u.Path, "/")
			//	ps := strings.SplitN(path, "/", 2)
			//	if len(ps) != 2 {
			//		return fmt.Errorf("u path err: %s", path)
			//	}
			//
			if strings.Contains(err.Error(), "u path err:") {
				logger.Critical("Error with backend RTMP library")
				logger.Critical("  Configured PlayURL   : %s", relay.PlayUrl)
				logger.Critical("  Configured PublishURL: %s", relay.PublishUrl)
				logger.Critical("All URLs must contain starting protocols such as rtmp:// or http:// prefixes")
			}

		}
	}()

	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
func (a *ActiveStreamerServer) SetTwitchMeta(context.Context, *activestreamer.StreamMeta) (*activestreamer.Ack, error) {
	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
func (a *ActiveStreamerServer) SetYouTubeMeta(context.Context, *activestreamer.StreamMeta) (*activestreamer.Ack, error) {
	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
func (a *ActiveStreamerServer) GetStreamMeta(context.Context, *activestreamer.ClientConfig) (*activestreamer.StreamMeta, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStreamMeta not implemented")
}
func (a *ActiveStreamerServer) SetStreamMeta(context.Context, *activestreamer.StreamMeta) (*activestreamer.Ack, error) {
	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
func (a *ActiveStreamerServer) Transact(context.Context, *activestreamer.ClientConfig) (*activestreamer.Ack, error) {
	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
func (a *ActiveStreamerServer) SetLogger(context.Context, *activestreamer.LoggerConfig) (*activestreamer.Ack, error) {
	return &activestreamer.Ack{
		Success: true,
		Message: S("Success"),
	}, nil
}
