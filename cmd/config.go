package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Unknwon/com"
	yaml "gopkg.in/yaml.v2"
)

// Target build target
type Target struct {
	Name   string
	Dir    string
	Assets []string
}

// Config gop.yml
type Config struct {
	Targets []Target
}

var config Config

func loadConfig(ymlPath string) error {
	if com.IsExist(ymlPath) {
		Println("find config file", ymlPath)
		bs, err := ioutil.ReadFile(ymlPath)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(bs, &config)
		if err != nil {
			return err
		}
	}

	if len(config.Targets) == 0 {
		projectName := filepath.Base(filepath.Dir(ymlPath))
		config.Targets = []Target{
			{
				Name: projectName,
				Dir:  "main",
			},
		}
	}
	return nil
}

const (
	dirLevelOutProject = iota // command run out of project root
	dirLevelRoot              // command run in project root
	dirLevelSrc               // command run in <project root>/src
	dirLevelTarget            // command run in <project root>/src/<target>
)

func analysisDirLevel() (int, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 0, "", err
	}

	if com.IsExist(filepath.Join(wd, "gop.yml")) {
		return dirLevelRoot, wd, nil
	} else if filepath.Base(wd) == "src" &&
		com.IsExist(filepath.Join(filepath.Dir(wd), "gop.yml")) {
		return dirLevelSrc, filepath.Dir(wd), nil
	} else if filepath.Base(filepath.Dir(wd)) == "src" &&
		com.IsExist(filepath.Join(filepath.Dir(filepath.Dir(wd)), "gop.yml")) {
		return dirLevelTarget, filepath.Dir(filepath.Dir(wd)), nil
	}
	return dirLevelOutProject, "", errors.New("unknow directory to run gop")
}
