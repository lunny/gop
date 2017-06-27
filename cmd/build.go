// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	config.Name = filepath.Base(wd)
	if err = loadConfig(filepath.Join(wd, "gop.yml")); err != nil {
		return err
	}

	var args = ctx.Args()
	var find = -1
	for i, arg := range args {
		if arg == "-o" {
			find = i
			break
		}
	}

	if find > -1 {
		if find < len(args)-1 {
			config.Name = args[find+1]
		} else {
			args = append(args[:find], "-o", config.Name)
		}
	} else {
		args = append(args, "-o", config.Name)
	}

	cmd := NewCommand("build").AddArguments(args...)
	envs := os.Environ()
	var gopathIdx = -1
	for i, env := range envs {
		if strings.HasPrefix(env, "GOPATH=") {
			gopathIdx = i
			break
		}
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

	return nil
}
