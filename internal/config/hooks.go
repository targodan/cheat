package config

// Hook contains the config information for one hook.
type Hook struct {
	Name   string   `yaml:"name"`
	Events []string `yaml:"events"`
	Path   string   `yaml:"path"`
}
