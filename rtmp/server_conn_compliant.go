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
	"io"
	"time"

	"github.com/gwuhaolin/livego/protocol/amf"

	"github.com/gwuhaolin/livego/utils/pio"
	"github.com/kris-nova/logger"
)

func (s *ServerConn) handshake() error {
	var err error
	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S0S1 := S0S1S2[:1536+1]
	S2 := S0S1S2[1536+1:]

	// < C0C1
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = io.ReadFull(s.conn.rw, C0C1); err != nil {
		return err
	}
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if C0[0] != 3 {
		return fmt.Errorf("rtmp: handshake version=%d invalid", C0[0])
	}

	S0[0] = 3

	clitime := pio.U32BE(C1[0:4])
	srvtime := clitime
	srvver := uint32(0x0d0e0a0d)
	cliver := pio.U32BE(C1[4:8])

	if cliver != 0 {
		var ok bool
		var digest []byte
		if ok, digest = hsParse1(C1, HandshakeClientPartial30, HandshakeServerKey); !ok {
			err = fmt.Errorf("rtmp: handshake server: C1 invalid")
			return err
		}
		hsCreate01(S0S1, srvtime, srvver, HandshakeServerPartial36)
		hsCreate2(S2, digest)
	} else {
		copy(S1, C2)
		copy(S2, C1)
	}

	// > S0S1S2
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = s.conn.rw.Write(S0S1S2); err != nil {
		return err
	}
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if err = s.conn.rw.Flush(); err != nil {
		return err
	}

	// < C2
	s.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = io.ReadFull(s.conn.rw, C2); err != nil {
		return err
	}
	s.conn.SetDeadline(time.Time{})
	logger.Debug(rtmpMessage(thisFunctionName(), hs))
	return nil
}

// connectRX
//
// Example raw data from logs:
//   0: connect
//   1: 1
//   2: map[app:twinx flashVer:FMLE/3.0 (compatible; FMSc/1.0) swfUrl:rtmp://localhost:1935/twinx tcUrl:rtmp://localhost:1935/twinx type:nonprivate]
func (s *ServerConn) connectRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	// ---
	if len(x.batchedValues) == 0 {
		return errors.New("missing values")
	}
	if len(x.batchedValues) < 3 {
		return fmt.Errorf("invalid connect command length [%d] < 3", len(x.batchedValues))
	}
	rxID := x.batchedValues[1]
	id, ok := rxID.(float64)
	if !ok {
		return errors.New("invalid ID field")
	}
	s.transactionID = int64(id)
	// ---

	if id != CommandConnectWellKnownID {
		return fmt.Errorf("invalid connect id: %v", rxID)
	}
	rxConnInfoMap := x.batchedValues[2]
	rxConnInfo, err := ConnectInfoMapToInstance(rxConnInfoMap)
	if err != nil {
		return fmt.Errorf("building connect info: %v", err)
	}
	s.connectInfo = rxConnInfo
	s.connectPacket = x
	logger.Debug(rtmpMessage(thisFunctionName(), ack))

	// Server code should just TX right away
	_, err = s.connectTX()
	return err
}

