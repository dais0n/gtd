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
	todoMark     = "\u2610 "
	todoDoneMark = "\u2611 "
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
	}
	app.Run(os.Args)
	return 0
}

func addTodoAction(c *cli.Context) error {
	var title string
	var tag string
	var parentnum string
	var memo string

	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return fmt.Errorf("Failed to read todofile: %v", err)
	}

	if c.Args().Present() {
		title = c.Args().First()
		tag = c.String("tag")
		parentnum = c.String("parent")
		memo = c.String("memo")
	} else {
		cli.ShowCommandHelp(c, "add")
		return fmt.Errorf("Failed to parse options")
	}

	todo := Todo{Title: title, Done: false, Tag: tag, Memo: memo}

	parentnumlist, err := parseTodoNum(parentnum)
	todos, err = appendTodo(todos, todo, parentnumlist)

	return todoWrite("/var/tmp/todo.json", todos)
}

func listTodoAction(c *cli.Context) error {
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

func tagTodoAction(c *cli.Context) error {
	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}

	if c.Bool("all") {
		tags := mapset.NewSet()
		displayAllTags(todos, tags)
	} else if tag := c.Args().First(); tag != "" {
		displayTagTodo(todos, tag)
	} else {
		cli.ShowCommandHelp(c, "list")
		return fmt.Errorf("Failed to parse options")
	}

	return nil
}

func doneTodoAction(c *cli.Context) error {
	var todos []Todo

	if !c.Args().Present() {
		return fmt.Errorf("Failed to parse options")
	}

	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}

	todonum := c.Args().First()
	todonumlist, err := parseTodoNum(todonum)
	todos, err = doneTodo(todos, todonumlist)
	if err != nil {
		return fmt.Errorf("Failed to read jsonfile: %v", err)
	}

	if err := todoWrite("/var/tmp/todo.json", todos); err != nil {
		return fmt.Errorf("Failed to write jsonfile: %v", err)
	}

	return nil
}

func cleanTodoAction(c *cli.Context) error {
	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return fmt.Errorf("failed to read jsonfile: %v", err)
	}
	todos, err = cleanAllTodos(todos)
	if err := todoWrite("/var/tmp/todo.json", todos); err != nil {
		return fmt.Errorf("Failed to write jsonfile: %v", err)
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

func appendTodo(todos []Todo, todo Todo, todonumlist []int) ([]Todo, error) {
	if len(todonumlist) == 0 {
		return append(todos, todo), nil
	}
	if len(todonumlist) == 1 {
		todos[todonumlist[0]].Children = append(todos[todonumlist[0]].Children, todo)
		return todos, nil
	} else {
		todochildren, _ := appendTodo(todos[todonumlist[0]].Children, todo, todonumlist[1:])
		todos[todonumlist[0]].Children = todochildren
		return todos, nil
	}
}

func displayTodo(todos []Todo, tab string) {
	for id, todo := range todos {
		if todo.Done {
			continue
		}
		fmt.Print(tab, todoMark)
		fmt.Printf("%v: %v: %v (%v)\n", id, todo.Title, todo.Date, todo.Tag)
		if todo.Children != nil {
			displayTodo(todo.Children, tab+" ")
		}
	}
}

func displayAllTodo(todos []Todo, tab string) {
	for id, todo := range todos {
		if todo.Done {
			fmt.Print(tab, todoDoneMark)
		} else {
			fmt.Print(tab, todoMark)
		}
		fmt.Printf("%v: %v: %v (%v)\n", id, todo.Title, todo.Date, todo.Tag)
		if todo.Children != nil {
			displayAllTodo(todo.Children, tab+" ")
		}
	}
}

func doneTodo(todos []Todo, todonumlist []int) ([]Todo, error) {
	if len(todonumlist) == 1 {
		todos[todonumlist[0]].Done = true
		return todos, nil
	} else {
		todochildren, _ := doneTodo(todos[todonumlist[0]].Children, todonumlist[1:])
		todos[todonumlist[0]].Children = todochildren
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

func cleanAllTodos(todos []Todo) ([]Todo, error) {
	for id, todo := range todos {
		var todos_n []Todo
		if todo.Done == true {
			todos_n = deleteTodos(todos, id)
			return todos_n, nil
		}
		if todo.Children != nil {
			todos[id].Children, _ = cleanAllTodos(todos[id].Children)
		}
	}
	return todos, nil
}

func deleteTodos(todos []Todo, todonum int) []Todo {
	todos = append(todos[:todonum], todos[todonum+1:]...)
	n := make([]Todo, len(todos))
	copy(n, todos)
	return n
}
