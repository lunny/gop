// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unknwon/com"
	"github.com/go-yaml/yaml"
	"github.com/urfave/cli"
)

// CmdRelease represents
var CmdRelease = cli.Command{
	Name:            "release",
	Usage:           "Release this project",
	Description:     `Release this project`,
	Action:          runRelease,
	SkipFlagParsing: true,
}

type Config struct {
	Name   string
	Assets []string
}

var config Config

func loadConfig(ymlPath string) error {
	if com.IsExist(ymlPath) {
		Println("find config file", ymlPath)
		bs, err := ioutil.ReadFile(ymlPath)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(bs, &config)
		if err != nil {
			return err
		}
	}
	return nil
}

func runRelease(ctx *cli.Context) error {
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
		if arg == "-v" {
			showLog = true
			continue
		}
		if arg == "-o" {
			find = i
			break
		}
	}

	if find > -1 {
		if find < len(args)-2 {
			args = append(args[:find], args[find+2:]...)
		} else {
			args = args[:find]
		}
	}

	args = append(args, "-o", filepath.Join(wd, "bin", config.Name))
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

	Println(cmd)

	err = cmd.RunInDirPipeline("src", os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	for _, asset := range config.Assets {
		srcPath := filepath.Join(wd, "src", asset)
		dstPath := filepath.Join(wd, "bin", asset)
		if com.IsDir(srcPath) {
			os.RemoveAll(dstPath)
			err = com.CopyDir(srcPath, dstPath)
			if err != nil {
				Errorf("copy dir %s to %s failed: %v\n", srcPath, dstPath, err)
			}
		} else if com.IsFile(srcPath) {
			os.RemoveAll(dstPath)
			err = com.Copy(srcPath, dstPath)
			if err != nil {
				Errorf("copy file %s to %s failed: %v\n", srcPath, dstPath, err)
			}
		}
	}

	return nil
}
