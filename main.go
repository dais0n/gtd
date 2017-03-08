package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/urfave/cli"
)

var Version = "0.0.1"

const (
	todo_mark      = "\u2610 "
	todo_done_mark = "\u2611 "
)

type Todo struct {
	Title    string `json:"title"`
	Done     bool   `json:"done"`
	Tag      string `json:"tag"`
	Date     string `json:"date"`
	Memo     string `json:"memo"`
	Children []Todo `json:"children"`
}

func main() {
	os.Exit(Run(os.Args))
}

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
			Aliases: []string{"n"},
			Action:  addTodoAction,
			Usage:   "add todo",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "parent, p",
					Usage: "specify parent `TODO_NUM` (ex: gtd add -p 1 task)",
				},
				cli.StringFlag{
					Name:  "tag, t",
					Usage: "add `TAG` to todo (ex: gtd add -t memo task)"}}}, {
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
			Usage:   "list todo",
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
	}
	app.Run(os.Args)
	return 0
}

func addTodoAction(c *cli.Context) error {
	var title string
	var tag string
	var parent string
	var todos []Todo

	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return fmt.Errorf("Failed to read todofile: %v", err)
	}

	if c.Args().Present() {
		title = c.Args().First()
		tag = c.String("tag")
		parent = c.String("parent")
	} else {
		cli.ShowCommandHelp(c, "add")
		return fmt.Errorf("Failed to parse options")
	}

	todo := Todo{Title: title, Done: false, Tag: tag}

	parentnums, err := parseTodoNum(parent)
	todos, err = appendTodo(todos, todo, parentnums)

	return todoWrite("/var/tmp/todo.json", todos)
}

func listTodoAction(c *cli.Context) error {
	var todos []Todo

	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return err
	}

	fmt.Println("\u2705  YOUR TO DO LIST")

	if c.Bool("all") {
		displayAllTodo(todos, "")
		return nil
	}

	displayTodo(todos, "")
	return nil
}

func doneTodoAction(c *cli.Context) error {
	var todos []Todo

	if !c.Args().Present() {
		return fmt.Errorf("Failed to parse options")
	}

	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return fmt.Errorf("Failed to read jsonfile: %v")
	}

	todonum := c.Args().First()
	todonumlist, err := parseTodoNum(todonum)
	todos, err = doneTodo(todos, todonumlist)
	if err != nil {
		return fmt.Errorf("Failed to read jsonfile: %v", err)
	}

	if err := todoWrite("/var/tmp/todo.json", todos); err != nil {
		return err
	}

	return nil
}

func tagTodoAction(c *cli.Context) error {
	var todos []Todo

	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return err
	}

	if c.Bool("all") {
		var tags = mapset.NewSet()
		displayAllTags(todos, tags)
	} else if tag := c.Args().First(); tag != "" {
		displayTagTodo(todos, tag)
	} else {
		cli.ShowCommandHelp(c, "list")
		return fmt.Errorf("Failed to parse options")
	}

	return nil
}

// parse "1.1.0" to [1,1,0]
func parseTodoNum(todonum string) ([]int, error) {
	var todonums []int
	for _, v := range strings.Split(todonum, ".") {
		if id, err := strconv.Atoi(v); err != nil {
			return nil, err
		} else {
			todonums = append(todonums, id)
		}
	}
	return todonums, nil
}

func todoWrite(path string, todos []Todo) error {
	indent := strings.Repeat(" ", 4)
	json, err := json.MarshalIndent(todos, "", indent)
	if err != nil {
		return fmt.Errorf("Failed to encode json: %v", err)
	}
	ioutil.WriteFile("/var/tmp/todo.json", json, 0644)
	return nil
}

func todoRead(path string) ([]Todo, error) {
	var todos []Todo
	if data, err := ioutil.ReadFile(path); err != nil {
		return nil, fmt.Errorf("Failed to read jsonfile %v", err)
	} else {
		if err := json.Unmarshal(data, &todos); err != nil {
			return nil, fmt.Errorf("Failed to decode data %v", err)
		}
		return todos, nil
	}
}

func appendTodo(todos []Todo, todo Todo, todonum []int) ([]Todo, error) {
	if len(todonum) == 0 {
		return append(todos, todo), nil
	}
	if len(todonum) == 1 {
		todos[todonum[0]].Children = append(todos[todonum[0]].Children, todo)
		return todos, nil
	} else {
		todos_children, _ := appendTodo(todos[todonum[0]].Children, todo, todonum[1:])
		todos[todonum[0]].Children = todos_children
		return todos, nil
	}
}

func displayTodo(todos []Todo, tab string) {
	for id, todo := range todos {
		if todo.Done {
			continue
		}
		fmt.Print(tab, todo_mark)
		fmt.Printf("%v: %v: %v (%v)\n", id, todo.Title, todo.Date, todo.Tag)
		if todo.Children != nil {
			displayTodo(todo.Children, tab+" ")
		}
	}
}

func displayAllTodo(todos []Todo, tab string) {
	for id, todo := range todos {
		if todo.Done {
			fmt.Print(tab, todo_done_mark)
		} else {
			fmt.Print(tab, todo_mark)
		}
		fmt.Printf("%v: %v: %v (%v)\n", id, todo.Title, todo.Date, todo.Tag)
		if todo.Children != nil {
			displayAllTodo(todo.Children, tab+" ")
		}
	}
}

func doneTodo(todos []Todo, todonum []int) ([]Todo, error) {
	if len(todonum) == 1 {
		todos[todonum[0]].Done = true
		return todos, nil
	} else {
		todos_children, _ := doneTodo(todos[todonum[0]].Children, todonum[1:])
		todos[todonum[0]].Children = todos_children
		return todos, nil
	}
}

func displayTagTodo(todos []Todo, tag string) {
	for id, todo := range todos {
		if todo.Tag == tag {
			fmt.Printf("%v: %v: %v (%v)\n", id, todo.Title, todo.Date, todo.Tag)
		}
		if todo.Children != nil {
			displayTagTodo(todo.Children, tag)
		}
	}
}

func displayAllTags(todos []Todo, tags mapset.Set) {
	for _, todo := range todos {
		add := tags.Add(todo.Tag)
		if todo.Tag != "" && add {
			fmt.Println(todo.Tag)
		}
		if todo.Children != nil {
			displayAllTags(todo.Children, tags)
		}
	}
}
