package taskmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/Petro-vich/2025-08-15/internal/archiver"
	"github.com/Petro-vich/2025-08-15/internal/downloader"
)

const (
	TaskStatusCreated    = "created"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusError      = "error"
)

type task struct {
	ID     uuid.UUID
	Status string
	Path   string
	Files  []string
}

type TaskManager struct {
	tasks  map[uuid.UUID]*task
	config *Config
	mutex  sync.Mutex
}

func NewTaskManager(config *Config) *TaskManager {
	return &TaskManager{
		tasks:  make(map[uuid.UUID]*task),
		config: config,
	}
}

func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func (tm *TaskManager) TasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if len(tm.tasks) >= tm.config.MaxActiveTasks {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "Server busy. No more than 3 tasks allowed"})
			return
		}

		newTask := &task{
			ID:     uuid.New(),
			Status: TaskStatusCreated,
			Files:  []string{},
		}
		newTask.Path = fmt.Sprintf("archives/%s", newTask.ID.String())
		tm.tasks[newTask.ID] = newTask

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": newTask.ID.String()})

	case http.MethodGet:
		if len(tm.tasks) < 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}

		var tasks []map[string]interface{}
		for _, ts := range tm.tasks {
			tasks = append(tasks, map[string]interface{}{
				"id":     ts.ID.String(),
				"status": ts.Status,
				"path":   ts.Path,
				"files":  ts.Files,
			})
		}
		json.NewEncoder(w).Encode(tasks)

	}

}

func (tm *TaskManager) TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid task ID"})
		return
	}

	task, exists := tm.tasks[taskID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Task ID not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     task.ID.String(),
		"status": task.Status,
		"path":   task.Path,
		"files":  task.Files,
	})

}

func (tm TaskManager) AddURLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ID format"})
		return
	}

	task, exists := tm.tasks[taskID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Task ID not found"})
		return
	}

	if len(task.Files) >= tm.config.MaxFilesPerTask {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "No more than 3 items allowed per task"})
		return
	}

	var data struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid data format"})
		return
	}

	extensions := filepath.Ext(data.URL)

	if !slices.Contains(tm.config.AllowedExtensions, extensions) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid file format"})
		log.Print(extensions)
		return
	}

	task.Files = append(task.Files, data.URL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File added successfully"})

	if len(task.Files) == 3 {
		err := tm.createdArchive(task)
		if err != nil {
			task.Status = TaskStatusError
			log.Printf("Failed to create archive: %v", err)
		}
	}
}

func (tm *TaskManager) DownloadArchiveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid task ID"})
		return
	}
	task, exists := tm.tasks[taskID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Task not found"})
		return
	}
	if task.Status != TaskStatusCompleted {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Archive not ready"})
		return
	}

	zipFilePath := fmt.Sprintf("archives/%s.zip", taskID.String())
	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Archive file not found"})
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", taskID.String()))
	http.ServeFile(w, r, zipFilePath)
}

func (tm TaskManager) createdArchive(task *task) error {
	task.Status = TaskStatusProcessing
	if err := archiver.EnsureDir("tmp"); err != nil {
		return fmt.Errorf("failed to create tmp dir: %w", err)
	}
	var downloadedFiles []string
	for i, ts := range task.Files {
		numberFile := strconv.Itoa(i)
		filename := filepath.Base(ts)
		outPath := filepath.Join("tmp", numberFile+"_"+filename)
		if err := downloader.DownloadFile(ts, outPath); err != nil {
			log.Printf("Failed to download file: %v", err)
			return fmt.Errorf("failed to download file: %w", err)
		}
		log.Printf("File %s downloaded successfully", filename)
		downloadedFiles = append(downloadedFiles, outPath)
	}
	return tm.createZip(task.ID, downloadedFiles)
}

func (tm TaskManager) createZip(taskID uuid.UUID, files []string) error {
	task := tm.tasks[taskID]
	zipFilePath := filepath.Join("archives", taskID.String()+".zip")
	if err := archiver.CreateZip(zipFilePath, files); err != nil {
		return err
	}
	task.Status = TaskStatusCompleted
	log.Printf("Archive %s created successfully", zipFilePath)
	return nil
}
