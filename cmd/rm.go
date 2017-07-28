// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lunny/gop/util"
	"github.com/urfave/cli"
)

// CmdInit represents
var CmdRemove = cli.Command{
	Name:        "rm",
	Usage:       "remove a dependency",
	Description: `remove a dependency`,
	Action:      runRemove,
	/*Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "force, -f",
			Value: "",
			Usage: "rm",
		},
	},*/
}

func runRemove(ctx *cli.Context) error {
	if len(ctx.Args()) <= 0 {
		return errors.New("You have to indicate more than one package")
	}

	_, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	parentPkg, _ := util.NormalizeName(ctx.Args()[0])

	dstPath := filepath.Join(projectRoot, "src", "vendor", parentPkg)
	fmt.Println("removing", parentPkg)
	os.RemoveAll(dstPath)

	return nil
}
