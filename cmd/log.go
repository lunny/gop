// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import "fmt"

var (
	showLog bool
)

func Println(a ...interface{}) {
	if showLog {
		fmt.Println(a...)
	}
}

func Printf(format string, a ...interface{}) {
	if showLog {
		fmt.Printf(format, a...)
	}
}

func Error(a ...interface{}) {
	fmt.Println(a...)
}

func Errorf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}
