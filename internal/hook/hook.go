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

	OnSheetViewPre
	OnSheetViewPost
	OnSheetEditPre
	OnSheetEditPost
	OnSheetRemovePre
	OnSheetRemovePost
)

var TypeNames = map[Type]string{
	OnStart: "OnStart",
	OnStop:  "OnStop",

	OnSheetViewPre:    "OnSheetViewPre",
	OnSheetViewPost:   "OnSheetViewPost",
	OnSheetEditPre:    "OnSheetEditPre",
	OnSheetEditPost:   "OnSheetEditPost",
	OnSheetRemovePre:  "OnSheetRemovePre",
	OnSheetRemovePost: "OnSheetRemovePost",
}

func FindTypeFromName(name string) (Type, error) {
	for t, n := range TypeNames {
		if name == n {
			return t, nil
		}
	}
	return 0, errors.New("unknown type \"" + name + "\"")
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
