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
	"fmt"
)

type Client struct {
	handler Handler
	getter  GetWriter
}

func NewRtmpClient(h Handler, getter GetWriter) *Client {
	return &Client{
		handler: h,
		getter:  getter,
	}
}

func (c *Client) Connect(urladdr *URLAddr, method ClientMethod) error {
	cc := NewConnClient()
	switch method {

	// Connect Publish
	case ClientMethodPublish:
		if err := cc.StartPublish(urladdr); err != nil {
			return fmt.Errorf("publish dial %s: %v", urladdr.SafeURL(), err)
		}
		writer := NewVirtualWriter(cc)
		c.handler.HandleWriter(writer)

	// Connect Play
	case ClientMethodPlay:
		if err := cc.StartPublish(urladdr); err != nil {
			return fmt.Errorf("play dial %s: %v", urladdr.SafeURL(), err)
		}
		reader := NewVirtualReader(cc)
		c.handler.HandleReader(reader)
		if c.getter != nil {
			writer := c.getter.GetWriter(reader.Info())
			c.handler.HandleWriter(writer)
		}
	default:
		return fmt.Errorf("unsupported client method method: %s", method)
	}
	return nil
}

func (c *Client) GetHandle() Handler {
	return c.handler
}
