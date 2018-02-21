// Copyright 2017 The Gop Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*

Package gop is a project manangement tool for building and manage your golang applications out of GOPATH.
Also this means it's not go-getable. GOP copy all denpendencies to src/vendor directory and all application's source is also in this directory.
A normal process using gop is below:

    git clone xxx@mydata.com:bac/aaa.git
    cd aaa
    gop ensure -g
    gop build
	gop test

Features

1. GOPATH compatible
2. Multiple build targets support
3. Put your projects out of global GOPATH

Installation

Please ensure you have installed the go command, GOP will invoke it on building or testing

	go get github.com/lunny/gop

Directory structure

This is an example project's directory.

	<project root>
	├── gop.yml
	├── bin
	├── doc
	└── src
		├── main
		│   └── main.go
		├── models
		│   └── models.go
		├── routes
		│   └── routes.go
		└── vendor
			└── github.com
				├── go-xorm
				│   ├── builder
				│   ├── core
				│   └── xorm
				└── lunny
					├── log
					└── tango

Gop.yml

The project yml configuration file. This is an example. Of course, if you didn't define any target, the default target is src/main and the target name is the project name.

    targets:
	- name: myproject1
		dir: main
		assets:
			- templates
			- public
			- config.ini
			- key.pem
			- cert.pem
	- name: myproject2
		dir: web
		assets:
			- templates
			- public
			- config.ini

Command

1. init

Create the default directory structs.

	gop init

2. ensure

Auto copy dependencies from $GOPATH to local project directory. -g will let you automatically call go get <package> when the package is missing on GOPATH. -u will always go get <package> on all the dependencies and copy them to vendor.

	gop ensure [-g|-u] [target_name]

3. status

List all dependencies of this project and show the status.

	gop status [target_name]

4. add

Add a package to this project. -u will override the package dir on vendor.

	gop add [-u] <package>

5. rm

Remove a package from this project.

	gop rm <package>

6. build

Run go build on the src directory.

	gop build [target_name]

7. run

Run go run on the src directory.

	gop run [target_name]

8. test

Run go test on the src directory.

	gop test [target_name]

9. release

Run go release on the src directory.

	gop release [target_name]
*/
package main
