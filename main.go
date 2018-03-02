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
	// Version of gop
	Version = "0.6.0302"
)

func main() {
	app := cli.NewApp()
	app.Name = "gop"
	app.Usage = "Build golang applications out of GOPATH"
	app.Version = Version
	app.Commands = []cli.Command{
		cmd.CmdInit,
		cmd.CmdBuild,
		cmd.CmdEnsure,
		cmd.CmdTest,
		cmd.CmdStatus,
		cmd.CmdAdd,
		cmd.CmdRemove,
		cmd.CmdRelease,
		cmd.CmdRun,
		cmd.CmdUpdate,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Failed to run app with", os.Args, ":", err)
	}
}
