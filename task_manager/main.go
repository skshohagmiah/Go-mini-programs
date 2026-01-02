package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskManager struct {
	Tasks    []Task
	nextId   int
	filename string
}

func NewTaskManager(filename string) (*TaskManager, error) {
	tm := &TaskManager{
		Tasks:    make([]Task, 0),
		nextId:   1,
		filename: filename,
	}

	if err := tm.loadFromFile(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("Failed to load tasks : %w", err)
		}

		log.Println("No existing tasks file found, starting fresh")
	}

	return tm, nil
}

func (tm *TaskManager) loadFromFile() error {
	file, err := os.Open(tm.filename)
	if err != nil {
		return err
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Failed to read file : %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &tm.Tasks); err != nil {
		return fmt.Errorf("Failed to unmarshal tasks: %w", err)
	}

	for _, task := range tm.Tasks {
		if task.ID >= tm.nextId {
			tm.nextId = task.ID + 1
		}
	}

	log.Printf("Loaded %d tasks from file\n", len(tm.Tasks))
	return nil
}

func (tm *TaskManager) saveToFile() error {
	data, err := json.MarshalIndent(tm.Tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshall tasks : %w", err)
	}

	if err := os.WriteFile(tm.filename, data, 0644); err != nil {
		return fmt.Errorf("Failed to write file : %w", err)
	}

	return nil
}

func (tm *TaskManager) CreateTask(title, description string) (*Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("Title connot be empty")
	}

	task := Task{
		ID:          tm.nextId,
		Title:       title,
		Description: strings.TrimSpace(description),
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	tm.Tasks = append(tm.Tasks, task)

	tm.nextId++

	if err := tm.saveToFile(); err != nil {
		return nil, err
	}

	return &task, nil
}

func (tm *TaskManager) GetTasks() []Task {
	return tm.Tasks
}

func (tm *TaskManager) GetTask(id int) (*Task, error) {
	for i := range tm.Tasks {
		if tm.Tasks[i].ID == id {
			return &tm.Tasks[i], nil
		}
	}

	return nil, fmt.Errorf("task with id %d not found", id)
}

func (tm *TaskManager) UpdateTask(id int, completed bool) error {
	for i := range tm.Tasks {
		if tm.Tasks[i].ID == id {
			tm.Tasks[i].Completed = completed
			return tm.saveToFile()
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

func (tm *TaskManager) DeleteTask(id int) error {
	for i := range tm.Tasks {
		if tm.Tasks[i].ID == id {
			tm.Tasks = append(tm.Tasks[:i], tm.Tasks[i+1:]...)
			return tm.saveToFile()
		}
	}
	return fmt.Errorf("Task with id %d not found", id)
}

var taskManager *TaskManager

func handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		tasks := taskManager.GetTasks()
		json.NewEncoder(w).Encode(tasks)

	case http.MethodPost:
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		task, err := taskManager.CreateTask(req.Title, req.Description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	default:
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
	}
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "applicaton/json")

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(path)

	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, err := taskManager.GetTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(task)

	case http.MethodPut:
		// Update task
		var req struct {
			Completed bool `json:"completed"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := taskManager.UpdateTask(id, req.Completed); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Task updated successfully"}`)

	case http.MethodDelete:
		// Delete task
		if err := taskManager.DeleteTask(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Task deleted successfully"}`)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return

	}
	fmt.Fprintf(w, `
	Task Manager API
	================
	
	Endpoints:
	- GET    /tasks       - Get all tasks
	- POST   /tasks       - Create a new task
	- GET    /tasks/{id}  - Get a specific task
	- PUT    /tasks/{id}  - Update a task
	- DELETE /tasks/{id}  - Delete a task
	
	Example usage:
	curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Buy groceries","description":"Milk, eggs, bread"}'
	`)
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var err error
	taskManager, err = NewTaskManager("tasks.json")

	if err != nil {
		log.Fatalf("Failed to initiaze task manager: %v", err)
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/tasks", handleTasks)
	http.HandleFunc("/tasks/", handleTask)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on port %s... \n", port)
	log.Printf("Visit http://localhost:%s for API documentation\n", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Sever error : %w", err)
		}
	}()

	// Simulate graceful shutdown after some time (for demo purposes)
	// In a real app, you'd listen for OS signals
	log.Println("Server is running. Press Ctrl+C to stop.")

	// Keep server running
	select {}
}
