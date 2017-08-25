// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unknwon/com"
	"github.com/urfave/cli"
)

// CmdStatus represents
var CmdStatus = cli.Command{
	Name:            "status",
	Usage:           "List this project's dependencies",
	Description:     `List this project's dependencies`,
	Action:          runStatus,
	SkipFlagParsing: true,
}

func runStatus(cmd *cli.Context) error {
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
	srcDir := filepath.Join(projectRoot, "src", curTarget.Dir)
	vendorDir := filepath.Join(projectRoot, "src", "vendor")

	imports, err := ListImports(".", filepath.Join(projectRoot, "src"), srcDir, "", true)
	if err != nil {
		return err
	}
	for i, imp := range imports {
		pkg := filepath.Join(projectRoot, "src", imp)
		if com.IsExist(pkg) {
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
		p := filepath.Join(vendorDir, imp)
		exist, err := isDirExist(p)
		if err != nil {
			return err
		}
		if exist {
			fmt.Print("[X] ")
		} else {
			fmt.Print("[ ] ")
		}
		fmt.Println(imp)
	}

	return nil
}

var standardPath = map[string]bool{
	"builtin": true,

	// go list -f '"{{.ImportPath}}": true,'  std
	"archive/tar":         true,
	"archive/zip":         true,
	"bufio":               true,
	"bytes":               true,
	"compress/bzip2":      true,
	"compress/flate":      true,
	"compress/gzip":       true,
	"compress/lzw":        true,
	"compress/zlib":       true,
	"container/heap":      true,
	"container/list":      true,
	"container/ring":      true,
	"context":             true,
	"crypto":              true,
	"crypto/aes":          true,
	"crypto/cipher":       true,
	"crypto/des":          true,
	"crypto/dsa":          true,
	"crypto/ecdsa":        true,
	"crypto/elliptic":     true,
	"crypto/hmac":         true,
	"crypto/md5":          true,
	"crypto/rand":         true,
	"crypto/rc4":          true,
	"crypto/rsa":          true,
	"crypto/sha1":         true,
	"crypto/sha256":       true,
	"crypto/sha512":       true,
	"crypto/subtle":       true,
	"crypto/tls":          true,
	"crypto/x509":         true,
	"crypto/x509/pkix":    true,
	"database/sql":        true,
	"database/sql/driver": true,
	"debug/dwarf":         true,
	"debug/elf":           true,
	"debug/gosym":         true,
	"debug/macho":         true,
	"debug/pe":            true,
	"encoding":            true,
	"encoding/ascii85":    true,
	"encoding/asn1":       true,
	"encoding/base32":     true,
	"encoding/base64":     true,
	"encoding/binary":     true,
	"encoding/csv":        true,
	"encoding/gob":        true,
	"encoding/hex":        true,
	"encoding/json":       true,
	"encoding/pem":        true,
	"encoding/xml":        true,
	"errors":              true,
	"expvar":              true,
	"flag":                true,
	"fmt":                 true,
	"go/ast":              true,
	"go/build":            true,
	"go/doc":              true,
	"go/format":           true,
	"go/parser":           true,
	"go/printer":          true,
	"go/scanner":          true,
	"go/token":            true,
	"hash":                true,
	"hash/adler32":        true,
	"hash/crc32":          true,
	"hash/crc64":          true,
	"hash/fnv":            true,
	"html":                true,
	"html/template":       true,
	"image":               true,
	"image/color":         true,
	"image/color/palette": true,
	"image/draw":          true,
	"image/gif":           true,
	"image/jpeg":          true,
	"image/png":           true,
	"index/suffixarray":   true,
	"io":                  true,
	"io/ioutil":           true,
	"log":                 true,
	"log/syslog":          true,
	"math":                true,
	"math/big":            true,
	"math/cmplx":          true,
	"math/rand":           true,
	"mime":                true,
	"mime/multipart":      true,
	"net":                 true,
	"net/http":            true,
	"net/http/cgi":        true,
	"net/http/cookiejar":  true,
	"net/http/fcgi":       true,
	"net/http/httptest":   true,
	"net/http/httputil":   true,
	"net/http/pprof":      true,
	"net/mail":            true,
	"net/rpc":             true,
	"net/rpc/jsonrpc":     true,
	"net/smtp":            true,
	"net/textproto":       true,
	"net/url":             true,
	"os":                  true,
	"os/exec":             true,
	"os/signal":           true,
	"os/user":             true,
	"path":                true,
	"path/filepath":       true,
	"reflect":             true,
	"regexp":              true,
	"regexp/syntax":       true,
	"runtime":             true,
	"runtime/cgo":         true,
	"runtime/debug":       true,
	"runtime/pprof":       true,
	"runtime/race":        true,
	"sort":                true,
	"strconv":             true,
	"strings":             true,
	"sync":                true,
	"sync/atomic":         true,
	"syscall":             true,
	"testing":             true,
	"testing/iotest":      true,
	"testing/quick":       true,
	"text/scanner":        true,
	"text/tabwriter":      true,
	"text/template":       true,
	"text/template/parse": true,
	"time":                true,
	"unicode":             true,
	"unicode/utf16":       true,
	"unicode/utf8":        true,
	"unsafe":              true,
}

var goRepoPath = map[string]bool{}

func init() {
	for p := range standardPath {
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

// IsGoRepoPath returns true if package is from standard library.
func IsGoRepoPath(importPath string) bool {
	return goRepoPath[importPath]
}

var (
	// Debug indicated whether it is debug mode
	Debug = false
)

func isDirExist(dirName string) (bool, error) {
	f, err := os.Stat(dirName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !f.IsDir() {
		return false, errors.New("the same name file exist")
	}
	return true, nil
}

// ListImports list all the dependencies packages name
func ListImports(importPath, rootPath, srcPath, tags string, isTest bool) ([]string, error) {
	oldGOPATH := os.Getenv("GOPATH")
	ctxt := build.Default
	ctxt.BuildTags = strings.Split(tags, " ")
	ctxt.GOPATH = oldGOPATH
	if Debug {
		log.Printf("Import/root path: %s : %s\n", importPath, rootPath)
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
	imports := make([]string, 0, numImports)
	for _, name := range rawImports {
		if IsGoRepoPath(name) {
			continue
		}

		if name == "C" || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "./") {
			continue
		}

		exist, err := isDirExist(filepath.Join(rootPath, name))
		if err != nil {
			return nil, err
		}

		if Debug {
			log.Printf("Found dependency: %s\n", name)
		}

		if exist {
			imports = append(imports, name)

			moreImports, err := ListImports("./"+name, rootPath, rootPath, tags, isTest)
			if err != nil {
				return nil, err
			}
			for _, pkgName := range moreImports {
				imports = append(imports, pkgName)
			}
		} else {
			imports = append(imports, name)

			oldGOPATH := os.Getenv("GOPATH")
			moreImports, err := ListImports(name, filepath.Join(oldGOPATH, "src", name), filepath.Join(oldGOPATH, "src", name), tags, false)
			if err != nil {
				return nil, err
			}
			for _, pkgName := range moreImports {
				imports = append(imports, pkgName)
			}
		}
	}
	return imports, nil
}