func (s *ServerConn) connectTX() (*ChunkStream, error) {
	var txPacket *ChunkStream
	var err error

	// WindowAcknowledgement
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), "WindowAcknowledgement"), tx))
	txPacket = s.conn.newChunkStreamWindowAcknowledgementMessage(DefaultWindowAcknowledgementSizeBytes)
	err = s.conn.Write(txPacket)
	if err != nil {
		return nil, err
	}

	// SetPeerBandwidth
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), "SetPeerBandwidth"), tx))
	txPacket = s.conn.newChunkStreamSetPeerBandwidth(DefaultPeerBandwidthSizeBytes)
	err = s.conn.Write(txPacket)
	if err != nil {
		return nil, err
	}

	// SetChunkSize
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), "SetChunkSize"), tx))
	txPacket = s.conn.newChunkStreamSetChunkSize(DefaultRTMPChunkSizeBytesLarge)
	err = s.conn.Write(txPacket)
	if err != nil {
		return nil, err
	}

	// Compliant connect response [response]
	// TODO Use any existing meta fields
	resp := make(amf.Object)
	//if s.connectInfo == nil {
	//	resp[ConnRespFMSVer] = DefaultServerFMSVersion
	//} else {
	//	resp[ConnRespFMSVer] = s.connectInfo.FlashVer
	//}
	resp[ConnRespFMSVer] = DefaultServerFMSVersion
	resp[ConnRespCapabilities] = 31

	// Compliant connect response [event]
	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamConnectSuccess
	event[ConnEventDescription] = "Connection succeeded."
	event[ConnEventObjectEncoding] = s.connectInfo.ObjectEncoding

	// Write out _result message
	err = s.writeMsg(s.connectPacket.CSID, s.connectPacket.StreamID, CommandType_Result, s.transactionID, resp, event)
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), "Response [_result]"), tx))
	return nil, err
}

// Example raw data from logs:
//   0: createStream
//   1: 2
//   2: <nil>
func (s *ServerConn) createStreamRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	if len(x.batchedValues) == 0 {
		return errors.New("missing values")
	}
	if len(x.batchedValues) < 3 {
		return fmt.Errorf("invalid createStream command length [%d] < 3", len(x.batchedValues))
	}

	rxID := x.batchedValues[1]
	id, ok := rxID.(float64)
	if !ok {
		return errors.New("invalid ID field, unable to type cast float64")
	}
	s.transactionID = int64(id)
	logger.Debug(rtmpMessage(thisFunctionName(), ack))

	_, err := s.createStreamTX()
	return err
}

func (s *ServerConn) createStreamTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), "Response [_result]"), tx))
	err := s.writeMsg(s.connectPacket.CSID, s.connectPacket.StreamID, CommandType_Result, s.transactionID, nil, 1)
	return nil, err
}

