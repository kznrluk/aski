# Aski - ChatGPT Client for Terminal

![using aski](https://raw.githubusercontent.com/kznrluk/aski/main/docs/use.gif)

[![Go Report Card](https://goreportcard.com/badge/github.com/kznrluk/aski)](https://goreportcard.com/report/github.com/kznrluk/aski)

[🇯🇵 README-JA.md here!](https://github.com/kznrluk/aski/blob/main/README-JA.md)

`aski` is a feature-rich ChatGPT client for the terminal. In addition to regular conversations, it offers various features such as saving and restoring conversation history, moving to and editing arbitrary conversations.

## Features
- Written in Go, so it can be used on any platform that supports Go.
- Works in the shell, compatible with PowerShell and Terminal.
- GPT4 supported
- Save and restore conversation history
- Move to any point in the conversation
- File attachment with GLOB support
- Profile feature for easily switching between different conversation contexts
- Support for Streaming API and REST API

## Installation
Currently, only `go install` is supported.

```
go install github.com/kznrluk/aski@main
```

## Usage

```bash
$ aski
```

This command starts an interactive ChatGPT session and takes user input from stdin.
By default, it uses a generic profile, unless otherwise specified.

## Options

```
- `-h, --help`: Displays help message.
- `-p, --profile`: Choose the profile to use for this conversation. Profiles are defined in the .aski/config.yaml file.
- `-f, --file`: Specifies a file to send with the conversation.
- `-c, --content`: Outputs the answer for the content of the argument without using the interactive mode and ends the program. Useful for integration with other applications.
- `-r, --restore`: Restores the conversation history from a history file. With this option, you can continue a previous conversation. Forward match.
- `--rest`: Communicate with the REST API. Useful when streaming is unstable or appropriate responses cannot be received.
```

## Inline Commands

![history copmmand](https://raw.githubusercontent.com/kznrluk/aski/main/docs/history.png)


```
> :

  :history       - Show conversation history.
  :summary       - Show conversation summary.
  :move          - Change HEAD to another message.
  :config        - Open configuration directory.
  :editor        - Open an external text editor to add new message.
  :editor sha1   - Edit the argument message and continue the conversation.
  :editor latest - Edits the nearest own statement from HEAD.
  :modify sha1   - Modify the past conversation. HEAD does not move.
                   Past conversations will be modified from the next transmission.
  :param         - Check or overwrite the values of custom parameters in the profile.
                   It is not necessary to change them in general use.
  :exit          - Exit the program.
```

All commands except `:exit` are searched by forward match. For example, typing `:h` will execute `:history`.

## Using an External Editor

![external editor](https://raw.githubusercontent.com/kznrluk/aski/main/docs/editor.gif)

For prompts with line breaks or long input that is difficult to input on the console, you can use an external editor.

```
aski@GPT4> :editor
```

The editor set in the EDITOR environment variable will start. Once the editor is closed, the content entered will be sent to ChatGPT. The default is notepad on Windows and vim on macOS and Linux.

Additionally, by changing the environment variable to 'code', it is possible to edit using VS Code as well.

## Handling File Content

`file` is an option to send the contents of the specified file as user context to ChatGPT.

```bash
$ aski -f file1 -f file2 -f file3 ...
```

With this option, you can send the contents of all the specified files to ChatGPT.

```bash
$ echo -e "Hello,\nWorld!" > hello.txt
$ aski -f hello.txt
```

In the example above, the hello.txt file with the content "Hello,\nWorld!" will be included in the conversation and sent.

Files can be passed in multiples using pattern search. For example, the following command sends the contents of all .txt files in the current directory.

```bash
$ aski -f *.txt

# You can also send specific files.
$ aski -f hello.txt -f world.txt ...
```

## Pipe

aski supports pipe input in *nix based shells.

```
$ cat test.txt
What is 
a pipe in
Linux?

$ cat test.txt | aski
A pipe in Linux is a concept in UNIX-based operating systems (such as Linux and macOS) that allows you to chain multiple commands together so that the output of one command becomes the input of another.😊
It is mainly used in shells (bash, zsh, etc.). Pipes use a vertical bar (`|`) to stream data between commands, enabling them to work together efficiently.💻
```

## Profiles

By using profiles, you can easily switch between different conversation contexts and settings. Profiles have the following features.

**UserName**

The username. It will be sent as the sender of the messages sent to ChatGPT.

**Model**

The name of the model you want to use. It must be a valid value that can be used with the OpenAI API.

[Models - OpenAI API](https://platform.openai.com/docs/models/chatgpt)

**Current**

Indicates whether the profile is currently active. Profiles set to true will be used.

**AutoSave**

Indicates whether to automatically save the conversation history. Profiles set to true will automatically save the conversation history.

**Summarize**

Indicates whether to display the conversation summary. When starting a conversation with a profile set to true, a summary of the conversation will be generated using GPT3.5.

**SystemContext**

The system context that will be sent to ChatGPT. It is sent at the beginning of the conversation to tell ChatGPT what kind of conversation you want to have.

**UserMessages**

These parameters overwrite the ones used when sending data to ChatGPT. If a key is not specified or has a zero value, the default value provided by the API will be used. Please refer to the ChatGPT API Reference for the available parameters. 
In general, there is no need to modify these parameters. Additionally, the program currently does not support changing the `N` parameter.

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
    Messages:
      - Role: user
        Content: Hi, nice to meet you!
      - Role: assistant
        Content: Hi, What's your name?
      - Role: user
        Content: My name is Aski.
  - ProfileName: Emoji
    UserName: AskiUser
    Model: gpt-3.5-turbo
    Current: false
    AutoSave: true
    Summarize: true
    SystemContext: |
      You are a kind and helpful chat AI.
      Sometimes you may say things that are incorrect, but that is unavoidable.
    Messages:
      - Role: user
        Content: Please use a lot of emojis!
      - Role: assistant
        Content: Ok! I will use a lot of emojis!
    CustomParameters:
      temperature: 1
      stop: ["hello"]
```

SystemContext is always sent first, followed by UserMessages. If a file is specified, the file information will be attached between the SystemContext and UserMessages.

The default profile to be used can be changed by setting the value of Current to true, or by using the following command:

```
aski profile
```

## License

MIT