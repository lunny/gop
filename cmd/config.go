// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// Target build target
type Target struct {
	Name   string
	Dir    string
	Assets []string
}

// Config gop.yml
type Config struct {
	Targets []Target
}

var config Config

func loadConfig(ymlPath string) error {
	exist, _ := isFileExist(ymlPath)
	if exist {
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

	if len(config.Targets) == 0 {
		projectName := filepath.Base(filepath.Dir(ymlPath))
		config.Targets = []Target{
			{
				Name: projectName,
				Dir:  "main",
			},
		}
	}
	return nil
}

const (
	dirLevelOutProject = iota // command run out of project root
	dirLevelRoot              // command run in project root
	dirLevelSrc               // command run in <project root>/src
	dirLevelTarget            // command run in <project root>/src/<target>
)

func analysisDirLevel() (int, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 0, "", err
	}

	wd, err = filepath.Abs(wd)
	if err != nil {
		return 0, "", err
	}

	exist, _ := isFileExist(filepath.Join(wd, "gop.yml"))
	if exist {
		return dirLevelRoot, wd, nil
	} else if filepath.Base(wd) == "src" {
		exist, _ = isFileExist(filepath.Join(filepath.Dir(wd), "gop.yml"))
		if exist {
			return dirLevelSrc, filepath.Dir(wd), nil
		}
	}

	srcDir := filepath.Dir(wd)
	for filepath.Base(srcDir) != "src" {
		srcDir = filepath.Dir(srcDir)
		if srcDir == "" || srcDir == "/" {
			break
		}
	}

	exist, _ = isFileExist(filepath.Join(filepath.Dir(srcDir), "gop.yml"))
	if srcDir == "" || !exist {
		return dirLevelOutProject, "", errors.New("unknow directory to run gop")
	}

	return dirLevelTarget, filepath.Dir(srcDir), nil
}
