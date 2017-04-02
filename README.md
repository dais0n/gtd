# gtd
gtd = GoToDo

gtd is simple todo list tool on CLI, written in go

## Demo
![gtd demo](https://raw.githubusercontent.com/wiki/dais0n/gtd/screenshot.gif)

## Usage
```
NAME:
   gtd - todo app

USAGE:
   gtd [global options] command [command options] [arguments...]

VERSION:
   0.1.1

AUTHOR:
   Takuya Omura <t.omura8383@gmail.com>

COMMANDS:
     add, a      add todo
     list, l     list todo
     tags, t     list tags
     done, d     done todo
     clean, c    clean done todo
     delete, d   delete todo
     setting, s  edit config file
     memo, m     edit memo file associated with task (ex, gtd memo 4)
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Installation
```
go get -u github.com/dais0n/gtd/cmd/gtd
```

## Configuration
run ``` gtd setting```. this config file is made by gtd command in ${HOME}/.config/gtd/config.toml

```
gtdfile = "/path/to/gtd.json" # this file is todo data file. Default is in ${HOME}/gtd.json
outputdir = "/path/to/output" # this folder is output dir added file by gtd output command
filtercmd = "fzf" # this command is used when you type gtd add -m cmd
editor = "vim" # this editor is used by open memo file and config file
memodir = "~" # search taget below this directory when you add memo
```
