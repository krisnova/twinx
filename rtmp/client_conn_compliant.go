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
	"time"

	"github.com/gwuhaolin/livego/protocol/amf"

	"github.com/gwuhaolin/livego/utils/pio"
	"github.com/kris-nova/logger"
)

func (cc *ClientConn) handshake() error {
	var err error
	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]
	S0S1S2 := random[1536*2+1:]
	C0[0] = 3
	// > C0C1
	cc.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = cc.conn.rw.Write(C0C1); err != nil {
		return err
	}
	cc.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if err = cc.conn.rw.Flush(); err != nil {
		return err
	}

	// < S0S1S2
	cc.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = io.ReadFull(cc.conn.rw, S0S1S2); err != nil {
		return err
	}

	S1 := S0S1S2[1 : 1536+1]
	if ver := pio.U32BE(S1[4:8]); ver != 0 {
		C2 = S1
	} else {
		C2 = S1
	}

	// > C2
	cc.conn.SetDeadline(time.Now().Add(TimeoutDurationSeconds))
	if _, err = cc.conn.rw.Write(C2); err != nil {
		return err
	}
	cc.conn.SetDeadline(time.Time{})
	logger.Debug(rtmpMessage(thisFunctionName(), hs))
	return nil
}

func (cc *ClientConn) connectRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return nil
}

func (cc *ClientConn) connectTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	event := make(amf.Object)
	event[ConnInfoKeyApp] = cc.urladdr.App()
	event[ConnInfoKeyType] = "nonprivate"
	event[ConnInfoKeyFlashVer] = DefaultServerFMSVersion
	event[ConnInfoKeyTcURL] = cc.urladdr.SafeURL()
	event[ConnInfoKeySWFURL] = cc.urladdr.SafeURL()
	cc.curcmdName = CommandConnect
	return cc.writeMsg(CommandConnect, cc.transID, event)
}

func (cc *ClientConn) createStreamRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	if len(x.batchedValues) == 0 {
		return errors.New("missing values")
	}
	if len(x.batchedValues) < 4 {
		return fmt.Errorf("invalid createStream command length [%d] < 4", len(x.batchedValues))
	}
	event, err := ConnEventMapToInstance(x.batchedValues[3])
	if err != nil {
		return fmt.Errorf("unable to parse ConnEvent amf object")
	}
	if event.Code == CommandNetStreamConnectSuccess {
		cc.connected = true
	} else {
		return fmt.Errorf("createStream failure: %s", event.Code)
	}
	return nil
}

//  The command structure from the client to the server is as follows:
//
//    +--------------+----------+----------------------------------------+
//    | Field Name   |   Type   |             Description                |
//    +--------------+----------+----------------------------------------+
//    | Command Name |  String  | Name of the command. Set to            |
//    |              |          | "createStream".                        |
//    +--------------+----------+----------------------------------------+
//    | Transaction  |  Number  | Transaction ID of the command.         |
//    | ID           |          |                                        |
//    +--------------+----------+----------------------------------------+
//    | Command      |  Object  | If there exists any command info this  |
//    | Object       |          | is set, else this is set to null type. |
//    +--------------+----------+----------------------------------------+
//
func (cc *ClientConn) createStreamTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	cc.transID++
	cc.curcmdName = CommandCreateStream
	_, err := cc.writeMsg(CommandCreateStream, cc.transID, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to send createStream message: %v", err)
	}
	//logger.Debug(rtmpMessage(fmt.Sprintf("%s.Connected=true", thisFunctionName()), tx))
	//cc.connected = true
	return nil, nil
}

func (cc *ClientConn) playRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	logger.Debug(rtmpMessage(thisFunctionName(), ack))
	return nil
}

func (cc *ClientConn) playTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	cc.transID++
	cc.curcmdName = CommandPlay
	return cc.writeMsg(CommandPlay, 0, nil, cc.urladdr.Key())
}

