# Thriftgo

[English](README.md) | 中文

**Thriftgo** 是 go 语言实现的 thrift 编译器。它具有与 apache/thrift 编译器类似的命令行界面，并通过插件机制增强了功能，使其更加强大。

## 安装

注意：在执行以下命令之前，请**确保**您的 `GOPATH` 环境已正确配置。

使用 `go install`：

`GO111MODULE=on go install github.com/cloudwego/thriftgo@latest`

或从源代码构建：

```shell
git clone https://github.com/cloudwego/thriftgo.git
cd thriftgo
export GO111MODULE=on
go mod tidy
go build
go install
```

## 使用

Thriftgo 命令行工具接受 IDL 文件作为输入，并将其编译为目标语言。每个后端都有一系列丰富的选项，用于定制生成的代码。

Thriftgo 目前只能生成 golang 代码。将来还会添加更多后端。

要使用默认配置将 thrift IDL 编译为 golang 文件，只需运行：

```shell
thriftgo -g go the-idl-file.thrift
```

运行 `thriftgo -h` 查看每个后端的所有可用选项及其含义。

## 插件

如果 Thriftgo 生成的代码不能满足您的需求，且提供的选项无法满足您的要求，您可以编写插件来利用 Thriftgo 的 IDL 解析器生成额外的代码。有关更多信息，请查阅插件包的文档。
