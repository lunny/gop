// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	assert.Equal(t, globalConfig.Get("sources.github.url_prefix"), "https://github.com")
	assert.Equal(t, globalConfig.Get("sources.github.pkg_prefix"), "github.com")
}

func TestSetConfig(t *testing.T) {
	globalConfig.Set("sources.gitea_demo.url_prefix", "https://try.gitea.io")
	globalConfig.Set("sources.gitea_demo.pkg_prefix", "gitea.io")

	assert.Equal(t, globalConfig.Get("sources.gitea_demo.url_prefix"), "https://try.gitea.io")
	assert.Equal(t, globalConfig.Get("sources.gitea_demo.pkg_prefix"), "gitea.io")
}
