// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

// CmdTest represents
var CmdTest = cli.Command{
	Name:            "test",
	Usage:           "Run the target test codes",
	Description:     `Run the target test codes`,
	Action:          runTest,
	SkipFlagParsing: true,
}

func runTest(ctx *cli.Context) error {
	var args = ctx.Args()
	var ensureFlagIdx = -1
	for i, arg := range args {
		if arg == "-v" {
			showLog = true
		} else if arg == "-e" {
			ensureFlagIdx = i
		}
	}

	if ensureFlagIdx > -1 {
		args = append(args[:ensureFlagIdx], args[ensureFlagIdx+1:]...)
	}

	envs := os.Environ()
	var gopathIdx = -1
	for i, env := range envs {
		if strings.HasPrefix(env, "GOPATH=") {
			gopathIdx = i
			break
		}
	}

	level, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	if err = loadConfig(filepath.Join(projectRoot, "gop.yml")); err != nil {
		return err
	}

	var targetName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		targetName = args[0]
		args = args[1:]
	}

	if err = analysisTarget(level, targetName, projectRoot); err != nil {
		return err
	}

	if ensureFlagIdx > -1 {
		globalGoPath, ok := os.LookupEnv("GOPATH")
		if !ok {
			return errors.New("Not found GOPATH")
		}

		if err = ensure(ctx, globalGoPath, projectRoot, curTarget, true); err != nil {
			return err
		}
	}

	newGopath := fmt.Sprintf("GOPATH=%s", projectRoot)
	if gopathIdx > 0 {
		envs[gopathIdx] = newGopath
	} else {
		envs = append(envs, newGopath)
	}

	cmd := NewCommand("test").AddArguments(args...)
	cmd.Env = envs
	err = cmd.RunInDirPipeline(filepath.Join(projectRoot, "src", curTarget.Dir), os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	return nil
}
