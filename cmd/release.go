// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Unknwon/com"
	"github.com/urfave/cli"
)

// CmdRelease represents
var CmdRelease = cli.Command{
	Name:            "release",
	Usage:           "Release the target according the gop.yml",
	Description:     `Release the target according the gop.yml`,
	Action:          runRelease,
	SkipFlagParsing: true,
}

func runRelease(ctx *cli.Context) error {
	_, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	if err = loadConfig(filepath.Join(projectRoot, "gop.yml")); err != nil {
		return err
	}

	var target = config.Targets[0]
	var args = ctx.Args()
	var find = -1
	for i, arg := range args {
		if arg == "-v" {
			showLog = true
		} else if arg == "-o" {
			find = i
		}
	}

	if find > -1 {
		if find < len(args)-2 {
			args = append(args[:find], args[find+2:]...)
		} else {
			args = args[:find]
		}
	}

	var ext string
	if os.Getenv("GOOS") == "windows" ||
		(os.Getenv("GOOS") == "" && runtime.GOOS == "windows") {
		ext = ".exe"
	}

	args = append(args, "-o", filepath.Join(projectRoot, "bin", target.Name, target.Name+ext))
	cmd := NewCommand("build").AddArguments(args...)
	envs := os.Environ()
	var gopathIdx = -1
	for i, env := range envs {
		if strings.HasPrefix(env, "GOPATH=") {
			gopathIdx = i
			break
		}
	}

	newGopath := fmt.Sprintf("GOPATH=%s", projectRoot)
	if gopathIdx > 0 {
		envs[gopathIdx] = newGopath
	} else {
		envs = append(envs, newGopath)
	}
	cmd.Env = envs

	err = cmd.RunInDirPipeline(filepath.Join(projectRoot, "src", target.Dir), os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	for _, asset := range target.Assets {
		srcPath := filepath.Join(projectRoot, "src", target.Dir, asset)
		dstPath := filepath.Join(projectRoot, "bin", target.Name, asset)
		exist, _ := isDirExist(srcPath)
		fileExist, _ := isFileExist(srcPath)
		if exist {
			os.RemoveAll(dstPath)
			err = com.CopyDir(srcPath, dstPath)
			if err != nil {
				Errorf("copy dir %s to %s failed: %v\n", srcPath, dstPath, err)
			}
		} else if fileExist {
			os.RemoveAll(dstPath)
			err = com.Copy(srcPath, dstPath)
			if err != nil {
				Errorf("copy file %s to %s failed: %v\n", srcPath, dstPath, err)
			}
		}
	}

	return nil
}
