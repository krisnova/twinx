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

// ChunkStream
//
// 5.1.
//
// The format of a message that can be split into chunks to support
// multiplexing depends on a higher level protocol. The message format
// SHOULD however contain the following fields which are necessary for
// creating the chunks.
//
type ChunkStream struct {
	// Timestamp of the message. This field can transport 4
	// bytes.
	Timestamp uint32

	// Length of the message payload. If the message header cannot
	// be elided, it should be included in the length. This field
	// occupies 3 bytes in the chunk header.
	Length uint32

	// A range of type IDs are reserved for protocol control
	// messages. These messages which propagate information are handled
	// by both RTMP Chunk Stream protocol and the higher-level protocol.
	// All other type IDs are available for use by the higher-level
	// protocol, and treated as opaque values by RTMP Chunk Stream. In
	// fact, nothing in RTMP Chunk Stream requires these values to be
	// used as a type; all (non-protocol) messages could be of the same
	// type, or the application could use this field to distinguish
	// simultaneous tracks rather than types. This field occupies 1 byte
	// in the chunk header.
	TypeID uint32

	// The message stream ID can be any arbitrary value.
	// Different message streams multiplexed onto the same chunk stream
	// are demultiplexed based on their message stream IDs. Beyond that,
	// as far as RTMP Chunk Stream is concerned, this is an opaque value.
	// This field occupies 4 bytes in the chunk header in little endian
	// format.
	StreamID uint32

	// Data is the set of bytes in the Chunk. The chunk payload.
	Data []byte

	Format    uint32
	CSID      uint32
	timeDelta uint32
	exited    bool
	index     uint32
	remain    uint32
	got       bool
	tmpFormat uint32
}
