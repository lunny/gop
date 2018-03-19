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
	defaultMainFile = `package main

func main() {
		
}
`

	defaultYaml = `targets:
- name: %s
  dir: main
  assets:
  - templates
  - public
`

	defaultVscodeCfgFile = `{
	"go.gopath": "${workspaceRoot}"
}`
)

var (
	// Debug indicated whether it is debug mode
	Debug = false
)

// CmdInit represents
var CmdInit = cli.Command{
	Name:        "init",
	Usage:       "Init a new project",
	Description: `Init a new project`,
	Action:      runInit,
	Flags: []cli.Flag{
		cli.BoolTFlag{
			Name:  "verbose, v",
			Usage: "Enables verbose progress and debug output",
		},
		cli.StringFlag{
			Name:  "editor, e",
			Usage: "Generate specifial editor configuration. Could be vscode or blank with no editor support",
		},
	},
}

func runInit(ctx *cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	homeDir, err := Home()
	if err != nil {
		return err
	}

	err = loadGlobalConfig(filepath.Join(homeDir, ".gop.yml"))
	if err != nil {
		return err
	}

	showLog = ctx.IsSet("verbose")

	os.MkdirAll(filepath.Join(wd, "src"), os.ModePerm)
	os.MkdirAll(filepath.Join(wd, "src", "vendor"), os.ModePerm)
	os.MkdirAll(filepath.Join(wd, "src", "main"), os.ModePerm)
	os.MkdirAll(filepath.Join(wd, "bin"), os.ModePerm)

	ymlPath := filepath.Join(wd, "gop.yml")
	_, err = os.Stat(ymlPath)
	if err != nil {
		if os.IsNotExist(err) {
			y, err := os.Create(ymlPath)
			if err != nil {
				return err
			}
			defer y.Close()

			_, err = y.Write([]byte(fmt.Sprintf(defaultYaml, filepath.Base(wd))))
			if err != nil {
				return err
			}
		}
	}

	mainFile := filepath.Join(wd, "src", "main", "main.go")
	_, err = os.Stat(mainFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("os.State: %v", err)
		}

		f, err := os.Create(mainFile)
		if err != nil {
			return fmt.Errorf("os.Create: %v", err)
		}
		defer f.Close()
		_, err = f.Write([]byte(defaultMainFile))
		if err != nil {
			return fmt.Errorf("create main file failed: %v", err)
		}
	}

	var editor = globalConfig.Get("init.default_editor")
	if ctx.IsSet("editor") {
		editor = ctx.String("editor")
	}

	switch editor {
	case "vscode":
		os.MkdirAll(filepath.Join(wd, ".vscode"), os.ModePerm)
		cfgFile := filepath.Join(wd, ".vscode", "settings.json")

		_, err = os.Stat(cfgFile)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("os.State: %v", err)
			}

			f, err := os.Create(cfgFile)
			if err != nil {
				return fmt.Errorf("os.Create: %v", err)
			}
			defer f.Close()
			_, err = f.Write([]byte(defaultVscodeCfgFile))
			if err != nil {
				return fmt.Errorf("create main file failed: %v", err)
			}
		}
	}

	return nil
}
