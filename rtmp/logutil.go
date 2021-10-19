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
	rx     messageOperator = "[← 💻  ]"
	tx     messageOperator = "[  💻 →]"
	ack    messageOperator = "[  ✨  ]"
	hs     messageOperator = "[  🤝  ]"
	pub    messageOperator = "[  📝  ]"
	play   messageOperator = "[  ⏯  ]"
	conn   messageOperator = "[  📶  ]"
	stream messageOperator = "[→ 🌊 →]"
	fork   messageOperator = "[← 🍴 →]"
	proxy  messageOperator = "[← 💻 →]"
	warn   messageOperator = "[  ⚠  ]"
	danger messageOperator = "[  🧨  ]"
	start  messageOperator = "[  ⏱  ]"
	stop   messageOperator = "[  ⏹  ]"
	seek   messageOperator = "[  ⏩  ]"
	listen messageOperator = "[  🙉  ]"
	serve  messageOperator = "[  🍽  ]"
	new    messageOperator = "[  🆕  ]"
)

// Send an RTMP protocol message with an operator
func rtmpServerMessage(msg string, op messageOperator) string {
	return fmt.Sprintf("[rtmp.server] %s (%s)", op, msg)
}

// Send an RTMP protocol message with an operator
func rtmpClientMessage(msg string, op messageOperator) string {
	return fmt.Sprintf("[rtmp.client] %s (%s)", op, msg)
}

// Send an RTMP protocol message with an operator
func rtmpMessage(msg string, op messageOperator) string {
	return fmt.Sprintf("[rtmp] %s (%s)", op, msg)
}

var (
	DefaultUnimplementedError = errors.New("**UNIMPLEMENTED**")
)

func defaultUnimplemented() error {
	pc := make([]uintptr, 1)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return DefaultUnimplementedError
	}
	caller := runtime.FuncForPC(pc[0] - 1)
	if caller == nil {
		return DefaultUnimplementedError
	}
	return fmt.Errorf("function %s %v", caller.Name(), DefaultUnimplementedError)
}
