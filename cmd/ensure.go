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

// CmdEnsure represents
var CmdEnsure = cli.Command{
	Name:        "ensure",
	Usage:       "Ensure all the dependent packages installed accroding target",
	Description: `Ensure all the dependent packages installed accroding target`,
	Action:      runEnsure,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enables verbose progress and debug output",
		},
		cli.BoolFlag{
			Name:  "dry, d",
			Usage: "Dry run, print what would be done",
		},
		cli.BoolFlag{
			Name:  "get, g",
			Usage: "call go get to download the package if package is not in GOPATH",
		},
		cli.BoolFlag{
			Name:  "update, u",
			Usage: "call go get -u to update the package if package is exist in GOPATH",
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

var updatedPackage = make(map[string]struct{})

func ensure(ctx *cli.Context, globalGoPath, projectRoot string, target *Target, isTest bool) error {
	vendorDir := filepath.Join(projectRoot, "src", "vendor")
	imports, err := ListImports(projectRoot, target.Dir, projectRoot, filepath.Join(projectRoot, "src"), ctx.String("tags"), isTest)
	if err != nil {
		return err
	}
	for _, imp := range imports {
		if imp.Type == PkgTypeProjectGoPath || imp.Type == PkgTypeGoRoot {
			continue
		}
		if imp.Name == "C" || strings.HasPrefix(imp.Name, "../") || strings.HasPrefix(imp.Name, "./") {
			continue
		}

		// package dir
		srcDir := filepath.Join(globalGoPath, "src", imp.Name)
		// FIXME: dry will lost some packages with -g or -u
		if ctx.IsSet("dry") {
			fmt.Println("Dry copying", imp.Name)
			continue
		}

		// FIXME: imp only UNIX
		dstDir := filepath.Join(vendorDir, imp.Name)
		if ctx.IsSet("update") {
			if _, ok := updatedPackage[imp.Name]; ok {
				continue
			}

			fmt.Println("Downloading", imp.Name)
			cmdGet := NewCommand("get").AddArguments("-u", imp.Name)
			err = cmdGet.RunInDirPipeline("src", os.Stdout, os.Stderr)
			if err != nil {
				return err
			}

			os.RemoveAll(dstDir)
			err = CopyPkg(globalGoPath, imp.Name, dstDir, ctx.Bool("test"))
			if err != nil {
				return err
			}

			updatedPackage[imp.Name] = struct{}{}

			return ensure(ctx, globalGoPath, projectRoot, target, isTest)
		}

		exist, err := isDirExist(dstDir)
		if err != nil {
			return err
		}

		if !exist {
			exist, err = isDirExist(srcDir)
			if err != nil {
				return err
			}
			if !exist {
				if ctx.IsSet("get") {
					fmt.Println("Downloading", imp.Name)
					cmdGet := NewCommand("get").AddArguments(imp.Name)
					err = cmdGet.RunInDirPipeline(filepath.Join(projectRoot, "src"), os.Stdout, os.Stderr)
					if err != nil {
						err = download(ctx, imp.Name)
						if err != nil {
							return err
						}
					}

					// scan the package dependencies again since the new package added
					return ensure(ctx, globalGoPath, projectRoot, target, isTest)
				}

				fmt.Printf("Package %s not found on $GOPATH, please use -g option or go get at first\n", imp.Name)
				return nil
			}

			err = CopyPkg(globalGoPath, imp.Name, dstDir, ctx.Bool("test"))
			if err != nil {
				return err
			}

			return ensure(ctx, globalGoPath, projectRoot, target, isTest)
		}
	}
	return nil
}

func runEnsure(ctx *cli.Context) error {
	globalGoPath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("Not found GOPATH")
	}

	showLog = ctx.IsSet("verbose")

	level, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	if err = loadConfig(filepath.Join(projectRoot, "gop.yml")); err != nil {
		return err
	}

	var args = ctx.Args()
	var targetName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		targetName = args[0]
	}

	if err = analysisTarget(level, targetName, projectRoot); err != nil {
		return err
	}

	return ensure(ctx, globalGoPath, projectRoot, curTarget, ctx.Bool("test"))
}
