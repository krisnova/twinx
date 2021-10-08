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

import "github.com/kris-nova/logger"

type Client struct {
	conn    *ConnClient
	service *Service
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Dial(address string) error {
	clientConn := NewConnClient()
	err := clientConn.Dial(address)
	if err != nil {
		return err
	}
	c.service = NewService(clientConn.urladdr.Key())
	c.conn = clientConn
	return nil
}

func (c *Client) stream() error {
	// Here is where we handle the service.
	logger.Info("Streaming...")
	if c.conn.method == ClientMethodPublish {
		writer := NewVirtualWriter(c.conn)
		c.service.HandleWriter(writer)
	} else if c.conn.method == ClientMethodPlay {
		reader := NewVirtualReader(c.conn)
		c.service.HandleReader(reader.UID, reader)
	}
	for {
		// Stream until otherwise cancelled
		// TODO add a signal handler
	}
	return nil
}

func (c *Client) Play() error {
	err := c.conn.Play()
	if err != nil {
		return err
	}
	// We should be connected at this point
	return c.stream()
}

func (c *Client) Publish() error {
	err := c.conn.Publish()
	if err != nil {
		return err
	}
	// We should be connected at this point
	return c.stream()
}

func (c *Client) Client() *ConnClient {
	return c.conn
}
