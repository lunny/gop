# GOP

GOP is a golang project management tool only for golang application project which could be put anywhere(not in the GOPATH). Also this means it's not go-getable. GOP copy all denpendencies to src directory and all application's source is also in this directory. 

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

```
gop init
```

### ensure

```
gop ensure
```

### status

```
gop status
```

### add

```
gop add <package>
```

### rm

```
gop rm <package>
```

### test

```
gop test
```