package taskmanager

type Config struct {
	BindAddr          string   `toml:"ADDR"`
	MaxFilesPerTask   int      `toml:"MAX_FILES_PER_TASKS"`
	MaxActiveTasks    int      `toml:"MAX_ACTIVE_TASKS"`
	AllowedExtensions []string `toml:"ALLOWED_EXTENSIONS"`
}

func NewConfig() *Config {
	return &Config{
		BindAddr:          ":8080",
		MaxFilesPerTask:   3,
		MaxActiveTasks:    3,
		AllowedExtensions: []string{".pdf", ".jpg", ".jpeg"},
	}
}
