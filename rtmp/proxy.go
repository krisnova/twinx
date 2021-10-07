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
	"io"

	"github.com/kris-nova/logger"
)

const (
	StopProxy InternalProxyControlMessage = iota
)

type InternalProxyControlMessage int

// RTMPProxy is a RTMP proxy (relay) that can connect
// multiple (multiplexed) streams.
//
// Will proxy a publish client stream as a play client
// stream to a configured endpoint.
type RTMPProxy struct {
	playClient    *ConnClient
	playAddr      *URLAddr
	publishClient *ConnClient
	publishAddr   *URLAddr

	chunkStreamCh  chan ChunkStream
	proxyMessageCh chan InternalProxyControlMessage
	isStreaming    bool
}

func NewRTMPProxy(play, publish *ConnClient) *RTMPProxy {
	return &RTMPProxy{
		playClient:     play,
		publishClient:  publish,
		chunkStreamCh:  make(chan ChunkStream, DefaultRTMPChunkSizeBytes*500),
		proxyMessageCh: make(chan InternalProxyControlMessage),
		isStreaming:    false,
	}
}

// rxChunkStreamPlay
//
// This will read bytes from a play client and send them
// over a packet channel (ChunkStream) in an unbounded buffer.
func (r *RTMPProxy) rxChunkStreamPlay() error {
	for {
		var x ChunkStream

		// Allow internal control outside of our
		// stack pointer's scope.
		if r.isStreaming == false {
			r.playClient.Close()
			break
		}
		err := r.playClient.Read(&x)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading packet from play(%s) client: %v", r.playAddr.SafeURL(), err)
		}

		switch x.TypeID {
		case SetChunkSizeMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case AbortMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case AcknowledgementMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case WindowAcknowledgementSizeMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case SetPeerBandwidthMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case UserControlMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case CommandMessageAMF0ID, CommandMessageAMF3ID:
			//xReader := bytes.NewReader(x.Data)
			//vs, err := r.playClient.DecodeBatch(xReader, amf.AMF0)
			//if err != nil && err != io.EOF {
			//	return fmt.Errorf("decoding bytes from play(%s) client: %v", r.playAddr.SafeURL(), err)
			//}
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case DataMessageAMF0ID, DataMessageAMF3ID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case SharedObjectMessageAMF0ID, SharedObjectMessageAMF3ID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		case AudioMessageID:
			// Send audio over our channel
			r.chunkStreamCh <- x
		case VideoMessageID:
			// Send video over our channel
			r.chunkStreamCh <- x
		case AggregateMessageID:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		default:
			logger.Critical("unsupported messageID: %s", typeIDString(&x))
		}
	}
	return nil
}

// txChunkStreamPublish will read from the packet (ChunkStream)
// channel and write to the configured client.
func (r *RTMPProxy) txChunkStreamPublish() {
	for {
		select {
		case rc := <-r.chunkStreamCh:
			r.publishClient.Write(rc)
		case ctl := <-r.proxyMessageCh:
			if ctl == StopProxy {
				r.publishClient.Close()
				return
			}
		}
	}
}

// Start will start proxying packets from one client to another
func (r *RTMPProxy) Start() error {
	if r.publishClient == nil {
		return errors.New("missing publish client")
	}
	if r.playClient == nil {
		return errors.New("missing play client")
	}
	if r.isStreaming {
		return errors.New("proxy already streaming")
	}

	// [ Play Client Connection ]
	err := r.playClient.Play()
	defer r.playClient.Close()
	if err != nil {
		return fmt.Errorf("play(%s) --> [twinx proxy]: %v", r.playAddr.SafeURL(), err)
	}

	// [ Publish Client Connection ]
	err = r.publishClient.Publish()
	defer r.publishClient.Close()
	if err != nil {
		return fmt.Errorf("[twinx proxy] --> publish(%s): %v", r.playAddr.SafeURL(), err)
	}

	r.isStreaming = true

	// **************************
	// Proxy Logic
	//
	go r.rxChunkStreamPlay()
	go r.txChunkStreamPublish()
	// **************************

	logger.Info("Proxy started.")
	logger.Info("play(%s) --> [twinx proxy] --> publish(%s)", r.playAddr.SafeURL(), r.publishAddr.SafeURL())

	return nil
}

func (r *RTMPProxy) Stop() {
	if r.isStreaming {
		r.proxyMessageCh <- StopProxy
		r.isStreaming = false
	}
}
