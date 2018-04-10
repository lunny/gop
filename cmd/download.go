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

// downloadFromGithubLike download from github like site, for example: github, gitea and etc.
//https://github.com/go-gitea/gitea/archive/master.zip
func downloadFromGithubLike(ctx *cli.Context, urlPrefix, pkgPrefix, pkg, refName, dstDir string) error {
	pkgCachePath := filepath.Join(dstDir, refName+".zip")
	if !ctx.Bool("override") && IsExist(pkgCachePath) {
		return nil
	}

	url := fmt.Sprintf("%s/%s/archive/%s.zip", urlPrefix, strings.TrimLeft(pkg, pkgPrefix), refName)
	Println("Downloading from", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := ioutil.TempFile(os.TempDir(), pkgPrefix)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(f.Name(), pkgCachePath)
}

// downloadFromGopm download from gopm.io
func downloadFromGopm(ctx *cli.Context, pkg, refName, dstDir string) error {
	pkgCachePath := filepath.Join(dstDir, refName+".zip")
	if !ctx.Bool("override") && IsExist(pkgCachePath) {
		return nil
	}

	url := fmt.Sprintf("https://gopm.io/api/v1/download?pkgname=%s&revision=%s", pkg, refName)
	Println("Downloading from", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := ioutil.TempFile(os.TempDir(), "gopm.io")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(f.Name(), pkgCachePath)
}

// CmdDownload represents download a package from github or gopm.io
var CmdDownload = cli.Command{
	Name:        "dl",
	Usage:       "Download one or more packages",
	Description: `Download one or more packages`,
	Action:      runDownload,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "override, o",
			Usage: "Download packages even it exists.",
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
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enables verbose progress and debug output",
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
	var rootDir = globalConfig.Repos.DefaultDir
	if ctx.String("target") != "" {
		rootDir = ctx.String("target")
	}
	paths := append([]string{rootDir}, pkgPaths...)
	dstDir := filepath.Join(paths...)

	var (
		err     error
		refName = "master"
	)

	source := ctx.String("source")
	switch source {
	case "origin":
		var found bool
		for _, vals := range globalConfig.Sources {
			if strings.HasPrefix(pkg, vals.PkgPrefix) {
				found = true
				err = downloadFromGithubLike(ctx, vals.UrlPrefix, vals.PkgPrefix, pkg, refName, dstDir)
				break
			}
		}

		if !found {
			return ErrNotSupported
		}
	case "gopm":
		err = downloadFromGopm(ctx, pkg, refName, dstDir)
	default:
		for _, vals := range globalConfig.Sources {
			if strings.HasPrefix(pkg, vals.PkgPrefix) {
				err = downloadFromGithubLike(ctx, vals.UrlPrefix, vals.PkgPrefix, pkg, refName, dstDir)
				break
			}
		}

		if err != nil {
			Println("Downloading failed:", err)
			err = downloadFromGopm(ctx, pkg, refName, dstDir)
		}
	}
	if err != nil {
		return err
	}

	if !ctx.Bool("recursive") {
		return nil
	}

	tmpBaseDir, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return err
	}
	// extract files to a tempory directory
	tmpDir := filepath.Join(append([]string{tmpBaseDir, "src"}, pkgPaths...)...)
	os.MkdirAll(filepath.Dir(tmpDir), os.ModePerm)

	tmpExtractDir, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return err
	}

	err = archiver.Zip.Open(filepath.Join(dstDir, refName+".zip"), tmpExtractDir)
	if err != nil {
		return err
	}

	f, err := os.Open(tmpExtractDir)
	if err != nil {
		return err
	}
	defer f.Close()
	dirs, err := f.Readdir(1)
	if err != nil {
		return err
	}
	if len(dirs) != 1 {
		return errors.New("unknow package")
	}

	if err = os.Rename(filepath.Join(tmpExtractDir, dirs[0].Name()), tmpDir); err != nil {
		return err
	}

	ctxt := build.Default
	globalGoPath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("Not found GOPATH")
	}

	ctxt.GOPATH = globalGoPath + string(filepath.ListSeparator) + tmpBaseDir
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
