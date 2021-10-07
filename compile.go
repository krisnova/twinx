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
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗
//  ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝
//     ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝
//     ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗
//     ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗
//     ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

package twinx

import "fmt"

// CompileTimeVersion is set at compile time in the associated Makefile
// Do not change this!
var CompileTimeVersion string

// CompileFlagPrintBanner will enable/disable the banner for the program.
var CompileFlagPrintBanner bool = true

func PrintBanner() {
	if CompileFlagPrintBanner {
		fmt.Printf(Banner())
	}
}

// 44 Chars wide
// 13 Chars "┃  Version: %s┃"

func Banner() string {

	fmt.Println(CompileTimeVersion)

	var spacesBuf string
	for i := len(CompileTimeVersion); i < (44 - 13); i++ {
		spacesBuf = fmt.Sprintf("%s ", spacesBuf)
	}

	var str string
	str += fmt.Sprintf("\n")
	str += fmt.Sprintf("┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n")
	str += fmt.Sprintf("┃                                          ┃\n")
	str += fmt.Sprintf("┃ ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗ ┃\n")
	str += fmt.Sprintf("┃ ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝ ┃\n")
	str += fmt.Sprintf("┃    ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝  ┃\n")
	str += fmt.Sprintf("┃    ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗  ┃\n")
	str += fmt.Sprintf("┃    ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗ ┃\n")
	str += fmt.Sprintf("┃    ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝ ┃\n")
	str += fmt.Sprintf("┃    A live streaming command line tool.   ┃\n")
	str += fmt.Sprintf("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n")
	str += fmt.Sprintf("┃ Author  : Kris Nóva <kris@nivenly.com>   ┃\n")
	str += fmt.Sprintf("┃ Version : %s%s┃\n", CompileTimeVersion, spacesBuf)
	str += fmt.Sprintf("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n")
	str += fmt.Sprintf("\n")
	str += fmt.Sprintf("\n")

	return str
}
