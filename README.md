# ASKI

[![Go Report Card](https://goreportcard.com/badge/github.com/kznrluk/aski)](https://goreportcard.com/report/github.com/kznrluk/aski)

`aski` は、ターミナルからChatGPTを扱うことのできるミニマルなアプリケーションです。

## 機能

- ゼロコンフィグ
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

次のようなフラグを使用してプロファイルを指定することもできます。

```bash
$ aski --profile <profile_name>
```

REST APIを使用するには、次のようなフラグを使用します。

```bash
$ aski --rest
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

## ライセンス

MIT