# Thriftgo

[英語](README.md) | [中文](README_cn.md) | 日本語

**Thriftgo** は、Go言語で実装された [thrift](https://thrift.apache.org/docs/idl) コンパイラです。apache/thrift コンパイラと似たコマンドラインインターフェースを持ち、プラグイン機構により強化されており、より強力です。

## インストール

注意: 以下のコマンドを実行する前に、**`GOPATH` 環境が正しく設定されていることを確認してください**。

`go install` を使用:

`GO111MODULE=on go install github.com/cloudwego/thriftgo@latest`

またはソースからビルド:

```shell
git clone https://github.com/cloudwego/thriftgo.git
cd thriftgo
export GO111MODULE=on
go mod tidy
go build
go install
```

## 使用方法

Thriftgo コマンドラインツールは IDL ファイルを受け取り、ターゲット言語にコンパイルします。各バックエンドには、生成されたコードをカスタマイズするための豊富なオプションが用意されています。

現時点では、Thriftgo は golang コードのみを生成できます。将来的には、他のバックエンドも追加される予定です。

デフォルト設定で thrift IDL を golang ファイルにコンパイルするには、次のコマンドを実行します:

```shell
thriftgo -g go the-idl-file.thrift
```

各バックエンドのすべての利用可能なオプションとその意味を確認するには、`thriftgo -h` を実行してください。

## プラグイン

Thriftgo が生成するコードがニーズを満たさず、提供されているオプションが要件を満たさない場合は、プラグインを作成して Thriftgo の IDL パーサーを利用して追加のコードを生成することができます。詳細については、プラグインパッケージのドキュメントを参照してください。
