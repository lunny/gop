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

	"github.com/lunny/gop/util"
	"github.com/urfave/cli"
)

// CmdInit represents
var CmdAdd = cli.Command{
	Name:        "add",
	Usage:       "add a new dependency",
	Description: `add a new dependency`,
	Action:      runAdd,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "update, u",
			Usage: "update the dependency package",
		},
	},
}

func runAdd(ctx *cli.Context) error {
	if len(ctx.Args()) <= 0 {
		return errors.New("You have to indicate more than one package")
	}

	globalGoPath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("Not found GOPATH")
	}

	parentPkg, _ := util.NormalizeName(ctx.Args()[0])
	absPkgPath := filepath.Join(globalGoPath, "src", parentPkg)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	dstPath := filepath.Join(wd, "src", parentPkg)
	info, err := os.Stat(dstPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("copying", parentPkg)
			return CopyDir(absPkgPath, dstPath, func(path string) bool {
				return strings.HasPrefix(path, ".git")
			})
		}
		return err
	}

	if ctx.IsSet("update") {
		if !info.IsDir() {
			return fmt.Errorf("Dest dir %s is a file", dstPath)
		}

		fmt.Println("copying", parentPkg)
		os.RemoveAll(dstPath)
		return CopyDir(absPkgPath, dstPath, func(path string) bool {
			return strings.HasPrefix(path, ".git")
		})
	}
	return fmt.Errorf("Pkg %s is added already", parentPkg)
}
