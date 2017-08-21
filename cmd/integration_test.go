// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Unknwon/com"
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
	app.Commands = []cli.Command{
		cmd,
	}

	var cmdArgs = []string{"gop", cmd.Name}

	return app.Run(append(cmdArgs, args...))
}

func TestInit(t *testing.T) {
	tmpDir := os.TempDir()
	err := runCommand(CmdInit, tmpDir)
	assert.NoError(t, err)
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "main")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "main", "main.go")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "gop.yml")))

	err = runCommand(CmdBuild, filepath.Join(tmpDir, "src"))
	assert.NoError(t, err)

	err = runCommand(CmdEnsure, filepath.Join(tmpDir, "src"))
	assert.NoError(t, err)

	err = runCommand(CmdStatus, filepath.Join(tmpDir, "src"))
	assert.NoError(t, err)
}

func TestAddAndRm(t *testing.T) {
	tmpDir := os.TempDir()
	err := runCommand(CmdInit, tmpDir)
	assert.NoError(t, err)
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "main")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "main", "main.go")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "gop.yml")))

	cmdGet := NewCommand("get", "github.com/lunny/tango")
	_, err = cmdGet.Run()
	assert.NoError(t, err)

	err = runCommand(CmdAdd, filepath.Join(tmpDir, "src"), "github.com/lunny/tango")
	assert.NoError(t, err)
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny", "tango")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny", "log")))

	err = runCommand(CmdRemove, filepath.Join(tmpDir, "src"), "github.com/lunny/tango")
	assert.NoError(t, err)

	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny")))
	assert.False(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny", "tango")))
	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny", "log")))

	err = runCommand(CmdRemove, filepath.Join(tmpDir, "src"), "github.com/lunny/log")
	assert.NoError(t, err)

	assert.True(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny")))
	assert.False(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny", "tango")))
	assert.False(t, com.IsExist(filepath.Join(tmpDir, "src", "vendor", "github.com", "lunny", "log")))
}
