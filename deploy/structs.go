package deploy

import "fmt"

// Точка доступа
type AccessConfig struct {
	Path string `json:"path"`
	Port int    `json:"port"`
}

func (ac AccessConfig) GetPath() string {
	return fmt.Sprintf("%s:%d", ac.Path, ac.Port)
}
