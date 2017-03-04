package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

var Version = "0.0.1"

const checkmark = "\u2714 "

type Todo struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Tag   string `json:"tag"`
	Date  string `json:"date"`
	Memo  string `json:"memo"`
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
					Usage: "add `TAG` to todo (ex: gtd add -t memo task)",
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
	var todos []Todo

	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return err
	}

	if c.Args().Present() {
		title = c.Args().First()
	} else {
		cli.ShowCommandHelp(c, "add")
		return errors.New("Failed to parse options")
	}

	tag = c.String("tag")
	date := time.Now().Format("2006-01-02")
	todo := Todo{Title: title, Done: false, Date: date, Tag: tag}
	todos = append(todos, todo)

	if err := todoWrite("/var/tmp/todo.json", todos); err != nil {
		return err
	}

	return nil
}

func listTodoAction(c *cli.Context) error {
	var todos []Todo
	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return err
	}
	for i, v := range todos {
		if v.Done {
			fmt.Print(checkmark)
		}
		fmt.Printf("%v: %v: %v \n", i, v.Title, v.Date)
	}
	return nil
}

func doneTodoAction(c *cli.Context) error {
	var todonum int
	var todos []Todo
	if c.Args().Present() {
		donenum, err := strconv.Atoi(c.Args().First())
		if err != nil {
			return err
		}
		todonum = donenum
	} else {
		return fmt.Errorf("Failed to parse options")
	}
	todos, err := todoRead("/var/tmp/todo.json")
	if err != nil {
		return err
	}
	for i, _ := range todos {
		if todonum == i {
			todos[i].Done = true
		}
	}
	if err := todoWrite("/var/tmp/todo.json", todos); err != nil {
		return err
	}

	return nil
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
