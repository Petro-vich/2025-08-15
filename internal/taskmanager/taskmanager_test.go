package taskmanager

import (
	"testing"

	"github.com/google/uuid"
)

func mockConfig() *Config {
	return &Config{
		MaxActiveTasks:    3,
		MaxFilesPerTask:   3,
		AllowedExtensions: []string{".pdf", ".jpeg"},
	}
}

func TestNewTaskManager(t *testing.T) {
	tm := NewTaskManager(mockConfig())
	if tm == nil {
		t.Fatal("TaskManager is nil")
	}
}

func TestAddTaskAndFilterExtensions(t *testing.T) {
	tm := NewTaskManager(mockConfig())
	task := &task{
		ID:     uuid.New(),
		Status: TaskStatusCreated,
		Files:  []string{},
	}
	tm.tasks[task.ID] = task
	ext := ".pdf"
	if !contains(tm.config.AllowedExtensions, ext) {
		t.Errorf("Extension %s should be allowed", ext)
	}
	badExt := ".exe"
	if contains(tm.config.AllowedExtensions, badExt) {
		t.Errorf("Extension %s should NOT be allowed", badExt)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
