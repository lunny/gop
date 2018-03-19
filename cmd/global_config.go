// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

// CmdConfig represents config the gop global options
var CmdConfig = cli.Command{
	Name:        "config",
	Usage:       "Config global options",
	Description: `Config global options`,
	Subcommands: []cli.Command{
		{
			Name:        "get",
			Usage:       "Get global config options",
			Description: `Get global config options`,
			Action:      runConfigGet,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Show all the config options",
				},
			},
		},
		{
			Name:        "set",
			Usage:       "Set global config options",
			Description: `Set global config options`,
			Action:      runConfigSet,
		},
	},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enables verbose progress and debug output",
		},
	},
}

type GlobalConfig map[string]interface{}

var (
	globalConfig = GlobalConfig{
		"init": map[interface{}]interface{}{
			"default_editor": "",
		},
		"repos": map[interface{}]interface{}{
			"default_dir": "",
		},
	}
)

func (c GlobalConfig) Get(key string) string {
	keys := strings.Split(key, ".")
	if len(keys) == 1 {
		if c[keys[0]] != nil {
			return c[keys[0]].(string)
		}
	} else if len(keys) == 2 {
		if c[keys[0]] == nil {
			c[keys[0]] = make(map[interface{}]interface{})
		}
		v := c[keys[0]].(map[interface{}]interface{})
		if v[keys[1]] != nil {
			return v[keys[1]].(string)
		}
	}
	return ""
}

func (c GlobalConfig) Set(key, value string) {
	keys := strings.Split(key, ".")
	if len(keys) == 1 {
		c[keys[0]] = value
	} else if len(keys) == 2 {
		if c[keys[0]] == nil {
			c[keys[0]] = make(map[interface{}]interface{})
		}
		v := c[keys[0]].(map[interface{}]interface{})
		v[interface{}(keys[1])] = value
	}
}

func (c GlobalConfig) Println() {
	bs, err := yaml.Marshal(&globalConfig)
	if err == nil {
		fmt.Println(string(bs))
	}
}

func loadGlobalConfig(ymlPath string) error {
	exist, _ := isFileExist(ymlPath)
	if exist {
		Println("Found global config file", ymlPath)
		bs, err := ioutil.ReadFile(ymlPath)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(bs, &globalConfig)
		if err != nil {
			return err
		}
	}

	if globalConfig.Get("repos.default_dir") == "" {
		homeDir, err := Home()
		if err != nil {
			return err
		}
		globalConfig.Set("repos.default_dir", filepath.Join(homeDir, ".gop", "repos"))
		err = saveGlobalConfig(ymlPath)
		Println("Save default repos config failed:", err)
	}

	return nil
}

func saveGlobalConfig(ymlPath string) error {
	bs, err := yaml.Marshal(&globalConfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ymlPath, bs, 0660)
}

func runConfigGet(ctx *cli.Context) error {
	homeDir, err := Home()
	if err != nil {
		return err
	}

	err = loadGlobalConfig(filepath.Join(homeDir, ".gop.yml"))
	if err != nil {
		return err
	}

	if len(ctx.Args()) <= 0 {
		if !ctx.IsSet("all") {
			return errors.New("You have to indicate more than one package")
		}

		// show all the configs
		globalConfig.Println()
		return nil
	}

	v := globalConfig.Get(ctx.Args()[0])
	fmt.Println(v)

	return nil
}

func runConfigSet(ctx *cli.Context) error {
	if len(ctx.Args()) <= 1 {
		return errors.New("You have to indicate a key and a value")
	}

	homeDir, err := Home()
	if err != nil {
		return err
	}

	ymlPath := filepath.Join(homeDir, ".gop.yml")
	err = loadGlobalConfig(ymlPath)
	if err != nil {
		return err
	}

	globalConfig.Set(ctx.Args()[0], ctx.Args()[1])

	return saveGlobalConfig(ymlPath)
}
