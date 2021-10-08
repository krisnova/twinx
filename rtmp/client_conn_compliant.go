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

import "github.com/kris-nova/logger"

func (cc *ClientConn) handshake() error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) connectRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) connectTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) createStreamRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) createStreamTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) playRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) playTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) play2RX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) play2TX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) deleteStreamRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) deleteStreamTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) receiveAudioRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) receiveAudioTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) receiveVideoRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) receiveVideoTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) publishRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) publishTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) seekRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) seekTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) pauseRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (cc *ClientConn) pauseTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}
