package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Todo struct {
	ID        int       `json:"id"`
	Task      string    `json:"task"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdat"`
}

type TodoList struct {
	Todos    []Todo `json:"todos"`
	Filename string
}

func NewTodoList(filename string) *TodoList {
	tl := &TodoList{
		Todos:    []Todo{},
		Filename: filename,
	}

	tl.Load()
	return tl
}

func (tl *TodoList) Load() error {
	file, err := os.ReadFile(tl.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(file, &tl.Todos)

}

func (tl *TodoList) Save() error {
	data, err := json.MarshalIndent(tl.Todos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tl.Filename, data, 0644)

}

func (tl *TodoList) Add(task string) {
	id := 1

	if len(tl.Todos) > 0 {
		id = tl.Todos[len(tl.Todos)-1].ID + 1

	}

	todo := Todo{
		ID:        id,
		Task:      task,
		Completed: false,
		CreatedAt: time.Now(),
	}

	tl.Todos = append(tl.Todos, todo)
	tl.Save()

	fmt.Printf("Added todo #%d : %v\n", id, todo)

}

func (tl *TodoList) List() {
	if len(tl.Todos) == 0 {
		fmt.Println("No todos found.")
		return
	}

	fmt.Println("Todo List:")
	fmt.Println(strings.Repeat("_", 60))

	for _, todo := range tl.Todos {
		status := "[ ]"
		if todo.Completed {
			status = "[✓]"
		}

		fmt.Printf("%s #%d: %s (created : %s)\n", status, todo.ID, todo.Task, todo.CreatedAt.Format("2006-01-02 15:04"))
	}

	fmt.Println(strings.Repeat("_", 60))

}

func (tl *TodoList) Complete(id int) {
	for i := range tl.Todos {
		if tl.Todos[i].ID == id {
			tl.Todos[i].Completed = true
			tl.Save()
			fmt.Printf("Completed todo #%d \n", id)
			return
		}
	}
}

func (tl *TodoList) Delete(id int) {
	for i := range tl.Todos {
		if tl.Todos[i].ID == id {
			tl.Todos = append(tl.Todos[:i], tl.Todos[i+1:]...)
			tl.Save()
			fmt.Printf("Deleted todo #%d \n", id)
			return
		}
	}
	fmt.Printf("Todo #%d not found", id)
}

func printUsage() {
	fmt.Println("\n📋 Todo App - Usage:")
	fmt.Println("  go run main.go add <task>      - Add a new todo")
	fmt.Println("  go run main.go list            - List all todos")
	fmt.Println("  go run main.go complete <id>   - Mark todo as completed")
	fmt.Println("  go run main.go delete <id>     - Delete a todo")
	fmt.Println()
}

func main() {

	todoList := NewTodoList("todos.json")

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "add":

		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a task description")
			return
		}
		task := strings.Join(os.Args[2:], " ")
		todoList.Add(task)

	case "list":
		todoList.List()

	case "complete":
		if len(os.Args) < 3 {
			fmt.Println("error : please provide a todo ID")
			return
		}
		id, err := strconv.Atoi(os.Args[2])

		if err != nil {
			fmt.Println("Error : Invalid Id")
			return
		}

		todoList.Complete(id)

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a todo ID")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Error: Invalid ID")
			return
		}
		todoList.Delete(id)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}

}
