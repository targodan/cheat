package config

type Hooks struct {
	Events []string `yaml:"events"`
	Path   string   `yaml:"path"`
}
