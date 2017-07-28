// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"os"
	"runtime"

	"path/filepath"

	"github.com/urfave/cli"
)

// CmdRun represents
var CmdRun = cli.Command{
	Name:            "run",
	Usage:           "Run this project",
	Description:     `Run this project`,
	Action:          runRun,
	SkipFlagParsing: true,
}

func runRun(ctx *cli.Context) error {
	err := runBuild(ctx)
	if err != nil {
		return err
	}

	_, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	var ext string
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	cmd := Command{
		name: filepath.Join(projectRoot, "src", curTarget.Dir, curTarget.Name+ext),
		Env:  os.Environ(),
	}

	err = cmd.RunInDirPipeline(filepath.Join(projectRoot, "src", curTarget.Dir), os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	return nil
}
