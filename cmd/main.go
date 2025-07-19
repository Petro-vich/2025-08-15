package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/Petro-vich/2025-08-15/internal/taskmanager"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	config := taskmanager.NewConfig()
	_, err := toml.DecodeFile("configs/config.toml", config)
	if err != nil {
		log.Fatal(err)
	}

	taskManager := taskmanager.NewTaskManager(config)

	router.HandleFunc("/tasks", taskManager.TasksHandler).Methods("POST", "GET")
	router.HandleFunc("/tasks/{id}", taskManager.TaskByIDHandler).Methods("GET")
	router.HandleFunc("/tasks/{id}/files", taskManager.AddURLHandler).Methods("POST")
	router.HandleFunc("/tasks/{id}/archive", taskManager.DownloadArchiveHandler).Methods("GET")
	fmt.Println("Сервер запущен на адресе:", config.BindAddr)
	err = http.ListenAndServe(config.BindAddr, router)
	if err != nil {
		log.Printf("Ошибка запуска сервера: %v", err)
	}
}
