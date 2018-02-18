// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"os"
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

func isFileExist(fileName string) (bool, error) {
	f, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if f.IsDir() {
		return false, errors.New("the same name directory exist")
	}
	return true, nil
}
