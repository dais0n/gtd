package gtd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/urfave/cli"
)

const (
	ExitCodeOK        int = iota // 0
	ExitCodeError                // 1
	ExitCodeFileError            // 2
	timeFormat        = "2006-01-02"
)

var Version = "0.1.1"

func Run(args []string) int {
	app := cli.NewApp()
	app.Name = "gtd"
	app.Version = Version
	app.Usage = "todo app"
	app.Author = "Takuya Omura"
	app.Email = "t.omura8383@gmail.com"
	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Action:  addTodoAction,
			Usage:   "add todo",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "parent, p",
					Usage: "specify parent `TODO_NUM` (ex: gtd add -p 1 task)",
				},
				cli.StringFlag{
					Name:  "tag, t",
					Usage: "add `TAG` to todo (ex: gtd add -t life task)",
				},
				cli.BoolFlag{
					Name:  "memo, m",
					Usage: "add `MEMO` to todo (ex: gtd add -m task)",
				},
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list todo",
			Action:  listTodoAction,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "show all todos",
				},
			},
		},
		{
			Name:    "tags",
			Aliases: []string{"t"},
			Usage:   "list tags",
			Action:  tagTodoAction,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "show all tags",
				},
			},
		},
		{
			Name:    "done",
			Aliases: []string{"d"},
			Usage:   "done todo",
			Action:  doneTodoAction,
		},
		{
			Name:    "clean",
			Aliases: []string{"c"},
			Usage:   "clean done todo",
			Action:  cleanTodoAction,
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "delete todo",
			Action:  deleteTodoAction,
		},
		{
			Name:    "setting",
			Aliases: []string{"s"},
			Usage:   "edit config file",
			Action:  settingTodoAction,
		},
		{
			Name:    "memo",
			Aliases: []string{"m"},
			Usage:   "edit memo file associated with task (ex, gtd memo 4)",
			Action:  memoTodoAction,
		},
	}
	return msg(app.Run(os.Args))
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		return ExitCodeError
	}
	return ExitCodeOK
}

func ask(prompt string) bool {
	fmt.Print(prompt, ": ")
	var stdin string
	fmt.Scan(&stdin)
	return stdin == "y" || stdin == "Y"
}

func addTodoAction(c *cli.Context) error {
	var title string
	var tag string
	var parentnum string
	var memo string
	var todos Todos

	var cfg config
	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}

	err = todos.UnmarshallJson(cfg.GtdFile)
	if err != nil {
		return fmt.Errorf("Failed to read todofile: %v", err)
	}

	if c.Args().Present() {
		title = c.Args().First()
		tag = c.String("tag")
		parentnum = c.String("parent")
	} else {
		cli.ShowCommandHelp(c, "add")
		return fmt.Errorf("Failed to parse options")
	}

	if c.Bool("memo") {
		var buf bytes.Buffer
		var stdin string
		if cfg.MemoDir == "" {
			return fmt.Errorf("please setting MemoDir ")
		}
		if cfg.FilterCmd == "" {
			return fmt.Errorf("please setting FilterCmd (peco/fzf)")
		}

		if ask("existing file? (y/n)") {
			cmd := fmt.Sprintf("find %s -type f | %s", cfg.MemoDir, cfg.FilterCmd)
			cfg.filtercmd(cmd, &buf)
		} else {
			fmt.Println("new file name: ")
			fmt.Scan(&stdin)
			stdin = "/" + stdin
			cmd := fmt.Sprintf("find %s -type d | %s", cfg.MemoDir, cfg.FilterCmd)
			cfg.filtercmd(cmd, &buf)
		}
		cmdresult := buf.String()
		memo = strings.TrimRight(cmdresult, "\n") + stdin
	}
	var date = time.Now().Format(timeFormat)
	todo := Todo{
		Title: title,
		Done:  false,
		Tag:   tag,
		Memo:  memo,
		Date:  date,
	}

	parentnumlist, err := parseTodoNum(parentnum)
	todolist, err := appendTodo(todos.Todos, todo, parentnumlist)
	if err != nil {
		return fmt.Errorf("Failed to append task: %v", err)
	}
	todos.SetTodos(todolist)

	return todos.marshallJson(cfg.GtdFile)
}

