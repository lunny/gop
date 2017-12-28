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

// CmdUpdate represents update a new dependency package and it's dependencies to this project
var CmdUpdate = cli.Command{
	Name:        "update",
	Usage:       "update a new dependency",
	Description: `update a new dependency`,
	Action:      runUpdate,
	Flags: []cli.Flag{
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

// update update one package to vendor
func update(ctx *cli.Context, name, projPath, globalGoPath string) error {
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
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("Dest dir %s is a file", dstPath)
	}

	fmt.Println("Copying", name)
	os.RemoveAll(dstPath)
	err = copyPkg(absPkgPath, dstPath, ctx.Bool("test"))
	if err != nil {
		return err
	}

	imports, err := ListImports(globalGoPath, name, projPath, absPkgPath, ctx.String("tags"), ctx.Bool("test"))
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

		if err := update(ctx, imp.Name, projPath, globalGoPath); err != nil {
			return err
		}
	}
	return nil
}

func runUpdate(ctx *cli.Context) error {
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
		if err = update(ctx, name, projectRoot, globalGoPath); err != nil {
			return err
		}
	}
	return nil
}
