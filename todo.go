package gtd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

const (
	todoMark     = "\u2610 "
	todoDoneMark = "\u2713 "
	memoMark     = " \u270F"
)

type Todos struct {
	Todos []Todo `json:"todos"`
}

type Todo struct {
	Title    string `json:"title"`
	Done     bool   `json:"done"`
	Tag      string `json:"tag"`
	Date     string `json:"date"`
	Memo     string `json:"memo"`
	Children []Todo `json:"children"`
}

func (todos *Todos) SetTodos(todolist []Todo) {
	todos.Todos = todolist
}

func (todos *Todos) marshallJson(path string) error {
	indent := strings.Repeat(" ", 4)
	json, err := json.MarshalIndent(todos, "", indent)
	if err != nil {
		return fmt.Errorf("Failed to encode json: %v", err)
	}
	ioutil.WriteFile(path, json, 0644)
	return nil
}

func (todos *Todos) UnmarshallJson(path string) error {
	file, err := os.Stat(path)
	if err != nil || file.Size() == 0 {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("Failed to create file: %v", err)
		}
		defer f.Close()
		_, err = f.Write([]byte("{}"))
		if err != nil {
			return fmt.Errorf("Failed to write file: %v", err)
		}
	}
	if data, err := ioutil.ReadFile(path); err != nil {
		return fmt.Errorf("Failed to read jsonfile %v", err)
	} else {
		if err := json.Unmarshal(data, todos); err != nil {
			return fmt.Errorf("Failed to decode data %v", err)
		}
		return nil
	}
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

func appendTodo(todos []Todo, todo Todo, todonumlist []int) ([]Todo, error) {
	if len(todonumlist) == 0 {
		return append(todos, todo), nil
	}
	if len(todos) == 0 || len(todos)-1 < todonumlist[0] {
		return todos, fmt.Errorf("Faild to access index")
	}
	if len(todonumlist) == 1 {
		todos[todonumlist[0]].Children = append(todos[todonumlist[0]].Children, todo)
		return todos, nil
	} else {
		todochildren, err := appendTodo(todos[todonumlist[0]].Children, todo, todonumlist[1:])
		if err != nil {
			return todos, err
		}
		todos[todonumlist[0]].Children = todochildren
		return todos, nil
	}
}

func displayTodo(todos []Todo, tab string, todoid string) {
	for id, todo := range todos {
		if todo.Done {
			continue
		}
		fmt.Print(tab, todoMark)
		if todo.Tag != "" {
			fmt.Printf("%v: %v (%v)", todoid+strconv.Itoa(id), todo.Title, todo.Tag)
		} else {
			fmt.Printf("%v: %v", todoid+strconv.Itoa(id), todo.Title)
		}
		if todo.Memo != "" {
			fmt.Print(memoMark)
		}
		fmt.Println()
		if todo.Children != nil {
			parentid := todoid + strconv.Itoa(id) + "."
			displayTodo(todo.Children, tab+" ", parentid)
		}
	}
}

func displayAllTodo(todos []Todo, tab string, todoid string) {
	for id, todo := range todos {
		if todo.Done {
			fmt.Print(tab, todoDoneMark)
		} else {
			fmt.Print(tab, todoMark)
		}
		if todo.Tag != "" {
			fmt.Printf("%v: %v (%v)", todoid+strconv.Itoa(id), todo.Title, todo.Tag)
		} else {
			fmt.Printf("%v: %v", todoid+strconv.Itoa(id), todo.Title)
		}
		if todo.Memo != "" {
			fmt.Print(memoMark)
		}
		fmt.Println()
		if todo.Children != nil {
			parentid := todoid + strconv.Itoa(id) + "."
			displayAllTodo(todo.Children, tab+" ", parentid)
		}
	}
}

func doneTodo(todos []Todo, todonumlist []int) ([]Todo, error) {
	if len(todos) == 0 || len(todos)-1 < todonumlist[0] {
		return todos, fmt.Errorf("Faild to access index")
	}
	if len(todonumlist) == 1 {
		todos[todonumlist[0]].Done = !todos[todonumlist[0]].Done
		return todos, nil
	} else {
		todochildren, err := doneTodo(todos[todonumlist[0]].Children, todonumlist[1:])
		if err != nil {
			return todos, err
		}
		todos[todonumlist[0]].Children = todochildren
		return todos, nil
	}
}

func displayTagTodo(todos []Todo, tag string, todoid string) {
	for id, todo := range todos {
		if todo.Tag == tag {
			fmt.Printf("%v: %v: %v (%v)", todoid+strconv.Itoa(id), todo.Title, todo.Date, todo.Tag)
			if todo.Memo != "" {
				fmt.Print(memoMark)
			}
			fmt.Println()
		}
		if todo.Children != nil {
			parentid := todoid + strconv.Itoa(id) + "."
			displayTagTodo(todo.Children, tag, parentid)
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
			todos_n = removeTodo(todos, id)
			return todos_n, nil
		}
		if todo.Children != nil {
			todos[id].Children, _ = cleanAllTodos(todos[id].Children)
		}
	}
	return todos, nil
}

func deleteTodos(todos []Todo, todonumlist []int) ([]Todo, error) {
	if len(todos) == 0 || len(todos)-1 < todonumlist[0] {
		return todos, fmt.Errorf("Faild to access index")
	}
	if len(todonumlist) == 1 {
		todos = removeTodo(todos, todonumlist[0])
		return todos, nil
	} else {
		todochildren, err := doneTodo(todos[todonumlist[0]].Children, todonumlist[1:])
		if err != nil {
			return todos, err
		}
		todos[todonumlist[0]].Children = todochildren
		return todos, nil
	}
}

func searchTodo(todos []Todo, todonumlist []int) (*Todo, error) {
	if len(todos) == 0 || len(todos)-1 < todonumlist[0] {
		return nil, fmt.Errorf("Faild to access index")
	}
	if len(todonumlist) == 1 {
		return &todos[todonumlist[0]], nil
	} else {
		return searchTodo(todos[todonumlist[0]].Children, todonumlist[1:])
	}
}

func removeTodo(todos []Todo, todonum int) []Todo {
	todos = append(todos[:todonum], todos[todonum+1:]...)
	n := make([]Todo, len(todos))
	copy(n, todos)
	return n
}
