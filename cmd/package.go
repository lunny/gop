// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var goRepoPath = map[string]bool{
	"builtin": true,
}

func init() {
	standardPath, err := retrieveGoStdPkgs()
	if err != nil {
		panic(err)
	}

	for p := range standardPath {
		if strings.HasPrefix(p, "vendor/") {
			continue
		}
		for {
			goRepoPath[p] = true
			i := strings.LastIndex(p, "/")
			if i < 0 {
				break
			}
			p = p[:i]
		}
	}
}

// retrieveGoVersion
func retrieveGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	bs, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if len(bs) < 16 {
		return "", errors.New("retrieve go version failed")
	}

	v := strings.TrimLeft(string(bs), "go version")
	v = strings.Split(v, " ")[0]
	v = strings.TrimLeft(v, "go")
	return v, nil
}

// retrieveGoStdPkgs retrieve go std pkg names
func retrieveGoStdPkgs() (map[string]bool, error) {
	cmd := exec.Command("go", "list", "-f", `"{{.ImportPath}}": true,`, "std")
	bs, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	bs[len(bs)-2] = '}'
	res := make([]byte, 0, len(bs)+1)
	res = append(res, '{')
	res = append(res, bs...)

	var ret = make(map[string]bool)
	err = json.Unmarshal(res, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// IsGoRepoPath returns true if package is from standard library.
func IsGoRepoPath(importPath string) bool {
	return goRepoPath[importPath]
}

type PkgType int

const (
	PkgTypeUnknown       = iota // 0
	PkgTypeGoRoot               // 1
	PkgTypeGloablGoPath         // 2
	PkgTypeProjectGoPath        // 3
	PkgTypeProjectVendor        // 4
)

// GetPkgType returns which the name's type
func getPkgType(globalGoPath, projectRoot, name string) (PkgType, bool, error) {
	exist, err := isDirExist(filepath.Join(projectRoot, "src", name))
	if err != nil {
		return PkgTypeUnknown, false, err
	}
	if exist {
		return PkgTypeProjectGoPath, true, nil
	}

	vendorPath := filepath.Join(projectRoot, "src", "vendor", name)
	exist, err = isDirExist(vendorPath)
	if err != nil {
		return PkgTypeUnknown, false, err
	}
	if exist {
		f, err := os.Open(vendorPath)
		if err != nil {
			return PkgTypeUnknown, false, err
		}
		defer f.Close()
		files, err := f.Readdirnames(0)
		if err != nil {
			return PkgTypeUnknown, false, err
		}

		for _, f := range files {
			if filepath.Ext(f) == ".go" {
				return PkgTypeProjectVendor, true, nil
			}
		}

		exist = false
	}

	exist, err = isDirExist(filepath.Join(globalGoPath, "src", "vendor", name))
	if err != nil {
		return PkgTypeUnknown, false, err
	}
	if exist {
		return PkgTypeGloablGoPath, true, nil
	}

	if IsGoRepoPath(name) {
		return PkgTypeGoRoot, true, nil
	}

	return PkgTypeGloablGoPath, false, nil
}

type Pkg struct {
	Name  string
	Type  PkgType
	Exist bool
}

// ListImports list all the dependencies packages name
func ListImports(gopath, importPath, projectRoot, srcPath, tags string, isTest bool) ([]Pkg, error) {
	ctxt := build.Default
	ctxt.BuildTags = strings.Split(tags, " ")
	ctxt.GOPATH = gopath

	Printf("Import/root path: %s : %s\n", importPath, projectRoot)
	Printf("Context GOPATH: %s\n", ctxt.GOPATH)
	Printf("Source path: %s\n", srcPath)

	pkg, err := ctxt.Import(importPath, srcPath, build.AllowBinary)
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			return nil, fmt.Errorf("fail to get imports(%s): %v", importPath, err)
		}
		log.Printf("Getting imports: %v\n", err)
	}

	rawImports := pkg.Imports
	numImports := len(rawImports)
	if isTest {
		rawImports = append(rawImports, pkg.TestImports...)
		numImports = len(rawImports)
	}
	imports := make([]Pkg, 0, numImports)
	oldGOPATH, _ := os.LookupEnv("GOPATH")
	for _, name := range rawImports {
		if name == "C" || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "./") {
			continue
		}

		pkgType, exist, err := getPkgType(oldGOPATH, projectRoot, name)
		if err != nil {
			return nil, err
		}

		Printf("Found dependency: %s--%v\n", name, pkgType)

		switch pkgType {
		case PkgTypeGoRoot:
		case PkgTypeGloablGoPath:
			imports = append(imports, Pkg{
				Name: name,
				Type: PkgTypeGloablGoPath,
			})
			if exist {
				moreImports, err := ListImports(oldGOPATH, name, projectRoot, filepath.Join(oldGOPATH, "src"), tags, isTest)
				if err != nil {
					return nil, err
				}
				imports = append(imports, moreImports...)
			}
		case PkgTypeProjectGoPath:
			imports = append(imports, Pkg{
				Name: name,
				Type: PkgTypeProjectGoPath,
			})
			if exist {
				moreImports, err := ListImports(projectRoot, name, projectRoot, filepath.Join(projectRoot, "src"), tags, isTest)
				if err != nil {
					return nil, err
				}
				imports = append(imports, moreImports...)
			}
		case PkgTypeProjectVendor:
			imports = append(imports, Pkg{
				Name: name,
				Type: PkgTypeProjectVendor,
			})
			if exist {
				moreImports, err := ListImports(projectRoot, name, projectRoot, filepath.Join(projectRoot, "src", "vendor"), tags, isTest)
				if err != nil {
					return nil, err
				}
				imports = append(imports, moreImports...)
			}
		default:
			return nil, fmt.Errorf("unkonw type package")
		}
	}
	return imports, nil
}
