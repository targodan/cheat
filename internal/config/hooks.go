package config

// Hook contains the config information for one hook.
type Hook struct {
	Events []string `yaml:"events"`
	Path   string   `yaml:"path"`
}
