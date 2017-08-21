// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Unknwon/com"

	"github.com/lunny/gop/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func runCommand(cmd cli.Command, dir string, args ...string) error {
	if dir != "" {
		os.Chdir(dir)
	}

	app := cli.NewApp()
	app.Name = "Gop"
	app.Usage = "Build golang applications out of GOPATH"
	app.Version = Version
	app.Commands = []cli.Command{
		cmd,
	}

	var cmdArgs = []string{"gop", cmd.Name}

	return app.Run(append(cmdArgs, args...))
}

func TestInit(t *testing.T) {
	tmpDir := os.TempDir()
	err := runCommand(cmd.CmdInit, tmpDir)
	assert.NoError(t, err)
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "main")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "main", "main.go")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "gop.yml")))

	err = runCommand(cmd.CmdBuild, filepath.Join(tmpDir, "src"))
	assert.NoError(t, err)

	err = runCommand(cmd.CmdEnsure, filepath.Join(tmpDir, "src"))
	assert.NoError(t, err)

	err = runCommand(cmd.CmdStatus, filepath.Join(tmpDir, "src"))
	assert.NoError(t, err)
}
