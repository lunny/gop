// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PkgType int

const (
	PkgTypeUnknown       = iota // 0
	PkgTypeGoRoot               // 1
	PkgTypeGloablGoPATH         // 2
	PkgTypeProjectGOPATH        // 3
	PkgTypeProjectVendor        // 4
)

// GetPkgType returns which the name's type
func getPkgType(globalGoPath, projectRoot, name string) (PkgType, bool, error) {
	exist, err := isDirExist(filepath.Join(projectRoot, "src", name))
	if err != nil {
		return PkgTypeUnknown, false, err
	}
	if exist {
		return PkgTypeProjectGOPATH, true, nil
	}

	exist, err = isDirExist(filepath.Join(projectRoot, "src", "vendor", name))
	if err != nil {
		return PkgTypeUnknown, false, err
	}
	if exist {
		return PkgTypeProjectVendor, true, nil
	}

	exist, err = isDirExist(filepath.Join(globalGoPath, "src", "vendor", name))
	if err != nil {
		return PkgTypeUnknown, false, err
	}
	if exist {
		return PkgTypeGloablGoPATH, true, nil
	}

	if IsGoRepoPath(name) {
		return PkgTypeGoRoot, true, nil
	}

	return PkgTypeGloablGoPATH, false, nil
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
	if Debug {
		log.Printf("Import/root path: %s : %s\n", importPath, projectRoot)
		log.Printf("Context GOPATH: %s\n", ctxt.GOPATH)
		log.Printf("Srouce path: %s\n", srcPath)
	}
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

		if Debug {
			log.Printf("Found dependency: %s--%v\n", name, pkgType)
		}

		switch pkgType {
		case PkgTypeGoRoot:
		case PkgTypeGloablGoPATH:
			imports = append(imports, Pkg{
				Name: name,
				Type: PkgTypeGloablGoPATH,
			})
			if exist {
				moreImports, err := ListImports(oldGOPATH, name, projectRoot, filepath.Join(oldGOPATH, "src"), tags, isTest)
				if err != nil {
					return nil, err
				}
				imports = append(imports, moreImports...)
			}
		case PkgTypeProjectGOPATH:
			imports = append(imports, Pkg{
				Name: name,
				Type: PkgTypeProjectGOPATH,
			})
			moreImports, err := ListImports(projectRoot, name, projectRoot, filepath.Join(projectRoot, "src"), tags, isTest)
			if err != nil {
				return nil, err
			}
			imports = append(imports, moreImports...)
		case PkgTypeProjectVendor:
			imports = append(imports, Pkg{
				Name: name,
				Type: PkgTypeProjectVendor,
			})
			moreImports, err := ListImports(projectRoot, name, projectRoot, filepath.Join(projectRoot, "src", "vendor"), tags, isTest)
			if err != nil {
				return nil, err
			}
			imports = append(imports, moreImports...)
		default:
			return nil, fmt.Errorf("unkonw type package")
		}
	}
	return imports, nil
}
