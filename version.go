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

package twinx

import "fmt"

// Version is set at compile time in the associated Makefile
// Do not change this!
var Version string

func PrintBanner() {
	fmt.Printf(Banner())
}

func Banner() string {
	var str string
	str += fmt.Sprintf("\n")
	str += fmt.Sprintf("┌──────────────────────────────────────────┐\n")
	str += fmt.Sprintf("│                                          │\n")
	str += fmt.Sprintf("│ ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗ │\n")
	str += fmt.Sprintf("│ ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝ │\n")
	str += fmt.Sprintf("│    ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝  │\n")
	str += fmt.Sprintf("│    ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗  │\n")
	str += fmt.Sprintf("│    ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗ │\n")
	str += fmt.Sprintf("│    ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝ │\n")
	str += fmt.Sprintf("│    A live streaming command line tool.   │\n")
	str += fmt.Sprintf("└──────────────────────────────────────────┘\n")
	str += fmt.Sprintf("\n")
	return str
}
