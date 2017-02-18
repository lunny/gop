# GOP

GOP is a project manangement tool for golang application projects which you can place it anywhere(not in the GOPATH). Also this means it's not go-getable. GOP copy all denpendencies to src directory and all application's source is also in this directory. 

So a common process is below:

```
git clone xxx@mydata.com:bac/aaa.git
cd aaa
gop ensure
gop build
gop test
```

## Installation

Please ensure you have install the `go` command, GOP  will invoke it on building or testing

```
go get github.com/lunny/gop
```

## Directory structure

This is an example project's directory.

```
<projct root>
├── bin
└── src
    ├── github.com
    │   ├── go-xorm
    │   │   ├── builder
    │   │   ├── core
    │   │   └── xorm
    │   └── lunny
    │       ├── log
    │       └── tango
    ├── main.go
    └── models
        └── models.go
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
gop ensure
```

### status

List all dependencies of this project and show the status.

```
gop status
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
gop build
```

### test

Run `go test` on the src directory.

```
gop test
```