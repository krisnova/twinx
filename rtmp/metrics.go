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
	"sync"
	"time"
)

// Metrics is a data structure that will aggregate metrics about our RTMP
// connections, and streams.
type Metrics struct {
	ServerAddrRX         string
	ServerTotalPacketsRX int
	ServerTotalBytesRX   int
	ServerPacketOffset   int
	ServerKeyHash        string
	PacketsPerSecond     float64
	StartTime            time.Time

	// Proxies is a map indexed on SafeURL()
	Proxies map[string]*ProxyMetrics

	sync.Mutex
}

type ProxyMetrics struct {
	ProxyAddrTX         string
	ProxyTotalBytesTX   int
	ProxyTotalPacketsTX int
	ProxyKeyHash        string
}

var m *Metrics

// M is a metrics singleton
func M() *Metrics {
	if m == nil {
		m = &Metrics{
			Proxies: make(map[string]*ProxyMetrics),
		}
		go m.begin()
		return m
	}
	return m
}

func P(name string) *ProxyMetrics {
	p, ok := M().Proxies[name]
	if !ok {
		p := &ProxyMetrics{
			ProxyAddrTX: name,
		}
		M().Proxies[name] = p
		return p
	}
	return p
}

func (metrics *Metrics) begin() {
	metrics.StartTime = time.Now()
	time.Sleep(time.Second * 1)
	go func() {
		for {
			// Flush every 10 seconds
			time.Sleep(time.Second * 10)
			metrics.Lock()
			metrics.StartTime = time.Now()
			metrics.PacketsPerSecond = 0
			metrics.ServerPacketOffset = metrics.ServerTotalPacketsRX
			metrics.Unlock()
		}
	}()
	for {
		// 	sec := d / Second
		//	nsec := d % Second
		//	return float64(sec) + float64(nsec)/1e9

		if metrics.ServerTotalBytesRX != 0 {
			d := float64(metrics.ServerTotalPacketsRX-metrics.ServerPacketOffset) / time.Since(metrics.StartTime).Seconds()
			metrics.PacketsPerSecond = d
		}
	}
}

//func (metrics *Metrics) everySecond() {
//	metrics.Lock()
//	defer metrics.Unlock()
//	metrics.SecondsElapsed++
//
//}

func (metrics *Metrics) String() string {
	var s string
	s += fmt.Sprintf("*************************************************************\n")
	s += fmt.Sprintf(" Server Listen Addr [%s]\n", metrics.ServerAddrRX)
	s += fmt.Sprintf("     Bytes RX :  [%d]\n", metrics.ServerTotalBytesRX)
	s += fmt.Sprintf("   Packets RX :  [%d]\n", metrics.ServerTotalPacketsRX)
	s += fmt.Sprintf(" Packets /sec :  [%f]\n", metrics.PacketsPerSecond)
	for _, proxy := range metrics.Proxies {
		s += fmt.Sprintf("  → Proxy Forward Addr [%s]\n", proxy.ProxyAddrTX)
		s += fmt.Sprintf("      Bytes  TX :  [%d]\n", proxy.ProxyTotalBytesTX)
		s += fmt.Sprintf("     Packets TX :  [%d]\n", proxy.ProxyTotalPacketsTX)
	}
	return s
}

func PrintMetrics(duration time.Duration) {
	for {
		time.Sleep(duration)
		fmt.Printf(M().String())
	}
}
