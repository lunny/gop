# GOP

[English](README.md)

**注意：从v0.2到v0.3，目录结构已经完全改变并且不兼容，请手动进行迁移更改。**

GOP 是一个专为Golang应用开发的工程管理工具，通过这个工具你可以将你的工程放在任何地方（在GOPATH之外）。当然他肯定不支持Go Get了。GOP 会将所有的依赖项拷贝到 src/vendor 目录下，应用本身的源代码也在 src 下。

一个通常的使用过程如下：

```
git clone xxx@mydata.com:bac/aaa.git
cd aaa
gop ensure
gop build
gop test
```

## 特性

* GOPATH 兼容，工程本身作为GOPATH
* 多编译目标支持
* 将您的工程放到全局GOPATH之外

## 安装

情确保您能正常运行Go命令，GOP 将依赖 Go 命令编译和测试

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
```

## 命令

### init

初始化 Gop 工程目录

```
mkdir newproject
cd newproject
gop init
```

### ensure

自动从全局 GOPATH 拷贝所需要的依赖项到 src/vendor 目录下

```
gop ensure [target_name]
```

### status

列出当前目标所有依赖包并显示拷贝状态。

```
gop status [target_name]
```

### add

从 GOPATH 中拷贝指定的依赖包到 vendor 目录下。

```
gop add <package>
```

### rm

从工程 vendor 中删除某个包。

```
gop rm <package>
```

### build

`go build` 编译目标

```
gop build [target_name]
```

### run

`go run` 编译并运行目标

```
gop run [target_name]
```

### test

运行 `go test` 将执行单元测试.

```
gop test [target_name]
```

### release

运行 `go release` 将自动编译并拷贝资源到 bin 目录下

```
gop release [target_name]
```