// Copyright © 2021 Kris Nóva <kris@nivenly.com>
// Copyright (c) 2017 吴浩麟
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
	"sync"

	"github.com/kris-nova/logger"
)

// SafeMuxDemuxService (Safe Multiplexer/Demultiplexer Service)
//
// SafeMuxDemuxService is the main data structure that allows thread-safe packet
// multiplexing and routing at runtime.
//
// The multiplexer is what enables a compliant multiplexed RTMP transaction.
// Furthermore, this multiplexer allows the same process to transfer packets
// between streams.
type SafeMuxDemuxService struct {
	mux *sync.Map
}

const (
	// DefaultMaximumBufferSizeChunkStream
	//
	// In order to safely multiplex we build a FIFO
	// queue in memory that will be multiplexed onto
	// any configured readers and writers.
	//
	// As new packets are received, we need to specify
	// a maximum number of packets before the buffer
	// can no longer accept packets before draining.
	DefaultMaximumBufferSizeChunkStream int = 1024 * 1024
)

func NewMuxDemService() *SafeMuxDemuxService {
	return &SafeMuxDemuxService{
		mux: &sync.Map{},
	}
}

func (s *SafeMuxDemuxService) GetStream(key string) (*SafeBoundedBuffer, error) {
	var strm *SafeBoundedBuffer
	v, ok := s.mux.Load(key)
	if !ok {
		// New buffer
		//
		// A note on buffer size. A total buffer memory footprint can be measured:
		// (Chunk Meta + Current Chunk Size) * Queue Size = Total memory in bytes
		logger.Debug(rtmpMessage("New Stream Buffer", stream))
		fmt.Println(key)
		strm = NewSafeBoundedBuffer(key, DefaultMaximumBufferSizeChunkStream)
		s.mux.Store(key, stream)
		return strm, nil
	}
	if strm, ok := v.(*SafeBoundedBuffer); ok {
		return strm, nil
	}
	return nil, errors.New("unknown buffer type")
}

type SafeBoundedBuffer struct {
	name string

	writers map[string]ChunkStreamWriter

	writeMutex sync.Mutex

	upperBufferLimit int

	// packetBuffer is a FIFO queue
	packetBuffer []*ChunkStream

	droppedPackets int
	writtenPackets int
}

func NewSafeBoundedBuffer(name string, upperBufferLimit int) *SafeBoundedBuffer {
	return &SafeBoundedBuffer{
		name:             name,
		upperBufferLimit: upperBufferLimit,
		writeMutex:       sync.Mutex{},
		writers:          make(map[string]ChunkStreamWriter),
	}
}

// Write
//
// All packets can be written.
//
// Write is designed to drop any packets that cannot be managed instead
// of returning an error.
func (mx *SafeBoundedBuffer) Write(x *ChunkStream) {
	//fmt.Println("write to buffer: ", mx.name)

	mx.writeMutex.Lock()
	defer mx.writeMutex.Unlock()
	if len(mx.packetBuffer) >= mx.upperBufferLimit {
		// Drop the packet.
		mx.droppedPackets++
		logger.Debug("Dropping Audio/Video packet... Buffer overflow...")
		return
	}
	// Add the packet to the end of the queue
	mx.packetBuffer = append(mx.packetBuffer, x)
	//os.Exit(1)
}

func (mx *SafeBoundedBuffer) Stream() error {

	// Get FIFO chunk
	logger.Debug(rtmpMessage(fmt.Sprintf("Streaming [%s]", mx.name), stream))
	var x *ChunkStream
	for {
		mx.writeMutex.Lock()
		if len(mx.packetBuffer) > 0 {

			// Find the top element (x)
			x = mx.packetBuffer[0]

			// Process (x) for every configured output
			for _, w := range mx.writers {
				err := w.Write(x)
				if err != nil {
					// Unable to proxy
					logger.Critical(err.Error())
				}
			}

			// Drop (x) from the queue
			mx.packetBuffer = mx.packetBuffer[1:]
		}
		mx.writeMutex.Unlock()
	}
}

func (mx *SafeBoundedBuffer) AddWriter(key string, w ChunkStreamWriter) {
	mx.writeMutex.Lock()
	logger.Debug(rtmpMessage(fmt.Sprintf("AddWriter [%s]", mx.name), stream))
	mx.writers[key] = w
	defer mx.writeMutex.Unlock()
}

type ChunkStreamWriter interface {
	// Write (by design) does not block
	Write(x *ChunkStream) error
}
