# ASKI

[![Go Report Card](https://goreportcard.com/badge/github.com/kznrluk/aski)](https://goreportcard.com/report/github.com/kznrluk/aski)

`aski` は、ターミナルからChatGPTを扱うことのできるミニマルなアプリケーションです。

## 機能

- ファイル読み取り機能
- プロファイル機能 異なる会話コンテキストを簡単に切り替えることができます。
- Streaming APIとREST APIのサポート

## インストール
今のところ `go install` のみに対応しています。

```
go install github.com/kznrluk/aski
```

## 使い方

```bash
$ aski
```

このコマンドは、対話的なChatGPTセッションを開始し、標準入力からユーザーのインプットを受け取ります。
特に指定がなければデフォルトの汎用的なプロファイルを使用します。

## オプション

```
- `-h, --help`: ヘルプメッセージを表示します。
- `-p, --profile`: この会話で使用するプロファイルを選択します。プロファイルは、.aski/config.yamlファイルで定義されます。
- `-f, --file`: 会話とともに送信するファイルを指定します。
- `-c, --content`: 対話モードを利用せず、引数のコンテンツの回答を出力してプログラムを終了します。他アプリケーションとの連携に便利です。
- `-s, --system`: ChatGPTに渡すシステムコンテキストをオーバーライドするオプションです。
- `-r, --rest`: REST APIで通信します。ストリーミングが不安定な場合や、適切な応答が受信できない場合に便利です。
```

## ファイルの内容を扱う

`file` は、指定されたファイルの内容をユーザーコンテキストとしてChatGPTに送信するためのオプションです。

```bash
$ aski -f file1　-f file2 -f file3 ...
```

このオプションを使用することで、指定したファイルの内容をすべてChatGPTに送信することができます。

```bash
$ echo -e "Hello,\nWorld!" > hello.txt
$ aski -f hello.txt
```

上記の例では、"Hello,\nWorld!"というコンテンツを持つファイルhello.txtからコンテキストが送信されます。次にコンテキストは、ChatGPTを介して応答ストリームに送信されます。

ファイルは、パターン検索を使用して複数渡すことができます。例えば、以下のコマンドでは、現在のディレクトリからすべての.txtファイルのコンテンツを送信します。

```bash
$ aski -f *.txt

# 特定のファイルのみを送信することもできます。
$ aski -f hello.txt -f world.txt ...
```

## プロファイル

プロファイルを使用することで、異なる会話コンテキストを簡単に切り替えることができます。プロファイルには、以下の機能があります。

- ユーザー名
- システムコンテキスト
- デフォルトで送信する最初のユーザーメッセージ

UserMessagesやSystemContextに必要なメッセージを追加しておけば、Aski起動時にそれを読み込み、自動的にChatGPTに伝えます。

```yaml
OpenAIAPIKey:
Profiles:
  - ProfileName: Default
    UserName: AskiUser
    Current: true
    SystemContext: You are a kind and helpful chat AI. Sometimes you may say things that are incorrect, but that is unavoidable.
    UserMessages: []
  - ProfileName: Emoji
    UserName: AskiUser
    Current: false
    SystemContext: |
      You are a kind and helpful chat AI.
      Sometimes you may say things that are incorrect, but that is unavoidable.
    UserMessages:
      - |
        Please use a lot of emojis!

```

デフォルトで使用されるプロファイルは、Currentの値を true にするか、下記コマンドで変更できます。
```
aski profile
```

## ライセンス

MIT