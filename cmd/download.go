// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	"github.com/urfave/cli"
)

var (
	ErrNotSupported = errors.New("The package path is not supported")
)

// downloadFromGithub download from github
func downloadFromGithub(pkg, refName, dstDir string) error {
	//https://github.com/go-gitea/gitea/archive/master.zip
	if !strings.HasPrefix(pkg, "github.com") {
		return ErrNotSupported
	}

	pkgCachePath := filepath.Join(dstDir, refName+".zip")
	if IsExist(pkgCachePath) {
		return nil
	}

	url := fmt.Sprintf("https://%s/archive/%s.zip", pkg, refName)
	Println("Downloading from", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	os.MkdirAll(dstDir, os.ModePerm)
	f, err := os.Create(pkgCachePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		os.Remove(pkgCachePath)
	}
	return err
}

// downloadFromGopm download from gopm.io
func downloadFromGopm(pkg, refName, dstDir string) error {
	pkgCachePath := filepath.Join(dstDir, refName+".zip")
	if IsExist(pkgCachePath) {
		return nil
	}

	url := fmt.Sprintf("https://gopm.io/api/v1/download?pkgname=%s&revision=%s", pkg, refName)
	Println("Downloading from", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	os.MkdirAll(dstDir, os.ModePerm)
	f, err := os.Create(pkgCachePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		os.Remove(pkgCachePath)
	}
	return err
}

// CmdDownload represents download a package from github or gopm.io
var CmdDownload = cli.Command{
	Name:        "dl",
	Usage:       "Download one or more packages",
	Description: `Download one or more packages`,
	Action:      runDownload,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enables verbose progress and debug output",
		},
		cli.BoolFlag{
			Name:  "recursive, r",
			Usage: "Also download all the dependent packages",
		},
		cli.StringFlag{
			Name:  "source, s",
			Usage: "Download source",
		},
		cli.StringFlag{
			Name:  "target, t",
			Usage: "Download target directory",
		},
	},
}

func runDownload(ctx *cli.Context) error {
	if len(ctx.Args()) <= 0 {
		return errors.New("You have to indicate one or more packages")
	}

	homeDir, err := Home()
	if err != nil {
		return err
	}

	err = loadGlobalConfig(filepath.Join(homeDir, ".gop.yml"))
	if err != nil {
		return err
	}

	showLog = ctx.IsSet("verbose")
	names := ctx.Args()
	for _, name := range names {
		if err := download(ctx, name); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func download(ctx *cli.Context, pkg string) error {
	pkgPaths := strings.Split(pkg, "/")
	var rootDir = globalConfig.Get("repos.default_dir")
	if ctx.String("target") != "" {
		rootDir = ctx.String("target")
	}
	dstDir := filepath.Join(append([]string{rootDir}, pkgPaths...)...)

	var err error
	switch ctx.String("source") {
	case "origin":
		err = downloadFromGithub(pkg, "master", dstDir)
	case "gopm":
		err = downloadFromGopm(pkg, "master", dstDir)
	default:
		err := downloadFromGithub(pkg, "master", dstDir)
		if err != nil {
			Println("Downloading failed:", err)
			err = downloadFromGopm(pkg, "master", dstDir)
		}
	}
	if err != nil {
		return err
	}

	if !ctx.Bool("recursive") {
		return nil
	}

	// extract files to a tempory directory
	tmpDir, err := ioutil.TempDir(os.TempDir(), strings.Replace(pkg, "/", "_", -1))
	if err != nil {
		return err
	}

	err = archiver.Zip.Open(filepath.Join(dstDir, "master.zip"), tmpDir)
	if err != nil {
		return err
	}

	ctxt := build.Default
	globalGoPath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("Not found GOPATH")
	}

	ctxt.GOPATH = globalGoPath
	if Debug {
		log.Printf("Import/root path: %s\n", tmpDir)
		log.Printf("Context GOPATH: %s\n", ctxt.GOPATH)
		log.Printf("Srouce path: %s\n", tmpDir)
	}
	dependentPkgs, err := ctxt.Import(pkg, tmpDir, build.AllowBinary)
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			return fmt.Errorf("fail to get imports(%s): %v", pkg, err)
		}
		log.Printf("Getting imports: %v\n", err)
	}

	for _, subPkg := range dependentPkgs.Imports {
		if IsGoRepoPath(subPkg) ||
			strings.HasPrefix(subPkg, pkg) {
			continue
		}

		if err = download(ctx, subPkg); err != nil {
			return err
		}
	}

	return nil
}
