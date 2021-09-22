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
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kris-nova/logger"
)

const (
	ActiveStreamPIDWriteMode os.FileMode = 0600
)

type Stream struct {
	Shutdown        chan bool
	IsManagedDaemon bool
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

	// Do not handle error. If it cannot be removed just exit and let the user
	// figure out what to do.
	defer os.Remove(ActiveStreamPID)
	pidInt := os.Getpid()
	pidStr := fmt.Sprintf("%d", pidInt)
	err := ioutil.WriteFile(ActiveStreamPID, []byte(pidStr), ActiveStreamPIDWriteMode)
	if err != nil {
		return fmt.Errorf("unable to write PID file: %v", err)
	}

	for {
		logger.Info("Streaming...")
		select {
		case <-s.Shutdown:
			logger.Always("Graceful shutdown...")
			return nil
		default:
			break
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
