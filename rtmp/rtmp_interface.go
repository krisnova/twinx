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
	connectRX() error
	connectTX() error

	// 7.2.1.3 createStream
	createStreamRX() error
	createStreamTX() error

	// === 7.2.2 NetStream Commands ===

	// 7.2.2.1 play
	playRX() error
	playTX() error

	// 7.2.2.2 play2
	play2RX() error
	play2TX() error

	// 7.2.2.3 deleteStream
	deleteStreamRX() error
	deleteStreamTX() error

	// 7.2.2.4 receiveAudio
	receiveAudioRX() error
	receiveAudioTX() error

	// 7.2.2.5 receiveVideo
	receiveVideoRX() error
	receiveVideoTX() error

	// 7.2.2.6 publish
	publishRX() error
	publishTX() error

	// 7.2.2.7 seek
	seekRX() error
	seekTX() error

	// 7.2.2.8 pause
	pauseRX() error
	pauseTX() error
}

// Enforce the implementation at compile time
var (
//_ CompliantRTMPMember = NewServer()
//_ CompliantRTMPMember = NewClient()
)
