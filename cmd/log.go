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
