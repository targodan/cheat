package hook

import (
	"errors"
	"os"
	"os/exec"
)

type Type int

const (
	OnStart Type = iota
	OnStop

	OnSheetView
	OnSheetEditOpen
	OnSheetEditClose
	OnSheetRemove
)

var TypeNames = map[Type]string{
	OnStart: "OnStart",
	OnStop:  "OnStop",

	OnSheetView:      "OnSheetView",
	OnSheetEditOpen:  "OnSheetEditOpen",
	OnSheetEditClose: "OnSheetEditClose",
	OnSheetRemove:    "OnSheetRemove",
}

type Hook struct {
	Path string
}

func New(path string) (*Hook, error) {
	h := &Hook{
		Path: path,
	}

	fInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fInfo.IsDir() {
		return nil, errors.New("given hook path is a directory, but must be a file")
	}

	// Checking if executable by current user is non-trivial, this will happen on running the hook.

	return h, nil
}

func (h *Hook) Exec(args []string, env map[string]string) error {
	cmd := exec.Command(h.Path, args...)

	environ := os.Environ()
	for key, value := range env {
		environ = append(environ, key+"="+value)
	}
	cmd.Env = environ

	return cmd.Run()
}
