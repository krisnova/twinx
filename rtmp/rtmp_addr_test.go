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

import "testing"

func TestAddrs(t *testing.T) {

	happyCases := map[string]*Addr{
		"rtmp://localhost:1935/twinx/1234": &Addr{
			host:   "localhost:1935",
			scheme: "rtmp",
			app:    "twinx",
			key:    "1234",
		},
		"": &Addr{
			host:   "localhost:1935",
			scheme: "rtmp",
			app:    "twinx",
		},
		"localhost": &Addr{
			host:   "localhost:1935",
			scheme: "rtmp",
			app:    "twinx",
		},
		"rtmp://localhost": &Addr{
			host:   "localhost:1935",
			scheme: "rtmp",
			app:    "twinx",
		},
		"rtmp://localhost:1313": &Addr{
			host:   "localhost:1313",
			scheme: "rtmp",
			app:    "twinx",
		},
		"rtmp://localhost:1313/beeps/boops": &Addr{
			host:   "localhost:1313",
			scheme: "rtmp",
			app:    "beeps",
		},
	}
	for input, expected := range happyCases {
		actual, err := NewAddr(input)
		if err != nil {
			t.Errorf("happyCase error %v", err)
		}
		if actual == nil {
			t.Errorf("nil actual")
			t.FailNow()
		}
		if expected == nil {
			t.Errorf("nil expected")
			t.FailNow()
		}
		if !assertAddrs(actual, expected) {
			t.Errorf("Expected: %+v", expected)
			t.Errorf("Actual: %+v", actual)
		}
		if expected.key != "" {
			if !assertKeys(actual, expected) {
				t.Errorf("Expected key: %s", expected.key)
				t.Errorf("Actual key: %s", actual.key)
			}
		} else {
			// Validate a key was generated
			if actual.key == "" {
				t.Errorf("Failed generating key for raw: %s", actual.raw)
			}
		}
	}

}

func assertAddrs(a, b *Addr) bool {
	if a == nil || b == nil {
		return false
	}
	if a.app != b.app {
		return false
	}
	if a.host != b.host {
		return false
	}
	if a.scheme != b.scheme {
		return false
	}
	return true
}

func assertKeys(a, b *Addr) bool {
	if a.key != b.key {
		return false
	}
	return true
}
