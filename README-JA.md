# ASKI - ChatGPT Client for Console

[![Go Report Card](https://goreportcard.com/badge/github.com/kznrluk/aski)](https://goreportcard.com/report/github.com/kznrluk/aski)

`aski` は、コンソールでChatGPTを利用できるミニマルなクライアントです。

## 機能
- GPT4対応
- 会話履歴の保存と復元
- 任意時点の会話へ移動
- GLOBによるファイル添付機能
- プロファイル機能 異なる会話コンテキストを簡単に切り替え可能
- Streaming APIとREST APIのサポート

## インストール
今のところ `go install` のみに対応しています。

```
go install github.com/kznrluk/aski@main
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
- `-r, --restore`: 会話履歴をヒストリファイルから復元します。このオプションを使用すると、以前の会話を続けることができます。前方一致。
- `--rest`: REST APIで通信します。ストリーミングが不安定な場合や、適切な応答が受信できない場合に便利です。
```

## インラインコマンド

```
  :history - 会話履歴を表示します
  :summary - 会話の概要があれば表示します プロファイルでSummarizeがTrueになっている必要があります
  :move    - 次に投稿する会話の親となるメッセージを指定します。
  :config  - 設定やプロファイル、履歴の含まれているフォルダを開きます。
  :editor  - 外部エディタを利用します
  :exit    - プログラムを終了します
```

`:exit` 以外のコマンドは、前方一致で検索されます。例えば、`:h` と入力すると `:history` が実行されます。

## 外部エディタの利用
コンソールでの入力が難しい改行付きのプロンプトや、長文の入力を行う場合は、外部エディタを使用することができます。

```
aski@GPT4> :editor
```

EDITOR環境変数に設定されたエディタが起動します。エディタを終了すると、入力された内容がChatGPTに送信されます。Windowsの場合はデフォルトでnotepad、macOS, Linuxの場合はデフォルトでvimが起動します。

## ファイルの内容を扱う

`file` は、指定されたファイルの内容をユーザーコンテキストとしてChatGPTに送信するためのオプションです。

```bash
$ aski -f file1 -f file2 -f file3 ...
```

このオプションを使用することで、指定したファイルの内容をすべてChatGPTに送信することができます。

```bash
$ echo -e "Hello,\nWorld!" > hello.txt
$ aski -f hello.txt
```

上記の例では、"Hello,\nWorld!"というコンテンツを持つファイルhello.txtが会話に含まれて送信されるようになります。

ファイルは、パターン検索を使用して複数渡すことができます。例えば、以下のコマンドでは、現在のディレクトリからすべての.txtファイルのコンテンツを送信します。

```bash
$ aski -f *.txt

# 特定のファイルのみを送信することもできます。
$ aski -f hello.txt -f world.txt ...
```

## プロファイル

プロファイルを使用することで、異なる会話コンテキストや設定簡単に切り替えることができます。プロファイルには、以下の機能があります。

**UserName**

ユーザー名です。ChatGPTに送信されるメッセージの発言者として送信されます。

**Model**

使用するモデルの名前です。OpenAIのAPIで利用できる値である必要があります。

[Models - OpenAI API](https://platform.openai.com/docs/models/chatgpt)

**Current**

現在のプロファイルかどうかを示します。trueに設定されているプロファイルが使用されます。

**AutoSave**

会話履歴を自動的に保存するかどうかを示します。trueに設定されているプロファイルは、会話履歴を自動的に保存します。

**Summarize**

会話の概要を表示するかどうかを示します。trueに設定されているプロファイルで会話を始めた際、会話の概要をGPT3.5で生成します。

**SystemContext**

ChatGPTに送信されるシステムコンテキストです。会話の最初に送信され、どのような会話をしてほしいかをChatGPTに伝えます。

**UserMessages**

ChatGPTに送信されるユーザーコンテキストです。会話の最初に送信されます。SystemContextに含めたくない場合に利用します。


UserMessagesやSystemContextに必要なメッセージを追加しておけば、Askiは起動時にそれを読み込み、自動的にChatGPTに伝えます。

```yaml
OpenAIAPIKey:
Profiles:
  - ProfileName: Default
    UserName: AskiUser
    Model: gpt-3.5-turbo
    Current: true
    AutoSave: true
    Summarize: true
    SystemContext: You are a kind and helpful chat AI. Sometimes you may say things that are incorrect, but that is unavoidable.
    UserMessages: []
  - ProfileName: Emoji
    UserName: AskiUser
    Model: gpt-3.5-turbo
    Current: false
    AutoSave: true
    Summarize: true
    SystemContext: |
      You are a kind and helpful chat AI.
      Sometimes you may say things that are incorrect, but that is unavoidable.
    UserMessages:
      - |
        Please use a lot of emojis!

```

SystemContextは常に最初に送信され、UserMessagesはその次に送信されます。ファイルを指定した際はSystemContextとUserMessagesの間にファイルの情報が添付されます。

デフォルトで使用されるプロファイルは、Currentの値を true にするか、下記コマンドで変更できます。
```
aski profile
```

## ライセンス

MIT