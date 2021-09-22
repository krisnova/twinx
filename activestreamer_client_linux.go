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
// This is the "client".
// This is how the twinx command line tool interfaces with active streamers at runtime.
//

//go:build linux
// +build linux

package twinx

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/kris-nova/twinx/activestreamer"
	"google.golang.org/grpc"

	"github.com/kris-nova/logger"
)

const (
	ActiveStreamPID string = "/var/run/twinx.pid"
	ActiveStreamLog string = "/var/log/twinx.log"
)

type ActiveStream struct {
	Title       string
	Description string
	PID         int
	PID64       int64
	Client      *activestreamer.ActiveStreamerClient
}

// InfoChannel will return a channel that can be accessed
// to gain information about the stream.
func (x *ActiveStream) InfoChannel() chan string {
	ch := make(chan string)
	return ch
}

// Assure will run a sanity check against the active stream
// to assure that it is running, healthy, and that we can talk
// to it.
func (x *ActiveStream) Assure() error {

	// Note: See https://github.com/grpc/grpc-go/issues/1846
	// Example:
	//	passthrough:///unix:///run/example.sock

	conn, err := grpc.Dial(fmt.Sprintf("passthrough:///unix://%s", ActiveStreamSocket), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*3))
	if err != nil {
		return fmt.Errorf("error dialing socket: %v", err)
	}
	defer conn.Close()
	client := activestreamer.NewActiveStreamerClient(conn)
	x.Client = &client
	logger.Info("Successfully connected to server!")
	return nil
}

// GetActiveStream will attempt to lookup an active stream running locally.
func GetActiveStream() (*ActiveStream, error) {
	pidBytes, err := ioutil.ReadFile(ActiveStreamPID)
	if err != nil {
		return nil, fmt.Errorf("unable to access PID file: %v", err)
	}
	pidStr := string(pidBytes)
	//logger.Info("Success. Found PID: %s", pidStr)
	pidInt := StrInt0(pidStr)
	if pidInt == 0 {
		return nil, fmt.Errorf("unable to parse PID from string: %v", err)
	}
	return &ActiveStream{
		PID:   pidInt,
		PID64: int64(pidInt),
	}, nil
}

// NewActiveStream will create a new active stream as long as
// one does not exist.
func NewActiveStream(title, description string) (*ActiveStream, error) {
	// Check if PID file exists
	if Exists(ActiveStreamPID) {
		return nil, fmt.Errorf("existing PID File: %s", ActiveStreamPID)
	}

	// A very poor fork() implementation. Works for now.
	// Very important! When the parent process (this) exits, it WILL send SIGINT to the child.
	// It is up to the child to handle it accordingly. We handle this in twinx by checking ppid on the child!
	//
	// Command:
	// /bin/sh -c twinx activestreamer > /var/log/twinx.log &
	_, err := ExecCommand("/bin/sh", []string{"-c", fmt.Sprintf("twinx activestreamer > %s &", ActiveStreamLog)})
	if err != nil {
		return nil, fmt.Errorf("unable to fork(): %v", err)
	}
	//defer logger.Info(o.Stderr.String())
	//defer logger.Info(o.Stdout.String())

	// Now we wait for the "daemon" to write it's PID
	started := false
	for i := 0; i < 50; i++ {
		if Exists(ActiveStreamPID) {
			started = true
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	if !started {
		return nil, fmt.Errorf("unable to find PID for stream")
	}
	pidBytes, err := ioutil.ReadFile(ActiveStreamPID)
	if err != nil {
		return nil, fmt.Errorf("unable to access PID file: %v", err)
	}
	pidStr := string(pidBytes)
	//logger.Info("Success. Found PID: %s", pidStr)
	pidInt := StrInt0(pidStr)
	if pidInt == 0 {
		return nil, fmt.Errorf("unable to parse PID from string: %v", err)
	}

	return &ActiveStream{
		Title:       title,
		Description: description,
		PID:         pidInt,
		PID64:       int64(pidInt),
	}, nil
}

// StopActiveStream will stop an active stream.
func StopActiveStream(x *ActiveStream) error {
	if !Exists(ActiveStreamPID) {
		logger.Info("missing PID file")
		return nil
	}
	pidBytes, err := ioutil.ReadFile(ActiveStreamPID)
	if err != nil {
		return fmt.Errorf("unable to access PID file: %v", err)
	}
	pidStr := string(pidBytes)

	// Send SIGHUP (1)
	cmd, err := ExecCommand("kill", []string{"-1", pidStr})
	if err != nil {
		return fmt.Errorf("unable to kill: %v", err)
	}
	err = cmd.Command.Wait()
	if err != nil {
		return fmt.Errorf("error waiting on kill: %v", err)
	}
	return os.Remove(ActiveStreamPID)
}

// KillActiveStream will force kill an active stream.
func KillActiveStream(x *ActiveStream) error {
	if !Exists(ActiveStreamPID) {
		logger.Info("missing PID file")
		return nil
	}
	pidBytes, err := ioutil.ReadFile(ActiveStreamPID)
	if err != nil {
		return fmt.Errorf("unable to access PID file: %v", err)
	}
	pidStr := string(pidBytes)

	// Send SIGKILL (9)
	cmd, err := ExecCommand("kill", []string{"-9", pidStr})
	if err != nil {
		return fmt.Errorf("unable to kill: %v", err)
	}
	err = cmd.Command.Wait()
	if err != nil {
		return fmt.Errorf("error waiting on kill: %v", err)
	}
	return os.Remove(ActiveStreamPID)
}
