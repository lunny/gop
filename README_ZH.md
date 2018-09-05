# GOP

[English](README.md)

[![CircleCI](https://circleci.com/gh/lunny/gop.svg?style=shield)](https://circleci.com/gh/lunny/gop) [![codecov](https://codecov.io/gh/lunny/gop/branch/master/graph/badge.svg)](https://codecov.io/gh/lunny/gop)
[![](https://goreportcard.com/badge/github.com/lunny/gop)](https://goreportcard.com/report/github.com/lunny/gop) 

GOP 是一个专为在GOPATH之外的Go应用开发的管理工具。当然他肯定不支持`Go Get`了。GOP 会将所有的依赖项拷贝到 `src/vendor` 目录下，应用本身的源代码也在 `src` 下。

一个通常的使用过程如下：

```
git clone xxx@mydata.com:bac/aaa.git
cd aaa
gop ensure -g
gop build
gop test
```

## 特性

* GOPATH 兼容，工程本身作为GOPATH
* 多编译目标支持
* 将您的工程放到全局GOPATH之外

## 安装

请确保您能正常运行Go命令，GOP 将依赖 Go 命令编译和测试

```
go get github.com/lunny/gop
```

## 工程目录结构

工程目录结构示例如下：

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

可以看出主文件默认放在 src/main 下可以自动识别，当然也可以在 Gop.yml 中指定 

## Gop.yml

工程配置文件，必须存在并且放在和src平级。如果你没有定义任何目标，默认的目标将是 src/main， 目标名是工程名。

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
  monitors:
  - config.ini
```

## 命令

### init

初始化 Gop 工程目录

```
mkdir newproject
cd newproject
gop init
```

如果你使用Vscode开发，你也可以同时创建编辑器的配置文件

```
gop init -e=vscode
```

### ensure

自动从全局 GOPATH 拷贝所需要的依赖项到 src/vendor 目录下。`-g` 参数将会自动调用 `go get <package>` 下载不在全局 `GOPATH` 中的包并且拷贝到 `vendor` 下。 `-u` 则总是调用 `go get <package>` 更新每一个依赖包并且拷贝到 `vendor` 下。

```
gop ensure [-g|-u] [target_name]
```

### status

列出当前目标所有依赖包并显示拷贝状态。

```
gop status [target_name]
```

### add

从 GOPATH 中拷贝指定的一个或多个依赖包到 vendor 目录下。`-u` 将会覆盖原先存在 `vendor` 下面的包。

```
gop add [-u] <package1> <package2>
```

### update

从 GOPATH 更新一个或多个依赖包到 vendor 目录下。如果有缺失的包会自动添加。`-f` 将会强制更新所有依赖的包。

```
gop update [-f] <package1> <package2>
```

### rm

从工程 vendor 中删除某个包。

```
gop rm <package>
```

### build

`go build` 编译目标。如果希望在编译之前自动之行 `ensure` 命令，可以使用 `-e`。

```
gop build [-e] [target_name]
```

### run

`go run` 编译并运行目标。`-w` 参数将会监视源代码中的Go文件变动并自动重新编译和运行。`-e` 参数将会在每次进行
编译之前先调用`ensure`确保所有的依赖项都拷贝到了vendor下。

```
gop run [-w] [-e] [target_name]
```

### test

运行 `go test` 将执行单元测试. 如果希望在编译之前自动之行 `ensure` 命令，可以使用 `-e`。

```
gop test [-e] [target_name]
```

### release

运行 `go release` 将自动编译并拷贝资源到 bin 目录下

```
gop release [target_name]
```

## TODO

* [ ] 依赖项版本支持
* [ ] 支持在 `GOPATH` 目录内运行