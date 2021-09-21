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

package twinx

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/kris-nova/logger"
)

const (
	TwinxActiveStreamPID string = "/var/run/twinx.pid"
)

type ActiveStream struct {
	PID   int
	PID64 int64
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
	return nil
}

// GetActiveStream will attempt to lookup an active stream running locally.
func GetActiveStream() (*ActiveStream, error) {
	return nil, nil
}

// NewActiveStream will create a new active stream as long as
// one does not exist.
func NewActiveStream() (*ActiveStream, error) {
	// Check if PID file exists
	if Exists(TwinxActiveStreamPID) {
		return nil, fmt.Errorf("existing PID File: %s", TwinxActiveStreamPID)
	}

	// A very poor fork() implementation
	_, err := ExecCommand("twinx", []string{"daemon", "&"})
	if err != nil {
		return nil, fmt.Errorf("unable to fork(): %v", err)
	}

	// Now we wait for the "daemon" to write it's PID
	started := false
	for i := 0; i < 10; i++ {
		if Exists(TwinxActiveStreamPID) {
			started = true
			break
		}
		time.Sleep(time.Second * 1)
	}
	if !started {
		return nil, fmt.Errorf("unable to find PID for stream")
	}
	pidBytes, err := ioutil.ReadFile(TwinxActiveStreamPID)
	if err != nil {
		return nil, fmt.Errorf("unable to access PID file: %v", err)
	}
	pidStr := string(pidBytes)
	logger.Info("Success. Found PID: %s", pidStr)
	pidInt := StrInt0(pidStr)
	if pidInt == 0 {
		return nil, fmt.Errorf("unable to parse PID from string: %v", err)
	}
	return &ActiveStream{
		PID:   pidInt,
		PID64: int64(pidInt),
	}, nil
}

// StopActiveStream will stop an active stream.
func StopActiveStream(x *ActiveStream) error {
	return nil
}

// KillActiveStream will force kill an active stream.
func KillActiveStream(x *ActiveStream) error {
	return nil
}
