// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
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

type (
	GlobalConfig struct {
		Init struct {
			DefaultEditor string `yaml:"default_editor"`
		} `yaml:"init"`

		Repos struct {
			DefaultDir string `yaml:"default_dir"`
		} `yaml:"repos"`

		Sources map[string]Source `yaml:"sources"`
	}

	Source struct {
		UrlPrefix string `yaml:"url_prefix"`
		PkgPrefix string `yaml:"pkg_prefix"`
	}
)

func snakeCasedName(name string) string {
	newstr := make([]rune, 0)
	for idx, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if idx > 0 {
				newstr = append(newstr, '_')
			}
			chr -= ('A' - 'a')
		}
		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func (g *GlobalConfig) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	keys := strings.Split(key, ".")
	v := reflect.ValueOf(globalConfig).FieldByNameFunc(func(field string) bool {
		return snakeCasedName(field) == keys[0]
	})
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return ""
	}

	if len(keys) == 1 {
		return fmt.Sprintf("%v", v.Interface())
	} else if len(keys) == 2 {
		if v.Type().Kind() == reflect.Struct {
			vv := v.FieldByNameFunc(func(field string) bool {
				return snakeCasedName(field) == keys[1]
			})
			if vv.IsValid() {
				return fmt.Sprintf("%v", vv.Interface())
			}
		} else if v.Type().Kind() == reflect.Map {
			vv := v.MapIndex(reflect.ValueOf(keys[1]))
			if !vv.IsNil() && vv.IsValid() {
				return fmt.Sprintf("%v", vv.Interface())
			}
		}
	}

	return ""
}

func (g *GlobalConfig) Set(key, value string) {
	if len(key) == 0 {
		return
	}
	keys := strings.Split(key, ".")
	v := reflect.ValueOf(globalConfig).FieldByNameFunc(func(field string) bool {
		return snakeCasedName(field) == keys[0]
	})
	if v.IsNil() || !v.IsValid() {
		return
	}

	if len(keys) == 1 {
		v.Set(reflect.ValueOf(value))
	} else if len(keys) == 2 {
		if v.Type().Kind() == reflect.Struct {
			vv := v.FieldByNameFunc(func(field string) bool {
				return snakeCasedName(field) == keys[1]
			})
			if !vv.IsNil() && vv.IsValid() {
				vv.Set(reflect.ValueOf(value))
				return
			}
		} else if v.Type().Kind() == reflect.Map {
			v.SetMapIndex(reflect.ValueOf(keys[1]), reflect.ValueOf(value))
			return
		}
	} else if len(keys) == 3 {
		if v.Type().Kind() == reflect.Struct {
			vv := v.FieldByNameFunc(func(field string) bool {
				return snakeCasedName(field) == keys[1]
			})
			if !vv.IsNil() && vv.IsValid() {
				if vv.Type().Kind() == reflect.Struct {
					vvv := vv.FieldByNameFunc(func(field string) bool {
						return snakeCasedName(field) == keys[2]
					})
					if !vvv.IsNil() && vvv.IsValid() {
						vvv.Set(reflect.ValueOf(value))
						return
					}
				} else if vv.Type().Kind() == reflect.Map {
					vv.SetMapIndex(reflect.ValueOf(keys[2]), reflect.ValueOf(value))
				}
				return
			}
		} else if v.Type().Kind() == reflect.Map {
			vv := v.MapIndex(reflect.ValueOf(keys[1]))
			if !vv.IsValid() {
				//v.SetMapIndex(reflect.ValueOf(keys[1]), reflect.New(v.).Elem())
				vv = v.MapIndex(reflect.ValueOf(keys[1]))
			}
			if !vv.IsNil() && vv.IsValid() {
				vv.SetMapIndex(reflect.ValueOf(keys[2]), reflect.ValueOf(value))
				return
			}
		}
	}
}

var (
	globalConfig = GlobalConfig{
		Sources: map[string]Source{
			"github": {
				UrlPrefix: "https://github.com",
				PkgPrefix: "github.com",
			},
		},
	}
)

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

	if globalConfig.Repos.DefaultDir == "" {
		homeDir, err := Home()
		if err != nil {
			return err
		}
		globalConfig.Repos.DefaultDir = filepath.Join(homeDir, ".gop", "repos")
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
		bs, err := yaml.Marshal(&globalConfig)
		if err != nil {
			return err
		}
		fmt.Println(string(bs))
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
