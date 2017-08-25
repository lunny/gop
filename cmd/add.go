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
		cli.BoolFlag{
			Name:  "test, t",
			Usage: "include test files",
		},
		cli.StringFlag{
			Name:  "tags",
			Usage: "tags for import package find",
		},
	},
}

func copyPkg(srcPkgPath, dstPkgPath string, includeTest bool) error {
	return CopyDir(srcPkgPath, dstPkgPath, func(path string) bool {
		return strings.HasPrefix(path, filepath.Join(dstPkgPath, ".git")) ||
			strings.HasPrefix(path, filepath.Join(dstPkgPath, "vendor")) ||
			(!includeTest && strings.HasSuffix(path, "_test.go"))
	})
}

// add add one package to vendor
func add(ctx *cli.Context, name, projPath, globalGoPath string) error {
	if strings.HasPrefix(name, "../") || filepath.IsAbs(name) || strings.HasPrefix(name, "./") {
		return errors.New("relative pkg and absolute pkg is not supported, only packages on GOPATH")
	}

	absPkgPath := filepath.Join(globalGoPath, "src", name)
	dstPath := filepath.Join(projPath, "src", "vendor", name)

	_, err := os.Stat(absPkgPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		fmt.Println("Copying", name)
		err = copyPkg(absPkgPath, dstPath, ctx.Bool("test"))
		if err != nil {
			return err
		}
	} else if ctx.IsSet("update") {
		if !info.IsDir() {
			return fmt.Errorf("Dest dir %s is a file", dstPath)
		}

		fmt.Println("Copying", name)
		os.RemoveAll(dstPath)
		err = copyPkg(absPkgPath, dstPath, ctx.Bool("test"))
		if err != nil {
			return err
		}
	} else {
		return nil
	}

	imports, err := ListImports(name, absPkgPath, absPkgPath, ctx.String("tags"), ctx.Bool("test"))
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

		if err := add(ctx, imp, projPath, globalGoPath); err != nil {
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

	_, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	for _, name := range names {
		if err = add(ctx, name, projectRoot, globalGoPath); err != nil {
			return err
		}
	}
	return nil
}
