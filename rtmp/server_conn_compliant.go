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

func (s *ServerConn) handshake() error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) connectRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) connectTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) createStreamRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) createStreamTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) playRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) playTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) play2RX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) play2TX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) deleteStreamRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) deleteStreamTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) receiveAudioRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) receiveAudioTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) receiveVideoRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) receiveVideoTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) publishRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) publishTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) seekRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) seekTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) pauseRX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}

func (s *ServerConn) pauseTX(x *ChunkStream) error {
	logger.Debug(thisFunctionName())
	return defaultUnimplemented()
}
