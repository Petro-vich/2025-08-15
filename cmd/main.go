package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type task struct {
	ID     uuid.UUID
	Status string
	Path   string
	Files  []string
}

type TaskManager struct {
	tasks map[uuid.UUID]*task
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[uuid.UUID]*task),
	}
}

func (tm *TaskManager) tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if len(tm.tasks) >= 3 {
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "Сервер занят. Не более 3 задач"})
			return
		}

		//Вывести в отельный пакет
		newTask := &task{
			ID:     uuid.New(),
			Status: "created",
			Files:  []string{},
		}
		newTask.Path = fmt.Sprintf("archives/%s", newTask.ID.String())
		tm.tasks[newTask.ID] = newTask

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": newTask.ID.String()})

	case http.MethodGet:
		if len(tm.tasks) < 1 {
			fmt.Fprint(w, "Нет активных задач\n")
			return
		}

		var tasks []map[string]interface{}
		for _, ts := range tm.tasks {
			tasks = append(tasks, map[string]interface{}{
				"id":     ts.ID,
				"status": ts.Status,
				"path":   ts.Path,
				"files":  ts.Files,
			})
		}
		json.NewEncoder(w).Encode(tasks)

	}

}

func (tm *TaskManager) taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный ID задачи"})
		return
	}

	task, exists := tm.tasks[taskID]
	if !exists {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не найден"})
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     task.ID,
		"status": task.Status,
		"path":   task.Path,
		"files":  task.Files,
	})

}

func (tm TaskManager) addURLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])

	if err != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат ID"})
		return
	}

	task, exists := tm.tasks[taskID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не найден"})
		return
	}

	//Получить сразу несколько ссылок
	if len(task.Files)+1 >= 3 {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Разрешено не более 3 элементов в задачи"})
	}

	var data struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат данных"})
		return
	}

	extensions := filepath.Ext(data.URL)

	//формат через env
	if extensions != ".pdf" && extensions != ".jpg" {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Недопустимый формат"})
		log.Print(extensions)
		return
	}

	task.Files = append(task.Files, data.URL)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message"}: "Файл успешно добавлен")
	

	if len(tm.tasks[taskID].Files) == 3 {
		tm.createdArchive(task)
	}
}

func (tm TaskManager) createdArchive(task *task) {
	task.Status = "processing"
	for i, ts := range task.Files {
		resp, err := http.Get(ts)
		if err != nil {
			log.Printf("Ошибка получения файла: %v", err)
			continue
		}
		numberFile := strconv.Itoa(i)
		filename := filepath.Base(ts)
		out, err := os.Create(numberFile + "_" + filename)
		if err != nil {
			log.Printf("Ошибка создания файла: %v", err)
			continue
		}
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Printf("Ошибка копирования: %v", err)
		}
		log.Printf("Файл: %s успешно скачан\n", filename)
	}
	tm.createZip(task.ID)
}

func (tm TaskManager) createZip(taskID uuid.UUID) {
	task := tm.tasks[taskID]

	zipFileName := taskID.String() + ".zip"
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		log.Printf("Ошибка создания zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for i, fileURL := range task.Files {
		filename := fmt.Sprintf("%d_%s", i, filepath.Base(fileURL))
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Ошибка открытия файла %s: %v", filename, err)
			continue
		}
		w, err := zipWriter.Create(filename)
		if err != nil {
			log.Printf("Ошибка добавления файла в архив: %v", err)
			file.Close()
			continue
		}
		_, err = io.Copy(w, file)
		if err != nil {
			log.Printf("Ошибка копирования файла в архив: %v", err)
		}
	}
	log.Printf("Архив %s успешно создан", zipFileName)

}

func (tm TaskManager) test(w http.ResponseWriter, r *http.Request) {
	URL := "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3a/Cat03.jpg/960px-Cat03.jpg"
	resp, err := http.Get(URL)
	_ = err
	out, err := os.Create("tmp/Hello")
	_ = err
	_ = resp
	_ = out
}

func main() {

	router := mux.NewRouter()
	taskManager := NewTaskManager()

	router.HandleFunc("/tasks", taskManager.tasksHandler).Methods("POST", "GET")
	router.HandleFunc("/tasks/{id}", taskManager.taskByIDHandler).Methods("GET")
	router.HandleFunc("/tasks/{id}/files", taskManager.addURLHandler).Methods("POST")
	router.HandleFunc("/test", taskManager.test).Methods("GET")

	fmt.Print("Сервер запущен на порте: 8080\n")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Printf("Ошибка запуска сервера: %v", err)
	}
}
