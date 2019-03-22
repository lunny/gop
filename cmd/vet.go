// Copyright 2019 The Gop Authors. All rights reserved.
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

// CmdVet represents a vet command
var CmdVet = cli.Command{
	Name:        "vet",
	Usage:       "Vet",
	Description: `Vet`,
	Action:      runVet,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enables verbose progress and debug output",
		},
	},
}

func runVet(ctx *cli.Context) error {
	var args = ctx.Args()
	for _, arg := range args {
		if arg == "-v" {
			showLog = true
		}
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
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") && !strings.Contains(args[0], "/") {
		targetName = args[0]
		args = args[1:]
	}

	if err = analysisTarget(level, targetName, projectRoot); err != nil {
		return err
	}

	newGopath := fmt.Sprintf("GOPATH=%s", projectRoot)
	if gopathIdx > 0 {
		envs[gopathIdx] = newGopath
	} else {
		envs = append(envs, newGopath)
	}

	cmd := NewCommand("vet").AddArguments(args...)
	cmd.Env = envs
	err = cmd.RunInDirPipeline(filepath.Join(projectRoot, "src", curTarget.Dir), os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	return nil
}
