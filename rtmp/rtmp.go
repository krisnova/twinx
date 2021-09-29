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
	"fmt"
	"io"
	"sync"
	"time"
)

const (
	DefaultProtocol          string = "tcp"
	DefaultLocalHost         string = "localhost"
	DefaultLocalPort         string = "1935"
	DefaultScheme            string = "rtmp"
	DefaultRTMPApp           string = "twinx"
	DefaultGenerateKeyLength int    = 20
	DefaultGenerateKeyPrefix string = "twinx_"
	StreamKeyRandomBytePool  string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	TAG_AUDIO                   uint32 = 8
	TAG_VIDEO                   uint32 = 9
	TAG_SCRIPTDATAAMF0          uint32 = 18
	TAG_SCRIPTDATAAMF3          uint32 = 0xf
	MetadatAMF0                 uint8  = 0x12
	MetadataAMF3                uint8  = 0xf
	SOUND_MP3                   uint8  = 2
	SOUND_NELLYMOSER_16KHZ_MONO uint8  = 4
	SOUND_NELLYMOSER_8KHZ_MONO  uint8  = 5
	SOUND_NELLYMOSER            uint8  = 6
	SOUND_ALAW                  uint8  = 7
	SOUND_MULAW                 uint8  = 8
	SOUND_AAC                   uint8  = 10
	SOUND_SPEEX                 uint8  = 11
	SOUND_5_5Khz                uint8  = 0
	SOUND_11Khz                 uint8  = 1
	SOUND_22Khz                 uint8  = 2
	SOUND_44Khz                 uint8  = 3
	SOUND_8BIT                  uint8  = 0
	SOUND_16BIT                 uint8  = 1
	SOUND_MONO                  uint8  = 0
	SOUND_STEREO                uint8  = 1
	AAC_SEQHDR                  uint8  = 0
	AAC_RAW                     uint8  = 1
	AVC_SEQHDR                  uint8  = 0
	AVC_NALU                    uint8  = 1
	AVC_EOS                     uint8  = 2
	FRAME_KEY                   uint8  = 1
	FRAME_INTER                 uint8  = 2
	VIDEO_H264                  uint8  = 7
)

type Packet struct {
	IsAudio    bool
	IsVideo    bool
	IsMetadata bool
	TimeStamp  uint32 // dts
	StreamID   uint32
	Header     PacketHeader
	Data       []byte
}

type PacketHeader interface {
}

type AudioPacketHeader interface {
	PacketHeader
	SoundFormat() uint8
	AACPacketType() uint8
}

type VideoPacketHeader interface {
	PacketHeader
	IsKeyFrame() bool
	IsSeq() bool
	CodecID() uint8
	CompositionTime() int32
}

type Demuxer interface {
	Demux(*Packet) (ret *Packet, err error)
}

type Muxer interface {
	Mux(*Packet, io.Writer) error
}

type SampleRater interface {
	SampleRate() (int, error)
}

type CodecParser interface {
	SampleRater
	Parse(*Packet, io.Writer) error
}

type GetWriter interface {
	GetWriter(Info) WriteCloser
}

type Handler interface {
	HandleReader(ReadCloser)
	HandleWriter(WriteCloser)
}

type Alive interface {
	Alive() bool
}

type Closer interface {
	Info() Info
	Close(error)
}

type CalcTime interface {
	CalcBaseTimestamp()
}

type Info struct {
	Key   string
	URL   string
	UID   string
	Inter bool
}

func (info Info) IsInterval() bool {
	return info.Inter
}

func (info Info) String() string {
	return fmt.Sprintf("<key: %s, URL: %s, UID: %s, Inter: %v>",
		info.Key, info.URL, info.UID, info.Inter)
}

type ReadCloser interface {
	Closer
	Alive
	Read(*Packet) error
}

type WriteCloser interface {
	Closer
	Alive
	CalcTime
	Write(*Packet) error
}

type RWBaser struct {
	lock               sync.Mutex
	timeout            time.Duration
	PreTime            time.Time
	BaseTimestamp      uint32
	LastVideoTimestamp uint32
	LastAudioTimestamp uint32
}

func NewRWBaser(duration time.Duration) RWBaser {
	return RWBaser{
		timeout: duration,
		PreTime: time.Now(),
	}
}

func (rw *RWBaser) BaseTimeStamp() uint32 {
	return rw.BaseTimestamp
}

func (rw *RWBaser) CalcBaseTimestamp() {
	if rw.LastAudioTimestamp > rw.LastVideoTimestamp {
		rw.BaseTimestamp = rw.LastAudioTimestamp
	} else {
		rw.BaseTimestamp = rw.LastVideoTimestamp
	}
}

func (rw *RWBaser) RecTimeStamp(timestamp, typeID uint32) {
	if typeID == TAG_VIDEO {
		rw.LastVideoTimestamp = timestamp
	} else if typeID == TAG_AUDIO {
		rw.LastAudioTimestamp = timestamp
	}
}

func (rw *RWBaser) SetPreTime() {
	rw.lock.Lock()
	rw.PreTime = time.Now()
	rw.lock.Unlock()
}

func (rw *RWBaser) Alive() bool {
	rw.lock.Lock()
	b := !(time.Now().Sub(rw.PreTime) >= rw.timeout)
	rw.lock.Unlock()
	return b
}
