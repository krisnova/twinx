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
	"runtime"
	"strings"
)

func thisFunctionName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(2, fpcs)
	if n == 0 {
		return "unknown function"
	}
	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return "unknown function"
	}
	// filename and line number caller.FileLine(fpcs[0]-1)
	fullName := caller.Name()
	spl := strings.Split(fullName, ".")
	return spl[3]
}

type messageOperator string

const (
	rx  messageOperator = "[⬅ 💻  ]"
	tx  messageOperator = "[  💻 ➡]"
	ack messageOperator = "[  ✨  ]"
	hs  messageOperator = "[  🤝  ]"
)

// Send an RTMP protocol message with an operator
//
// Operators used in this convention.
//   ->  Transmit (TX) out to a remote
//   <-  Receive (RX) in to a local
//   *   Ack (ack) mutate a process based on the content of a message
func rtmpMessage(place string, op messageOperator) string {
	return fmt.Sprintf("[rtmp] %s (%s)", op, place)
}
