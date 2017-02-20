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

// CmdAdd represents add a new dependency package and it's dependencies to this project
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

func add(name, projPath, globalGoPath string, isUpdate bool) error {
	var absPkgPath, dstPath, parentPkg string

	if strings.HasPrefix(name, "../") || filepath.IsAbs(name) {
		parentPkg = name
		var err error
		absPkgPath, err = filepath.Abs(name)
		if err != nil {
			return err
		}
		parentPkg = filepath.Base(absPkgPath)
	} else {
		parentPkg, _ = util.NormalizeName(name)
		absPkgPath = filepath.Join(globalGoPath, "src", parentPkg)
	}

	dstPath = filepath.Join(projPath, "src", parentPkg)

	info, err := os.Stat(dstPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		fmt.Println("copying", parentPkg)
		err = CopyDir(absPkgPath, dstPath, func(path string) bool {
			return strings.HasPrefix(path, ".git")
		})
		if err != nil {
			return err
		}
	} else if isUpdate {
		if !info.IsDir() {
			return fmt.Errorf("Dest dir %s is a file", dstPath)
		}

		fmt.Println("copying", parentPkg)
		os.RemoveAll(dstPath)
		err = CopyDir(absPkgPath, dstPath, func(path string) bool {
			return strings.HasPrefix(path, ".git")
		})
		if err != nil {
			return err
		}
	}

	imports, err := ListImports(parentPkg, absPkgPath, absPkgPath, "", false)
	if err != nil {
		return err
	}

	for i, imp := range imports {
		var has bool
		for j := 0; j < i; j++ {
			if imports[j] == imp {
				has = true
				break
			}
		}
		if has {
			continue
		}

		if err := add(imp, projPath, globalGoPath, isUpdate); err != nil {
			return err
		}
	}
	return nil
}

func runAdd(ctx *cli.Context) error {
	if len(ctx.Args()) <= 0 {
		return errors.New("You have to indicate more than one package")
	}

	globalGoPath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("Not found GOPATH")
	}

	names := ctx.Args()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, name := range names {
		if err = add(name, wd, globalGoPath, ctx.IsSet("update")); err != nil {
			return err
		}
	}
	return nil
}
