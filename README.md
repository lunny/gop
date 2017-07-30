# GOP

[简体中文](README_ZH.md)

**Notice: We have changed the project structure and it isn't compitable with before. You have to change the old structure to new manually.**

GOP is a project manangement tool for golang application projects which you can place it anywhere(not in the GOPATH). Also this means it's **not** go-getable. GOP copy all denpendencies to src/vendor directory and all application's source is also in this directory. 

So a common process is below:

```
git clone xxx@mydata.com:bac/aaa.git
cd aaa
gop ensure
gop build
gop test
```

## Features

* GOPATH compitable
* Multiple build targets support
* Put your projects out of global GOPATH

## Installation

Please ensure you have installed the `go` command, GOP will invoke it on building or testing

```
go get github.com/lunny/gop
```

## Directory structure

This is an example project's directory.

```
<project root>
├── gop.yml
├── bin
├── doc
└── src
    ├── main
    │   └── main.go
    ├── models
    │   └── models.go
    ├── routes
    │   └── routes.go
    └── vendor
        └── github.com
            ├── go-xorm
            │   ├── builder
            │   ├── core
            │   └── xorm
            └── lunny
                ├── log
                └── tango
```

## Gop.yml

The project yml configuration file. This is an example.

```yml
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
```

## Command

### init

Create the default directory structs.

```
gop init
```

### ensure

Auto copy dependencies from $GOPATH to local project directory.

```
gop ensure [target_name]
```

### status

List all dependencies of this project and show the status.

```
gop status [target_name]
```

### add

Add a package to this project.

```
gop add <package>
```

### rm

Remove a package from this project.

```
gop rm <package>
```

### build

Run `go build` on the src directory.

```
gop build [target_name]
```

### run

Run `go run` on the src directory.

```
gop run [target_name]
```

### test

Run `go test` on the src directory.

```
gop test [target_name]
```

### release

Run `go release` on the src directory.

```
gop release [target_name]
```