# aski - ChatGPT / Claude Client for Terminal

![using aski](https://raw.githubusercontent.com/kznrluk/aski/main/docs/use.gif)

[![Go Report Card](https://goreportcard.com/badge/github.com/kznrluk/aski)](https://goreportcard.com/report/github.com/kznrluk/aski)

`aski` は、ターミナルで利用できる高機能なChatGPTクライアントです。通常の会話だけでなく、会話履歴の保存や復元、任意の会話への移動や編集など、様々な機能を備えています。

## 機能
- Go言語によるマルチプラットフォーム対応 & シングルバイナリ
- PowerShellやTerminalでの利用が可能
- GPT-4対応
- Claude3対応
- 会話履歴の保存と復元
- 任意時点の会話へ移動
- GLOBによるファイル添付機能
- プロファイル機能 異なる会話コンテキストを簡単に切り替え可能
- Streaming APIとREST APIのサポート

## インストール
リリースページからバイナリをダウンロードできます。

[Releases · kznrluk/aski](https://github.com/kznrluk/aski/releases)

もしくは `go install` コマンドを実行してください。

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
- `-h, --help`    : ヘルプメッセージを表示します。
- `-p, --profile` : この会話で使用するプロファイルを選択します。
                    プロファイルは.aski/profilesディレクトリ内のファイル名を指定するか、任意の場所のYAMLファイルを直接指定することができます。
- `-f, --file`    : 会話とともに送信するファイルを指定します。
- `-c, --content` : 対話モードを利用せず、引数のコンテンツの回答を出力してプログラムを終了します。他アプリケーションとの連携に便利です。
- `-r, --restore` : 会話履歴をヒストリファイルから復元します。このオプションを使用すると、以前の会話を続けることができます。前方一致。
- `-m, --model`   : 使用するモデルを指定します。OpenAIのAPIで利用できる値である必要があります。
                    [Models - OpenAI API](https://platform.openai.com/docs/models/chatgpt)
                    Claude3を使用する場合は `claude-3-opus-20240229` を指定します。
- `--rest`        : REST APIで通信します。ストリーミングが不安定な場合や、適切な応答が受信できない場合に便利です。
```

## インラインコマンド

![history copmmand](https://raw.githubusercontent.com/kznrluk/aski/main/docs/history.png)

```
> :

  :history       - 会話の履歴を表示します。
  :move          - 別のメッセージへのHEADを変更します。
  :config        - 設定ディレクトリを開きます。
  :editor        - 新しいメッセージを追加するために外部テキストエディタを開きます。
  :editor sha1   - 引数のメッセージを編集し、会話を続けます。
  :editor latest - HEADから一番近い自分の発言を編集します。
  :modify sha1   - 過去の会話を変更します。HEADは移動しません。
                   次回送信から過去の会話が変更されます。
  :param         - プロファイルのカスタムパラメータの値を確認したり書き換えたりします。
                   通常の使用では変更する必要はありません。
  :exit          - プログラムを終了します。
```

`:exit` 以外のコマンドは、前方一致で検索されます。例えば、`:h` と入力すると `:history` が実行されます。

## 外部エディタの利用

![external editor](https://raw.githubusercontent.com/kznrluk/aski/main/docs/editor.gif)

コンソールでの入力が難しい改行付きのプロンプトや、長文の入力を行う場合は、外部エディタを使用することができます。

```
aski@GPT4> :editor
```

EDITOR環境変数に設定されたエディタが起動します。エディタを終了すると、入力された内容がChatGPTに送信されます。Windowsの場合はデフォルトでnotepad、macOS, Linuxの場合はデフォルトでvimが起動します。

また、環境変数をcodeに変更すればVSCodeでも編集できます。

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

## Pipe

askiは*nix系のシェルでのパイプ入力に対応しています。

```
$ cat test.txt
Linuxにおける
Pipeって
なんですか？

$ cat test.txt | aski
LinuxのPipe（パイプ）は、UNIX系オペレーティングシステム（Linux, macOSなど）のコンセプトで、複数のコマンドを連結して、あるコマンドの出力を別のコマンドの入力として渡すことができます。😊
主にシェル（bash, zsh, etc.）で使用されます。パイプは、縦棒 (`|`) を使ってコマンド間でデータをストリーム化して連携させます。💻
```

## 設定と会話ヒストリ
askiが利用するファイルは基本的にホームディレクトリ直下の `.aski` ディレクトリに配置されています。

- Windows: `C:\Users\your_name\.aski`
- macOS: `/Users/your_name/.aski`
- Linux: `/home/your_name/.aski`

設定や会話ログはYAMLファイルとして保存されます。テキストエディタで編集し、好みの動作に変更することができます

### コンフィグファイル
コンフィグファイルには、現在のプロファイルとOpenAI APIキーが含まれています。プロファイルは `profile` ディレクトリに保存されているyamlファイルです。

```yaml
OpenAIAPIKey: sk-Bs.....................
AnthropicAPIKey: sk-.....................
CurrentProfile: gpt4.yaml
```

### プロファイル

プロファイルを使用することで、異なる会話コンテキストや設定簡単に切り替えることができます。プロファイルには、以下の機能があります。

**UserName**

ユーザー名です。ChatGPTに送信されるメッセージの発言者として送信されます。

**Model**

使用するモデルの名前です。OpenAIのAPIで利用できる値である必要があります。

[Models - OpenAI API](https://platform.openai.com/docs/models/chatgpt)

**AutoSave**

会話履歴を自動的に保存するかどうかを示します。trueに設定されているプロファイルは、会話履歴を自動的に保存します。

**ResponseFormat**

`text` か `json_object` を指定します。 `text` を指定した場合、ChatGPTは通常のテキスト形式で応答を行います。 `json_object` を指定し、プロンプトに `json` を含めて送信した場合、ChatGPTは有効なJSONオブジェクト形式で応答を行います。

**SystemContext**

ChatGPTに送信されるシステムコンテキストです。会話の最も最初に送信され、どのような会話をしてほしいかをChatGPTに伝えます。

**Messages**

ChatGPTに送信されるコンテキストです。会話の開始前に自動的に送信されます。 `user` 以外にも、 `assistant`、 `system` が利用できます。

UserMessagesやSystemContextに必要なメッセージを追加しておけば、Askiは起動時にそれを読み込み、自動的にChatGPTに伝えます。

**CustomParameters**

ChatGPTに送信する際に使用されるパラメーターを上書きします。キーが指定されていないかゼロ値の場合、APIによるデフォルト値が利用されます。
利用可能なパラメータはChatGPTのAPIリファレンスを参照してください。通常の利用では変更する必要はありません。また、パラメーター `N` の変更には現在のところ対応していません。

[API Reference - OpenAI API](https://platform.openai.com/docs/api-reference/chat/create)

```yaml
ProfileName: Default
UserName: AskiUser
Model: gpt-3.5-turbo
AutoSave: true
Summarize: true
SystemContext: You are a kind and helpful chat AI. Sometimes you may say things that are incorrect, but that is unavoidable.
Messages:
  - Role: user
    Content: Hi, nice to meet you!
  - Role: assistant
    Content: Hi, What's your name?
  - Role: user
    Content: My name is Aski.
CustomParameters:
  temperature: 1
  stop: ["hello"]
```

SystemContextは常に最初に送信され、UserMessagesはその次に送信されます。ファイルを指定した際はSystemContextとUserMessagesの間にファイルの情報が添付されます。

デフォルトで使用されるプロファイルは、Currentの値を true にするか、下記コマンドで変更できます。
```
aski profile
```

## ライセンス

MIT