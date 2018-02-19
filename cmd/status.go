// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

// CmdStatus represents
var CmdStatus = cli.Command{
	Name:        "status",
	Usage:       "List the target's dependent packages",
	Description: `List the target's dependent packages`,
	Action:      runStatus,
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

func runStatus(ctx *cli.Context) error {
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

	vendorDir := filepath.Join(projectRoot, "src", "vendor")

	imports, err := ListImports(projectRoot, curTarget.Dir, projectRoot,
		filepath.Join(projectRoot, "src"), ctx.String("tags"), ctx.Bool("test"))
	if err != nil {
		return err
	}
	for i, imp := range imports {
		pkg := filepath.Join(projectRoot, "src", imp.Name)
		exist, _ := isDirExist(pkg)
		if exist {
			continue
		}

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

		// FIXME: imp only UNIX
		p := filepath.Join(vendorDir, imp.Name)
		exist, err := isDirExist(p)
		if err != nil {
			return err
		}
		if exist {
			fmt.Print("[X] ")
		} else {
			fmt.Print("[ ] ")
		}
		fmt.Println(imp.Name)
	}

	return nil
}