//   +--------------+----------+-----------------------------------------+
//   | Field Name   |   Type   |             Description                 |
//   +--------------+----------+-----------------------------------------+
// 0 | Command Name |  String  | Name of the command. Set to "play".     |
//   +--------------+----------+-----------------------------------------+
// 1 | Transaction  |  Number  | Transaction ID set to 0.                |
//   | ID           |          |                                         |
//   +--------------+----------+-----------------------------------------+
// 2 | Command      |   Null   | Command information does not exist.     |
//   | Object       |          | Set to null type.                       |
//   +--------------+----------+-----------------------------------------+
// 3 | Stream Name  |  String  | Name of the stream to play.             |
//   |              |          | To play video (FLV) files, specify the  |
//   |              |          | name of the stream without a file       |
//   |              |          | extension (for example, "sample"). To   |
//   |              |          | play back MP3 or ID3 tags, you must     |
//   |              |          | precede the stream name with mp3:       |
//   |              |          | (for example, "mp3:sample". To play     |
//   |              |          | H.264/AAC files, you must precede the   |
//   |              |          | stream name with mp4: and specify the   |
//   |              |          | file extension. For example, to play the|
//   |              |          | file sample.m4v,specify "mp4:sample.m4v"|
//   |              |          |                                         |
//   +--------------+----------+-----------------------------------------+
// 4 | Start        |  Number  | An optional parameter that specifies    |
//   |              |          | the start time in seconds. The default  |
//   |              |          | value is -2, which means the subscriber |
//   |              |          | first tries to play the live stream     |
//   |              |          | specified in the Stream Name field. If a|
//   |              |          | live stream of that name is not found,it|
//   |              |          | plays the recorded stream of the same   |
//   |              |          | name. If there is no recorded stream    |
//   |              |          | with that name, the subscriber waits for|
//   |              |          | a new live stream with that name and    |
//   |              |          | plays it when available. If you pass -1 |
//   |              |          | in the Start field, only the live stream|
//   |              |          | specified in the Stream Name field is   |
//   |              |          | played. If you pass 0 or a positive     |
//   |              |          | number in the Start field, a recorded   |
//   |              |          | stream specified in the Stream Name     |
//   |              |          | field is played beginning from the time |
//   |              |          | specified in the Start field. If no     |
//   |              |          | recorded stream is found, the next item |
//   |              |          | in the playlist is played.              |
//   +--------------+----------+-----------------------------------------+
// 5 | Duration     |  Number  | An optional parameter that specifies the|
//   |              |          | duration of playback in seconds. The    |
//   |              |          | default value is -1. The -1 value means |
//   |              |          | a live stream is played until it is no  |
//   |              |          | longer available or a recorded stream is|
//   |              |          | played until it ends. If you pass 0, it |
//   |              |          | plays the single frame since the time   |
//   |              |          | specified in the Start field from the   |
//   |              |          | beginning of a recorded stream. It is   |
//   |              |          | assumed that the value specified in     |
//   |              |          | the Start field is equal to or greater  |
//   |              |          | than 0. If you pass a positive number,  |
//   |              |          | it plays a live stream for              |
//   |              |          | the time period specified in the        |
//   |              |          | Duration field. After that it becomes   |
//   |              |          | available or plays a recorded stream    |
//   |              |          | for the time specified in the Duration  |
//   |              |          | field. (If a stream ends before the     |
//   |              |          | time specified in the Duration field,   |
//   |              |          | playback ends when the stream ends.)    |
//   |              |          | If you pass a negative number other     |
//   |              |          | than -1 in the Duration field, it       |
//   |              |          | interprets the value as if it were -1.  |
//   +--------------+----------+-----------------------------------------+
// 6 | Reset        | Boolean  | An optional Boolean value or number     |
//   |              |          | that specifies whether to flush any     |
//   |              |          | previous playlist.                      |
//   +--------------+----------+-----------------------------------------+
// Example raw data from logs:
//   0: play
//   1: 0
//   2: <nil>
//   3: twinx_XVlBzgbaiCMRAjWwhTHc
func (s *ServerConn) playRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	if len(x.batchedValues) == 0 {
		return errors.New("missing values")
	}
	if len(x.batchedValues) < 4 {
		return fmt.Errorf("invalid play command length [%d] < 4", len(x.batchedValues))
	}

	rxID := x.batchedValues[1]
	id, ok := rxID.(float64)
	if !ok {
		return errors.New("invalid ID field, unable to type cast float64")
	}
	s.transactionID = int64(id)
	logger.Debug(rtmpMessage(thisFunctionName(), ack))

	_, err := s.playTX()
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerConn) playTX() (*ChunkStream, error) {

	err := s.conn.setRecorded()
	if err != nil {
		return nil, err
	}
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), "SetRecorded"), tx))

	// NetStream.Play.Reset
	event := make(amf.Object)
	//event[ConnEventLevel] = ConnEventStatus
	//event[ConnEventCode] = CommandNetStreamPlayReset
	//event[ConnEventDescription] = "Playing and resetting stream."
	//if err := s.writeMsg(s.connectPacket.CSID, s.connectPacket.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
	//	return nil, err
	//}
	//logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), CommandNetStreamPlayReset), tx))

	// NetStream.Play.Start
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPlayStart
	event[ConnEventDescription] = "Start live"
	if err := s.writeMsg(s.connectPacket.CSID, s.connectPacket.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
		return nil, err
	}
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), CommandNetStreamPlayStart), tx))

	// NetStream.Data.Start
	//event[ConnEventLevel] = ConnEventStatus
	//event[ConnEventCode] = CommandNetStreamDataStart
	//event[ConnEventDescription] = "Started playing stream."
	//if err := s.writeMsg(s.connectPacket.CSID, s.connectPacket.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
	//	return nil, err
	//}
	//logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), CommandNetStreamDataStart), tx))

	// 	NetStream.Publish.Notify
	//event[ConnEventLevel] = ConnEventStatus
	//event[ConnEventCode] = CommandNetStreamPublishNotify
	//event[ConnEventDescription] = "Started playing notify."
	//if err := s.writeMsg(s.connectPacket.CSID, s.connectPacket.StreamID, CommandTypeOnStatus, 0, nil, event); err != nil {
	//	return nil, err
	//}
	//logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), CommandNetStreamPublishNotify), tx))

	return nil, nil
}

