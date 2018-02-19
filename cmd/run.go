// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli"
	fsnotify "gopkg.in/fsnotify.v1"
)

// CmdRun represents
var CmdRun = cli.Command{
	Name:            "run",
	Usage:           "Run the target and monitor the source file changes",
	Description:     `Run the target and monitor the source file changes`,
	Action:          runRun,
	SkipFlagParsing: true,
}

var process *os.Process
var processLock sync.Mutex

func runBinary(exePath string) error {
	attr := &os.ProcAttr{
		Dir:   filepath.Dir(exePath),
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	var err error
	process, err = os.StartProcess(filepath.Base(exePath), []string{exePath}, attr)
	if err != nil {
		return err
	}

	return nil
}

func runRun(ctx *cli.Context) error {
	var watchFlagIdx = -1
	var args = ctx.Args()
	for i, arg := range args {
		if arg == "-v" {
			showLog = true
		} else if arg == "-w" {
			watchFlagIdx = i
		}
	}

	if watchFlagIdx > -1 {
		args = append(args[:watchFlagIdx], args[watchFlagIdx+1:]...)
	}

	var isWindows = runtime.GOOS == "windows"
	// gop run don't support cross compile
	err := runBuildNoCtx(args, isWindows)
	if err != nil {
		return err
	}

	_, projectRoot, err := analysisDirLevel()
	if err != nil {
		return err
	}

	var ext string
	if isWindows {
		ext = ".exe"
	}

	exePath := filepath.Join(projectRoot, "src", curTarget.Dir, curTarget.Name+ext)
	exePath, _ = filepath.Abs(exePath)

	if watchFlagIdx <= -1 {
		return runBinary(exePath)
	}

	go func() {
		processLock.Lock()
		err := runBinary(exePath)
		if err != nil {
			Println("Run failed:", err)
			process = nil
		}
		processLock.Unlock()
	}()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = filepath.Walk(filepath.Join(projectRoot, "src"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if strings.HasSuffix(event.Name, ".go") {
						processLock.Lock()
						if process != nil {
							Println("Killing old process")
							if err := process.Kill(); err != nil {
								log.Println("Killing old process error:", err)
								done <- false
								processLock.Unlock()
								return
							}
							process = nil
						}
						err := runBuildNoCtx(args, isWindows)
						if err != nil {
							log.Println("Build error:", err)
						} else {
							err = runBinary(exePath)
							if err != nil {
								log.Println("Run binary error:", err)
							}
						}
						processLock.Unlock()
					}
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					exist, _ := isDirExist(event.Name)
					if exist {
						watcher.Add(event.Name)
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					exist, _ := isDirExist(event.Name)
					if exist {
						watcher.Remove(event.Name)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
				done <- false
				return
			case <-time.After(200 * time.Millisecond):
			}
		}
	}()

	<-done

	if process != nil {
		if err := process.Kill(); err != nil {
			log.Println("error:", err)
		}
	}

	return nil
}
