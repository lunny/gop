// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
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

func runBinary(exePath string, wait bool) error {
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

	if wait {
		_, err = process.Wait()
	}
	return err
}

func killOldProcess(done chan bool) {
	fmt.Println("=== Killing the old process")
	if process != nil {
		if err := process.Kill(); err != nil {
			log.Println("Killing old process error:", err)
			done <- false
			processLock.Unlock()
			return
		}
		process = nil
	}
}

func reBuildAndRun(ctx *cli.Context, args cli.Args, isWindows, ensureFlag bool, exePath string, done chan bool) {
	processLock.Lock()
	if isWindows {
		killOldProcess(done)
	}

	fmt.Printf("=== Rebuilding %s ...\n", args)
	err := runBuildNoCtx(ctx, args, isWindows, ensureFlag)
	if err != nil {
		log.Println("Build error:", err)
	} else {
		if !isWindows {
			killOldProcess(done)
		}

		fmt.Printf("=== Running %s ...\n", exePath)
		err = runBinary(exePath, false)
		if err != nil {
			log.Println("Run binary error:", err)
		}
	}
	processLock.Unlock()
}

const (
	noNeedReBuildAndRun = iota
	needReBuildAndRun
	needReRun
)

func needReBuild(projectRoot, fileName string) int {
	if strings.HasSuffix(fileName, ".go") {
		return needReBuildAndRun
	} else if strings.HasSuffix(fileName, ".log") {
		return noNeedReBuildAndRun
	}

	for _, f := range curTarget.Monitors {
		if filepath.Join(projectRoot, "src", curTarget.Dir, f) == fileName {
			return needReRun
		}
	}
	return noNeedReBuildAndRun
}

func runRun(ctx *cli.Context) error {
	var (
		watchFlagIdx  = -1
		ensureFlagIdx = -1
		args          = ctx.Args()
	)
	for i, arg := range args {
		if arg == "-v" {
			showLog = true
		} else if arg == "-w" {
			watchFlagIdx = i
		} else if arg == "-e" {
			ensureFlagIdx = i
		}
	}

	if watchFlagIdx > -1 {
		args = append(args[:watchFlagIdx], args[watchFlagIdx+1:]...)
		if ensureFlagIdx > watchFlagIdx {
			ensureFlagIdx = ensureFlagIdx - 1
		}
	}
	if ensureFlagIdx > -1 {
		args = append(args[:ensureFlagIdx], args[ensureFlagIdx+1:]...)
	}

	var isWindows = runtime.GOOS == "windows"
	// gop run don't support cross compile
	err := runBuildNoCtx(ctx, args, isWindows, ensureFlagIdx > -1)
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
		return runBinary(exePath, true)
	}

	go func() {
		processLock.Lock()
		err := runBinary(exePath, false)
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
	var lastTimeLock sync.Mutex
	var lastTime time.Time
	var needChangeType int

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					needChange := needReBuild(projectRoot, event.Name)
					if needChange == needReBuildAndRun || needChange == needReRun {
						exist, _ := isFileExist(event.Name)
						if exist {
							lastTimeLock.Lock()
							lastTime = time.Now()
							needChangeType = needChange
							lastTimeLock.Unlock()
						}
					}
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					needChange := needReBuild(projectRoot, event.Name)
					if needChange == needReBuildAndRun || needChange == needReRun {
						lastTimeLock.Lock()
						lastTime = time.Now()
						needChangeType = needChange
						lastTimeLock.Unlock()
					}
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					exist, _ := isDirExist(event.Name)
					if exist {
						watcher.Add(event.Name)
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					watcher.Remove(event.Name)
					needChange := needReBuild(projectRoot, event.Name)
					if needChange == needReBuildAndRun || needChange == needReRun {
						lastTimeLock.Lock()
						lastTime = time.Now()
						needChangeType = needChange
						lastTimeLock.Unlock()
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
				done <- false
				return
			case <-time.After(200 * time.Millisecond):
				var reBuild bool
				var reType int
				now := time.Now()
				lastTimeLock.Lock()
				reBuild = !lastTime.IsZero() && now.Unix()-lastTime.Unix() >= 1
				if reBuild {
					lastTime = time.Time{}
					reType = needChangeType
				}
				lastTimeLock.Unlock()
				if reBuild {
					switch reType {
					case needReBuildAndRun:
						reBuildAndRun(ctx, args, isWindows, ensureFlagIdx > -1, exePath, done)
					case needReRun:
						processLock.Lock()
						killOldProcess(done)

						fmt.Printf("=== Running %s ...\n", exePath)
						err = runBinary(exePath, false)
						if err != nil {
							log.Println("Run binary error:", err)
						}
						processLock.Unlock()
					}
				}
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
