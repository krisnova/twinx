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

// CompliantMember is a member that validates that both servers and clients
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
type CompliantMember interface {

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
	connectTX() (*ChunkStream, error)

	// 7.2.1.3 createStream
	createStreamRX(x *ChunkStream) error
	createStreamTX() (*ChunkStream, error)

	// === 7.2.2 NetStream Commands ===

	// 7.2.2.1 play
	playRX(x *ChunkStream) error
	playTX() (*ChunkStream, error)

	// 7.2.2.2 play2
	play2RX(x *ChunkStream) error
	play2TX() (*ChunkStream, error)

	// 7.2.2.3 deleteStream
	deleteStreamRX(x *ChunkStream) error
	deleteStreamTX() (*ChunkStream, error)

	// 7.2.2.4 receiveAudio
	receiveAudioRX(x *ChunkStream) error
	receiveAudioTX() (*ChunkStream, error)

	// 7.2.2.5 receiveVideo
	receiveVideoRX(x *ChunkStream) error
	receiveVideoTX() (*ChunkStream, error)

	// 7.2.2.6 publish
	publishRX(x *ChunkStream) error
	publishTX() (*ChunkStream, error)

	// 7.2.2.7 seek
	seekRX(x *ChunkStream) error
	seekTX() (*ChunkStream, error)

	// 7.2.2.8 pause
	pauseRX(x *ChunkStream) error
	pauseTX() (*ChunkStream, error)

	// --------------------------------------------------------------------
	//
	// [OOS] Out of Spec
	// Below this marker, there are several (observed) methods that are not
	// found in the official RTMP spec.
	//
	// --------------------------------------------------------------------

	//2021-10-15T10:41:44-07:00 [Debug     ]    [0] (FCPublish)
	//2021-10-15T10:41:44-07:00 [Debug     ]    [1] (3)
	//2021-10-15T10:41:44-07:00 [Debug     ]    [2] (<nil>)
	//2021-10-15T10:41:44-07:00 [Debug     ]    [3] (live_733531528_k9ZMBZXSUfOuGCrquQbgeXmLa5Y5ve)
	oosFCPublishRX(x *ChunkStream) error
	oosFCPublishTX() (*ChunkStream, error)

	//2021-10-15T10:41:44-07:00 [Debug     ]    [0] (releaseStream)
	//2021-10-15T10:41:44-07:00 [Debug     ]    [1] (2)
	//2021-10-15T10:41:44-07:00 [Debug     ]    [2] (<nil>)
	//2021-10-15T10:41:44-07:00 [Debug     ]    [3] (live_733531528_k9ZMBZXSUfOuGCrquQbgeXmLa5Y5ve)
	oosReleaseStreamRX(x *ChunkStream) error
	oosReleaseStreamTX() (*ChunkStream, error)
}

type ChunkStreamRouter interface {
	// RoutePackets just reads packets and calls Route()
	RoutePackets() error

	Route(x *ChunkStream) error
}

type ChunkStreamReader interface {
	Read(x *ChunkStream) error
	NextChunk() (*ChunkStream, error)
}

// Enforce the implementation at compile time
var (
	_ CompliantMember = &ServerConn{}
	_ CompliantMember = &ClientConn{}
)
