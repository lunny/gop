// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Unknwon/com"
	"github.com/lunny/gop/util"
	"github.com/urfave/cli"
)

// CmdEnsure represents
var CmdEnsure = cli.Command{
	Name:        "ensure",
	Usage:       "Ensure all the dependencies in the src directory",
	Description: `Ensure all the dependencies in the src directory`,
	Action:      runEnsure,
	Flags: []cli.Flag{
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
	},
}

var updatedPackage = make(map[string]struct{})

func ensure(cmd *cli.Context, globalGoPath, projectRoot, targetDir string) error {
	vendorDir := filepath.Join(projectRoot, "src", "vendor")
	imports, err := ListImports(".", filepath.Join(projectRoot, "src"), targetDir, "", true)
	if err != nil {
		return err
	}
	for _, imp := range imports {
		pkg := filepath.Join(projectRoot, "src", imp)
		if com.IsExist(pkg) {
			continue
		}
		if IsGoRepoPath(imp) {
			continue
		}

		if imp == "C" || strings.HasPrefix(imp, "../") || strings.HasPrefix(imp, "./") {
			continue
		}

		// get parent package
		imp, _ = util.NormalizeName(imp)

		// package dir
		srcDir := filepath.Join(globalGoPath, "src", imp)
		// FIXME: dry will lost some packages with -g or -u
		if cmd.IsSet("dry") {
			fmt.Println("Dry copying", imp)
			continue
		}

		// FIXME: imp only UNIX
		dstDir := filepath.Join(vendorDir, imp)

		if cmd.IsSet("update") {
			if _, ok := updatedPackage[imp]; ok {
				continue
			}

			fmt.Println("Downloading", imp)
			cmdGet := NewCommand("get").AddArguments("-u", imp)
			err = cmdGet.RunInDirPipeline("src", os.Stdout, os.Stderr)
			if err != nil {
				return err
			}

			fmt.Println("Copying", imp)
			os.RemoveAll(dstDir)
			err = CopyDir(srcDir, dstDir, func(path string) bool {
				return strings.HasPrefix(path, ".git")
			})
			if err != nil {
				return err
			}

			updatedPackage[imp] = struct{}{}

			return ensure(cmd, globalGoPath, projectRoot, targetDir)
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
				if cmd.IsSet("get") {
					fmt.Println("Downloading", imp)
					cmdGet := NewCommand("get").AddArguments(imp)
					err = cmdGet.RunInDirPipeline("src", os.Stdout, os.Stderr)
					if err != nil {
						return err
					}

					// scan the package dependencies again since the new package added
					return ensure(cmd, globalGoPath, projectRoot, targetDir)
				}

				fmt.Printf("Package %s not found on $GOPATH, please use -g option or go get at first\n", imp)
				return nil
			}

			fmt.Println("Copying", imp)
			err = CopyDir(srcDir, dstDir, func(path string) bool {
				return strings.HasPrefix(path, ".git")
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func runEnsure(cmd *cli.Context) error {
	var tags string
	ctxt := build.Default
	ctxt.BuildTags = strings.Split(tags, " ")

	globalGoPath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return errors.New("Not found GOPATH")
	}

	level, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	if err = loadConfig(filepath.Join(projectRoot, "gop.yml")); err != nil {
		return err
	}

	var args = cmd.Args()
	var targetName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		targetName = args[0]
	}

	if err = analysisTarget(level, targetName, projectRoot); err != nil {
		return err
	}

	ctxt.GOPATH = globalGoPath
	targetDir := filepath.Join(projectRoot, "src", curTarget.Dir)

	return ensure(cmd, globalGoPath, projectRoot, targetDir)
}

// IsDir returns true if given path is a directory,
// or returns false when it's a file or does not exist.
func IsDir(dir string) bool {
	f, e := os.Stat(dir)
	if e != nil {
		return false
	}
	return f.IsDir()
}

// Copy copies file from source to target path.
func Copy(src, dest string) error {
	// Gather file information to set back later.
	si, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Handle symbolic link.
	if si.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		// NOTE: os.Chmod and os.Chtimes don't recoganize symbolic link,
		// which will lead "no such file or directory" error.
		return os.Symlink(target, dest)
	}

	sr, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sr.Close()

	dw, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dw.Close()

	if _, err = io.Copy(dw, sr); err != nil {
		return err
	}

	// Set back file information.
	if err = os.Chtimes(dest, si.ModTime(), si.ModTime()); err != nil {
		return err
	}
	return os.Chmod(dest, si.Mode())
}

// IsExist checks whether a file or directory exists.
// It returns false when the file or directory does not exist.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func statDir(dirPath, recPath string, includeDir, isDirOnly bool) ([]string, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	fis, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	statList := make([]string, 0)
	for _, fi := range fis {
		if strings.Contains(fi.Name(), ".DS_Store") {
			continue
		}

		relPath := path.Join(recPath, fi.Name())
		curPath := path.Join(dirPath, fi.Name())
		if fi.IsDir() {
			if includeDir {
				statList = append(statList, relPath+"/")
			}
			s, err := statDir(curPath, relPath, includeDir, isDirOnly)
			if err != nil {
				return nil, err
			}
			statList = append(statList, s...)
		} else if !isDirOnly {
			statList = append(statList, relPath)
		}
	}
	return statList, nil
}

// StatDir gathers information of given directory by depth-first.
// It returns slice of file list and includes subdirectories if enabled;
// it returns error and nil slice when error occurs in underlying functions,
// or given path is not a directory or does not exist.
//
// Slice does not include given path itself.
// If subdirectories is enabled, they will have suffix '/'.
func StatDir(rootPath string, includeDir ...bool) ([]string, error) {
	if !IsDir(rootPath) {
		return nil, errors.New("not a directory or does not exist: " + rootPath)
	}

	isIncludeDir := false
	if len(includeDir) >= 1 {
		isIncludeDir = includeDir[0]
	}
	return statDir(rootPath, "", isIncludeDir, false)
}

// CopyDir copy files recursively from source to target directory.
//
// The filter accepts a function that process the path info.
// and should return true for need to filter.
//
// It returns error when error occurs in underlying functions.
func CopyDir(srcPath, destPath string, filters ...func(filePath string) bool) error {
	// Check if target directory exists.
	if IsExist(destPath) {
		return errors.New("file or directory alreay exists: " + destPath)
	}

	err := os.MkdirAll(destPath, os.ModePerm)
	if err != nil {
		return err
	}

	// Gather directory info.
	infos, err := StatDir(srcPath, true)
	if err != nil {
		return err
	}

	var filter func(filePath string) bool
	if len(filters) > 0 {
		filter = filters[0]
	}

	for _, info := range infos {
		if filter != nil && filter(info) {
			continue
		}

		curPath := path.Join(destPath, info)
		if strings.HasSuffix(info, "/") {
			err = os.MkdirAll(curPath, os.ModePerm)
		} else {
			err = Copy(path.Join(srcPath, info), curPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
