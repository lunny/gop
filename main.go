// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/lunny/gop/cmd"

	"github.com/urfave/cli"
)

const (
	Version = "0.1"
)

func main() {
	app := cli.NewApp()
	app.Name = "Gop"
	app.Usage = "A golang project manangement tool for executable projects"
	app.Version = Version
	app.Commands = []cli.Command{
		cmd.CmdInit,
		cmd.CmdBuild,
		cmd.CmdEnsure,
		cmd.CmdTest,
		cmd.CmdStatus,
		cmd.CmdAdd,
		cmd.CmdRemove,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(4, "Failed to run app with %s: %v", os.Args, err)
	}
}
