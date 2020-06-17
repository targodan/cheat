package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	cp "github.com/cheat/cheat/internal/cheatpath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// Config encapsulates configuration parameters
type Config struct {
	Colorize   bool           `yaml:"colorize"`
	Editor     string         `yaml:"editor"`
	Cheatpaths []cp.Cheatpath `yaml:"cheatpaths"`
	Style      string         `yaml:"style"`
	Formatter  string         `yaml:"formatter"`
	Hooks      []*Hook        `yaml:"hooks"`
}

// New returns a new Config struct
func New(opts map[string]interface{}, confPath string, resolve bool) (Config, error) {

	// read the config file
	buf, err := ioutil.ReadFile(confPath)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config file: %v", err)
	}

	// initialize a config object
	conf := Config{}

	// unmarshal the yaml
	err = yaml.UnmarshalStrict(buf, &conf)
	if err != nil {
		return Config{}, fmt.Errorf("could not unmarshal yaml: %v", err)
	}

	// if a .cheat directory exists locally, append it to the cheatpaths
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get cwd: %v", err)
	}

	local := filepath.Join(cwd, ".cheat")
	if _, err := os.Stat(local); err == nil {
		path := cp.Cheatpath{
			Name:     "cwd",
			Path:     local,
			ReadOnly: false,
			Tags:     []string{},
		}

		conf.Cheatpaths = append(conf.Cheatpaths, path)
	}

	// process cheatpaths
	for i, cheatpath := range conf.Cheatpaths {

		// expand ~ in config paths
		expanded, err := homedir.Expand(cheatpath.Path)
		if err != nil {
			return Config{}, fmt.Errorf("failed to expand ~: %v", err)
		}

		// follow symlinks
		//
		// NB: `resolve` is an ugly kludge that exists for the sake of unit-tests.
		// It's necessary because `EvalSymlinks` will error if the symlink points
		// to a non-existent location on the filesystem. When unit-testing,
		// however, we don't want to have dependencies on the filesystem. As such,
		// `resolve` is a switch that allows us to turn off symlink resolution when
		// running the config tests.
		if resolve {
			evaled, err := filepath.EvalSymlinks(expanded)
			if err != nil {
				return Config{}, fmt.Errorf(
					"failed to resolve symlink: %s: %v",
					expanded,
					err,
				)
			}

			expanded = evaled
		}

		conf.Cheatpaths[i].Path = expanded
	}

	// if an editor was not provided in the configs, look to envvars
	if conf.Editor == "" {
		if os.Getenv("VISUAL") != "" {
			conf.Editor = os.Getenv("VISUAL")
		} else if os.Getenv("EDITOR") != "" {
			conf.Editor = os.Getenv("EDITOR")
		} else {
			return Config{}, fmt.Errorf("no editor set")
		}
	}

	// if a chroma style was not provided, set a default
	if conf.Style == "" {
		conf.Style = "bw"
	}

	// if a chroma formatter was not provided, set a default
	if conf.Formatter == "" {
		conf.Formatter = "terminal16m"
	}

	return conf, nil
}

func ConfigToEnvironment(conf *Config) map[string]string {
	env := map[string]string{
		"CHEAT_CONF_EDITOR":           conf.Editor,
		"CHEAT_CONF_FORMATTER":        conf.Formatter,
		"CHEAT_CONF_STYLE":            conf.Style,
		"CHEAT_CONF_CHEATPATHS_COUNT": fmt.Sprint(len(conf.Cheatpaths)),
		"CHEAT_CONF_HOOKS_COUNT":      fmt.Sprint(len(conf.Hooks)),
	}

	for i, cp := range conf.Cheatpaths {
		env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_PATH", i)] = cp.Path
		env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_NAME", i)] = cp.Name
		if cp.ReadOnly {
			env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_READONLY", i)] = "true"
		} else {
			env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_READONLY", i)] = "false"
		}
		env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_PATH", i)] = strings.Join(cp.Tags, ",")
	}

	for i, h := range conf.Hooks {
		env[fmt.Sprintf("CHEAT_CONF_HOOKS_%d_PATH", i)] = h.Path
		env[fmt.Sprintf("CHEAT_CONF_HOOKS_%d_TYPES", i)] = strings.Join(h.Events, ",")
	}

	return env
}
