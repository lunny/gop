// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// CmdBuild represents
var CmdBuild = cli.Command{
	Name:            "build",
	Usage:           "Build this project",
	Description:     `Build this project`,
	Action:          runBuild,
	SkipFlagParsing: true,
}

func runBuild(ctx *cli.Context) error {
	cmd := NewCommand("build").AddArguments(ctx.Args()...)
	envs := os.Environ()
	var gopathIdx = -1
	for i, env := range envs {
		if strings.HasPrefix(env, "GOPATH=") {
			gopathIdx = i
			break
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	newGopath := fmt.Sprintf("GOPATH=%s", wd)
	if gopathIdx > 0 {
		envs[gopathIdx] = newGopath
	} else {
		envs = append(envs, newGopath)
	}
	cmd.Env = envs

	err = cmd.RunInDirPipeline("src", os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	// FIXME: move the build binary to bin/

	return nil
}
