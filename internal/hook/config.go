package hook

type Config struct {
	Types []Type `yaml:"events"`
	Path  string `yaml:"path"`
}
