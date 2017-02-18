// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

const (
	defaultMainFile = `
package main

func main() {
		
}
`
)

// CmdInit represents
var CmdInit = cli.Command{
	Name:        "init",
	Usage:       "Init a new project",
	Description: `Init a new project`,
	Action:      runInit,
}

func runInit(ctx *cli.Context) error {
	os.MkdirAll("src", os.ModePerm)
	os.MkdirAll("bin", os.ModePerm)

	mainFile := filepath.Join("src", "main.go")
	_, err := os.Stat(mainFile)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(mainFile)
			if err != nil {
				return fmt.Errorf("os.Create: %v", err)
			}
			defer f.Close()
			_, err = f.Write([]byte(defaultMainFile))
			if err != nil {
				return fmt.Errorf("create main file failed: %v", err)
			}
			return nil
		}
		return fmt.Errorf("os.State: %v", err)
	}

	return nil
}