func (s *ServerConn) play2RX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) play2TX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) deleteStreamRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) deleteStreamTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) receiveAudioRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) receiveAudioTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) receiveVideoRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) receiveVideoTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

//   +--------------+----------+----------------------------------------+
//   | Field Name   |   Type   |             Description                |
//   +--------------+----------+----------------------------------------+
// 0 | Command Name |  String  | Name of the command, set to "publish". |
//   +--------------+----------+----------------------------------------+
// 1 | Transaction  |  Number  | Transaction ID set to 0.               |
//   | ID           |          |                                        |
//   +--------------+----------+----------------------------------------+
// 2 | Command      |  Null    | Command information object does not    |
//   | Object       |          | exist. Set to null type.               |
//   +--------------+----------+----------------------------------------+
// 3 | Publishing   |  String  | Name with which the stream is          |
//   | Name         |          | published.                             |
//   +--------------+----------+----------------------------------------+
// 4 | Publishing   |  String  | Type of publishing. Set to "live",     |
//   | Type         |          | "record", or "append".                 |
//   |              |          | record: The stream is published and the|
//   |              |          | data is recorded to a new file.The file|
//   |              |          | is stored on the server in a           |
//   |              |          | subdirectory within the directory that |
//   |              |          | contains the server application. If the|
//   |              |          | file already exists, it is overwritten.|
//   |              |          | append: The stream is published and the|
//   |              |          | data is appended to a file. If no file |
//   |              |          | is found, it is created.               |
//   |              |          | live: Live data is published without   |
//   |              |          | recording it in a file.                |
//   +--------------+----------+----------------------------------------+
//
// Example raw data from logs:
//   0: publish
//   1: 3
//   2: <nil>
//   3: twinx_XVlBzgbaiCMRAjWwhTHc
//   4: live
func (s *ServerConn) publishRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	if len(x.batchedValues) == 0 {
		return errors.New("missing values")
	}
	if len(x.batchedValues) < 5 {
		return fmt.Errorf("invalid publish command length [%d] < 5", len(x.batchedValues))
	}

	rxID := x.batchedValues[1]
	id, ok := rxID.(float64)
	if !ok {
		return errors.New("invalid ID field, unable to type cast float64")
	}
	s.transactionID = int64(id)
	publishInfo := &PublishInfo{
		Name: x.batchedValues[3].(string),
		Type: x.batchedValues[4].(string),
	}
	s.publishInfo = publishInfo
	logger.Debug(rtmpMessage(thisFunctionName(), ack))

	_, err := s.publishTX()
	return err
}

func (s *ServerConn) publishTX() (*ChunkStream, error) {

	event := make(amf.Object)
	event[ConnEventLevel] = ConnEventStatus
	event[ConnEventCode] = CommandNetStreamPublishStart
	event[ConnEventDescription] = "Start publishing."
	err := s.writeMsg(s.connectPacket.CSID, s.connectPacket.StreamID, CommandTypeOnStatus, 0, nil, event)
	if err != nil {
		return nil, err
	}
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), CommandNetStreamPublishStart), tx))

	return nil, nil
}

func (s *ServerConn) seekRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) seekTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) pauseRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (s *ServerConn) pauseTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (s *ServerConn) oosFCPublishRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return nil
}

func (s *ServerConn) oosFCPublishTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, nil
}

func (s *ServerConn) oosReleaseStreamRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return nil
}

func (s *ServerConn) oosReleaseStreamTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, nil
}