func listTodoAction(c *cli.Context) error {
	var cfg config
	var todos Todos

	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}
	err = todos.UnmarshallJson(cfg.GtdFile)
	if err != nil {
		return fmt.Errorf("Failed to read jsonfile: %v", err)
	}
	fmt.Println("\u2705  YOUR TO DO LIST")

	if c.Bool("all") {
		displayAllTodo(todos.Todos, "", "")
		return nil
	}
	displayTodo(todos.Todos, "", "")
	return nil
}

func tagTodoAction(c *cli.Context) error {
	var cfg config
	var todos Todos
	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}
	err = todos.UnmarshallJson(cfg.GtdFile)
	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}

	if c.Bool("all") {
		tags := mapset.NewSet()
		displayAllTags(todos.Todos, tags)
	} else if tag := c.Args().First(); tag != "" {
		displayTagTodo(todos.Todos, tag, "")
	} else {
		cli.ShowCommandHelp(c, "list")
		return fmt.Errorf("Failed to parse options")
	}
	return nil
}

func doneTodoAction(c *cli.Context) error {
	var todos Todos

	var cfg config
	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}

	if !c.Args().Present() {
		return fmt.Errorf("Failed to parse options")
	}

	err = todos.UnmarshallJson(cfg.GtdFile)
	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}

	todonum := c.Args().First()
	todonumlist, err := parseTodoNum(todonum)
	if err != nil {
		return fmt.Errorf("Failed to parse number: %v", err)
	}
	todolist, err := doneTodo(todos.Todos, todonumlist)
	if err != nil {
		return fmt.Errorf("Failed to done todo: %v", err)
	}
	todos.SetTodos(todolist)

	return todos.marshallJson(cfg.GtdFile)
}

func cleanTodoAction(c *cli.Context) error {
	var cfg config
	var todos Todos
	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}
	err = todos.UnmarshallJson(cfg.GtdFile)
	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}
	todolist, err := cleanAllTodos(todos.Todos)
	todos.SetTodos(todolist)

	return todos.marshallJson(cfg.GtdFile)
}

func deleteTodoAction(c *cli.Context) error {
	var cfg config
	var todos Todos
	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}
	err = todos.UnmarshallJson(cfg.GtdFile)

	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}
	todonum := c.Args().First()
	todonumlist, err := parseTodoNum(todonum)
	if err != nil {
		return fmt.Errorf("Failed to parse number: %v", err)
	}
	todolist, err := deleteTodos(todos.Todos, todonumlist)
	todos.SetTodos(todolist)

	return todos.marshallJson(cfg.GtdFile)
}

func settingTodoAction(c *cli.Context) error {
	var cfg config
	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}
	dir := os.Getenv("HOME")
	dir = filepath.Join(dir, ".config", "gtd")
	file := filepath.Join(dir, "config.toml")
	return cfg.runcmd(cfg.Editor, file)
}

func memoTodoAction(c *cli.Context) error {
	var cfg config
	var todos Todos
	err := cfg.load()
	if err != nil {
		return fmt.Errorf("falid to load configfile: %v", err)
	}

	err = todos.UnmarshallJson(cfg.GtdFile)
	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}

	todonum := c.Args().First()
	todonumlist, err := parseTodoNum(todonum)
	if err != nil {
		return fmt.Errorf("Failed to parse number: %v", err)
	}
	todolist, err := searchTodo(todos.Todos, todonumlist)
	if err != nil {
		return fmt.Errorf("Failed to access todo: %v", err)
	}
	return cfg.runcmd(cfg.Editor, todolist.Memo)
}
