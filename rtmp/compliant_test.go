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
	"os"
	"testing"
	"time"

	"github.com/kris-nova/logger"
)

const (
	TestClientAddr string = "localhost:1936/twinx/12345"
	TestServerAddr string = "localhost:1936/twinx/12345"
)

func TestMain(m *testing.M) {
	logger.BitwiseLevel = logger.LogEverything
	server := NewServer()
	go func() {
		err := server.ListenAndServe(TestServerAddr)
		if err != nil {
			logger.Critical("unable to start server")
			os.Exit(1)
		}
	}()
	os.Exit(m.Run())
}

func TestClientPlay(t *testing.T) {
	time.Sleep(time.Millisecond * 125)
	client := NewClient()
	err := client.Dial(TestClientAddr)
	if err != nil {
		t.Errorf("unable to dial client: %v", err)
		t.FailNow()
	}
	go func() {
		err := client.Play()
		if err != nil {
			t.Errorf("play error: %v", err)
		}
	}()
	defer client.conn.Close()
	time.Sleep(time.Millisecond * 125)

}

func TestClientPublish(t *testing.T) {
	time.Sleep(time.Millisecond * 125)

	client := NewClient()
	err := client.Dial(TestClientAddr)
	if err != nil {
		t.Errorf("unable to dial client: %v", err)
		t.FailNow()
	}
	go func() {
		err := client.Publish()
		if err != nil {
			t.Errorf("publish error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 125)
	defer client.conn.Close()

}

func TestClientPlayClientPublish(t *testing.T) {
	time.Sleep(time.Millisecond * 125)

	client := NewClient()
	err := client.Dial(TestClientAddr)
	if err != nil {
		t.Errorf("unable to dial client: %v", err)
		t.FailNow()
	}
	go func() {
		err := client.Publish()
		if err != nil {
			t.Errorf("play error: %v", err)
		}
	}()
	defer client.conn.Close()

	clientPlay := NewClient()
	err = clientPlay.Dial(TestClientAddr)
	if err != nil {
		t.Errorf("unable to dial client: %v", err)
		t.FailNow()
	}
	go func() {
		err := clientPlay.Publish()
		if err != nil {
			t.Errorf("play error: %v", err)
		}
	}()
	defer client.conn.Close()
	time.Sleep(time.Millisecond * 125)

}
