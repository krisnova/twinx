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
	"bytes"
	"fmt"
	"os"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/kris-nova/logger"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
)

type Cache struct {
	gop      *GopCache
	videoSeq *SpecialCache
	audioSeq *SpecialCache
	metadata *SpecialCache
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(configure.Config.GetInt("gop_num")),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

func (cache *Cache) Write(p av.Packet) {
	if p.IsMetadata {
		cache.metadata.Write(&p)
		return
	} else {
		if !p.IsVideo {
			ah, ok := p.Header.(av.AudioPacketHeader)
			if ok {
				if ah.SoundFormat() == av.SOUND_AAC &&
					ah.AACPacketType() == av.AAC_SEQHDR {
					cache.audioSeq.Write(&p)
					return
				} else {
					return
				}
			}

		} else {
			vh, ok := p.Header.(av.VideoPacketHeader)
			if ok {
				if vh.IsSeq() {
					cache.videoSeq.Write(&p)
					return
				}
			} else {
				return
			}

		}
	}
	cache.gop.Write(&p)
}

func (cache *Cache) Send(w av.WriteCloser) error {
	if err := cache.metadata.Send(w); err != nil {
		return err
	}

	if err := cache.videoSeq.Send(w); err != nil {
		return err
	}

	if err := cache.audioSeq.Send(w); err != nil {
		return err
	}

	if err := cache.gop.Send(w); err != nil {
		return err
	}

	return nil
}

var (
	maxGOPCap    int = 1024
	ErrGopTooBig     = fmt.Errorf("gop to big")
)

type array struct {
	index   int
	packets []*av.Packet
}

func newArray() *array {
	ret := &array{
		index:   0,
		packets: make([]*av.Packet, 0, maxGOPCap),
	}
	return ret
}

func (array *array) reset() {
	array.index = 0
	array.packets = array.packets[:0]
}

func (array *array) write(packet *av.Packet) error {
	if array.index >= maxGOPCap {
		return ErrGopTooBig
	}
	array.packets = append(array.packets, packet)
	array.index++
	return nil
}

func (array *array) send(w av.WriteCloser) error {
	var err error
	for i := 0; i < array.index; i++ {
		packet := array.packets[i]
		if err = w.Write(packet); err != nil {
			return err
		}
	}
	return err
}

type GopCache struct {
	start     bool
	num       int
	count     int
	nextindex int
	gops      []*array
}

func NewGopCache(num int) *GopCache {
	return &GopCache{
		count: num,
		gops:  make([]*array, num),
	}
}

func (gopCache *GopCache) writeToArray(chunk *av.Packet, startNew bool) error {
	var ginc *array
	if startNew {
		ginc = gopCache.gops[gopCache.nextindex]
		if ginc == nil {
			ginc = newArray()
			gopCache.num++
			gopCache.gops[gopCache.nextindex] = ginc
		} else {
			ginc.reset()
		}
		gopCache.nextindex = (gopCache.nextindex + 1) % gopCache.count
	} else {
		ginc = gopCache.gops[(gopCache.nextindex+1)%gopCache.count]
	}
	ginc.write(chunk)

	return nil
}

func (gopCache *GopCache) Write(p *av.Packet) {
	var ok bool
	if p.IsVideo {
		vh := p.Header.(av.VideoPacketHeader)
		if vh.IsKeyFrame() && !vh.IsSeq() {
			ok = true
		}
	}
	if ok || gopCache.start {
		gopCache.start = true
		gopCache.writeToArray(p, ok)
	}
}

func (gopCache *GopCache) sendTo(w av.WriteCloser) error {
	var err error
	pos := (gopCache.nextindex + 1) % gopCache.count
	for i := 0; i < gopCache.num; i++ {
		index := (pos - gopCache.num + 1) + i
		if index < 0 {
			index += gopCache.count
		}
		g := gopCache.gops[index]
		err = g.send(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gopCache *GopCache) Send(w av.WriteCloser) error {
	return gopCache.sendTo(w)
}

const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}
	if _, err := encoder.Encode(b, SetDataFrame, amf.AMF0); err != nil {
		logger.Critical(err.Error())
		os.Exit(1)
	}
	setFrameFrame = b.Bytes()
}

type SpecialCache struct {
	full bool
	p    *av.Packet
}

func NewSpecialCache() *SpecialCache {
	return &SpecialCache{}
}

func (specialCache *SpecialCache) Write(p *av.Packet) {
	specialCache.p = p
	specialCache.full = true
}

func (specialCache *SpecialCache) Send(w av.WriteCloser) error {
	if !specialCache.full {
		return nil
	}
	return w.Write(specialCache.p)
}