func (cc *ClientConn) play2RX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (cc *ClientConn) play2TX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (cc *ClientConn) deleteStreamRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (cc *ClientConn) deleteStreamTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (cc *ClientConn) receiveAudioRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (cc *ClientConn) receiveAudioTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (cc *ClientConn) receiveVideoRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (cc *ClientConn) receiveVideoTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

//  The command structure from server to client is as follows:
//
//    +--------------+----------+----------------------------------------+
//    | Field Name   |   Type   |             Description                |
//    +--------------+----------+----------------------------------------+
// 0  | Command Name |  String  | _result or _error; indicates whether   |
//    |              |          | the response is result or error.       |
//    +--------------+----------+----------------------------------------+
// 1  | Transaction  |  Number  | ID of the command that response belongs|
//    | ID           |          | to.                                    |
//    +--------------+----------+----------------------------------------+
// 2  | Command      |  Object  | If there exists any command info this  |
//    | Object       |          | is set, else this is set to null type. |
//    +--------------+----------+----------------------------------------+
// 3  | Stream       |  Number  | The return value is either a stream ID |
//    | ID           |          | or an error information object.        |
//    +--------------+----------+----------------------------------------+
//
func (cc *ClientConn) publishRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	cc.transID = int(x.batchedValues[1].(float64))
	return nil
}

// The command structure from the client to the server is as follows:
//
//    +--------------+----------+----------------------------------------+
//    | Field Name   |   Type   |             Description                |
//    +--------------+----------+----------------------------------------+
//    | Command Name |  String  | Name of the command, set to "publish". |
//    +--------------+----------+----------------------------------------+
//    | Transaction  |  Number  | Transaction ID set to 0.               |
//    | ID           |          |                                        |
//    +--------------+----------+----------------------------------------+
//    | Command      |  Null    | Command information object does not    |
//    | Object       |          | exist. Set to null type.               |
//    +--------------+----------+----------------------------------------+
//    | Publishing   |  String  | Name with which the stream is          |
//    | Name         |          | published.                             |
//    +--------------+----------+----------------------------------------+
//    | Publishing   |  String  | Type of publishing. Set to "live",     |
//    | Type         |          | "record", or "append".                 |
//    |              |          | record: The stream is published and the|
//    |              |          | data is recorded to a new file.The file|
//    |              |          | is stored on the server in a           |
//    |              |          | subdirectory within the directory that |
//    |              |          | contains the server application. If the|
//    |              |          | file already exists, it is overwritten.|
//    |              |          | append: The stream is published and the|
//    |              |          | data is appended to a file. If no file |
//    |              |          | is found, it is created.               |
//    |              |          | live: Live data is published without   |
//    |              |          | recording it in a file.                |
//    +--------------+----------+----------------------------------------+
//
func (cc *ClientConn) publishTX() (*ChunkStream, error) {

	// SetChunkSize
	logger.Debug(rtmpMessage(fmt.Sprintf("%s.%s", thisFunctionName(), "SetChunkSize"), tx))
	txPacket := cc.conn.newChunkStreamSetChunkSize(DefaultRTMPChunkSizeBytesLarge)
	err := cc.conn.Write(txPacket)
	if err != nil {
		return nil, err
	}

	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	cc.transID++
	cc.curcmdName = CommandPublish
	return cc.writeMsg(CommandPublish, cc.transID, nil, cc.urladdr.Key(), PublishCommandLive)
}

func (cc *ClientConn) seekRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (cc *ClientConn) seekTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}

func (cc *ClientConn) pauseRX(x *ChunkStream) error {
	logger.Debug(rtmpMessage(thisFunctionName(), rx))
	return defaultUnimplemented()
}

func (cc *ClientConn) pauseTX() (*ChunkStream, error) {
	logger.Debug(rtmpMessage(thisFunctionName(), tx))
	return nil, defaultUnimplemented()
}
