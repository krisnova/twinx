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

type Client struct {
	conn *ClientConn
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Dial(address string) error {
	clientConn := NewClientConn()
	err := clientConn.Dial(address)
	if err != nil {
		return err
	}
	c.conn = clientConn
	return nil
}

func (c *Client) Play() error {
	err := c.conn.Play()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Publish() error {
	err := c.conn.Publish()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Client() *ClientConn {
	return c.conn
}

// VirtualOBSOutputClientMetadata
//
// A virtual metadata object that resembles common OBS configuration.
//
// (map[2.1:false 3.1:false 4.0:false 4.1:false 5.1:false 7.1:false
// audiochannels:2 audiocodecid:10 audiodatarate:160 audiosamplerate:48000
// audiosamplesize:16 duration:0 encoder:obs-output module (libobs version 27.0.1-3)
// fileSize:0 framerate:30 height:720 stereo:true videocodecid:7 videodatarate:2500 width:1280])
func VirtualOBSOutputClientMetadata() *MetaData {
	return &MetaData{
		V21:             false,
		V31:             false,
		V40:             false,
		V41:             false,
		V51:             false,
		V71:             false,
		AudioChannels:   2,
		AudioCodecID:    10,
		AudioDataRate:   160,
		AudioSampleRate: 48000,
		AudioSampleSize: 16,
		Duration:        0,
		Encoder:         "obs-output module (libobs version 27.0.1-3)",
		FileSize:        0,
		FrameRate:       30,
		Height:          720,
		Stereo:          true,
		VideoCodecID:    7,
		VideoDataRate:   2500,
		Width:           1280,
	}
}
