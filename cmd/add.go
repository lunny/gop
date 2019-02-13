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

	"github.com/mholt/archiver"
	"github.com/urfave/cli"
)

// CmdAdd represents add a new dependency package and it's dependencies to this project
var CmdAdd = cli.Command{
	Name:        "add",
	Usage:       "Add one or more new dependency packages",
	Description: `Add one or more new dependency packages`,
	Action:      runAdd,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enables verbose progress and debug output",
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
		return strings.HasPrefix(path, ".git/") || strings.HasPrefix(path, ".hg/") ||
			strings.HasPrefix(path, "vendor/") ||
			(!includeTest && strings.HasSuffix(path, "_test.go"))
	})
}

// CopyPkg copy package from sources
func CopyPkg(globalGoPath, pkg, dstPath string, includeTest bool) error {
	_, err := copyPkgFromGlobalGoPath(globalGoPath, pkg, dstPath, includeTest)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	/*if err == nil && copied {
		return nil
	}

	_, err = copyPkgFromCache(pkg, dstPath, includeTest)
	*/return err
}

func copyPkgFromGlobalGoPath(globalGoPath, pkg, dstPath string, includeTest bool) (bool, error) {
	absPkgPath := filepath.Join(globalGoPath, "src", pkg)
	_, err := os.Stat(absPkgPath)
	if err != nil {
		return false, err
	}

	exist, err := isPkgExist(dstPath)
	if err != nil {
		return false, err
	}
	if !exist {
		fmt.Println("Copying", pkg)
		err = copyPkg(absPkgPath, dstPath, includeTest)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func copyPkgFromCache(pkg, dstPath string, includeTest bool) (bool, error) {
	absPkgPath := filepath.Join(globalConfig.Repos.DefaultDir, pkg, "master.zip")
	_, err := os.Stat(absPkgPath)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(dstPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}

		fmt.Println("Copying", pkg)
		err = archiver.Zip.Open(absPkgPath, dstPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// add add one package to vendor
func add(ctx *cli.Context, name, projPath, globalGoPath string) error {
	if strings.HasPrefix(name, "../") || filepath.IsAbs(name) || strings.HasPrefix(name, "./") {
		return errors.New("relative pkg and absolute pkg is not supported, only packages on GOPATH")
	}

	showLog = ctx.IsSet("verbose")

	dstPath := filepath.Join(projPath, "src", "vendor", name)
	err := CopyPkg(globalGoPath, name, dstPath, ctx.Bool("test"))
	if err != nil {
		return err
	}

	imports, err := ListImports(globalGoPath, name, projPath, dstPath, ctx.String("tags"), ctx.Bool("test"))
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

		if err := add(ctx, imp.Name, projPath, globalGoPath); err != nil {
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

	homeDir, err := Home()
	if err != nil {
		return err
	}

	err = loadGlobalConfig(filepath.Join(homeDir, ".gop.yml"))
	if err != nil {
		return err
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
