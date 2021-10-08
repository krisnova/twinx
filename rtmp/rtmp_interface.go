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
	"errors"
	"fmt"
	"runtime"
)

// CompliantRTMPMember is a member that validates that both servers and clients
// have associated methods to respond to each other.
//
// This allows a clearer understanding of communication between the entities and
// ensures that we always offer support for both server and client functionality.
//
// This is this packages interpretation of the RTMP spec, and our guarantees about
// our server and client code.
//
// We us "RX" and "TX" to identify the direction of packets.
//    RX = Receive  : Packets coming from a remote to this process
//    TX = Transmit : Packets sending to a remote from this process
type CompliantRTMPMember interface {

	// 5.2 Handshake
	//
	// handshake must be implemented by both
	// client and server code and should span
	// the entirety of the handshake for each member
	// or return an error
	//
	// on success the handshake is complete
	handshake() error

	// === 7.2.1 NetConnection Commands ===

	// 7.2.1.1. connect
	connectRX(x *ChunkStream) error
	connectTX(x *ChunkStream) error

	// 7.2.1.3 createStream
	createStreamRX(x *ChunkStream) error
	createStreamTX(x *ChunkStream) error

	// === 7.2.2 NetStream Commands ===

	// 7.2.2.1 play
	playRX(x *ChunkStream) error
	playTX(x *ChunkStream) error

	// 7.2.2.2 play2
	play2RX(x *ChunkStream) error
	play2TX(x *ChunkStream) error

	// 7.2.2.3 deleteStream
	deleteStreamRX(x *ChunkStream) error
	deleteStreamTX(x *ChunkStream) error

	// 7.2.2.4 receiveAudio
	receiveAudioRX(x *ChunkStream) error
	receiveAudioTX(x *ChunkStream) error

	// 7.2.2.5 receiveVideo
	receiveVideoRX(x *ChunkStream) error
	receiveVideoTX(x *ChunkStream) error

	// 7.2.2.6 publish
	publishRX(x *ChunkStream) error
	publishTX(x *ChunkStream) error

	// 7.2.2.7 seek
	seekRX(x *ChunkStream) error
	seekTX(x *ChunkStream) error

	// 7.2.2.8 pause
	pauseRX(x *ChunkStream) error
	pauseTX(x *ChunkStream) error
}

// Enforce the implementation at compile time
var (
	_ CompliantRTMPMember = &ServerConn{}
	_ CompliantRTMPMember = &ClientConn{}

	DefaultUnimplementedError = errors.New("**UNIMPLEMENTED**")
)

func defaultUnimplemented() error {
	pc := make([]uintptr, 1)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return DefaultUnimplementedError
	}
	caller := runtime.FuncForPC(pc[0] - 1)
	if caller == nil {
		return DefaultUnimplementedError
	}
	return fmt.Errorf("function %s %v", caller.Name(), DefaultUnimplementedError)
}
