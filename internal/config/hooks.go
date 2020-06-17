package config

// Hook
type Hook struct {
	Events []string `yaml:"events"`
	Path   string   `yaml:"path"`
}
