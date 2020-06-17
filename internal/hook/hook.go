package hook

import (
	"errors"
	"os"
	"os/exec"
)

// Event specifies the event type, i.e. the event upon which the
// hook will be executed.
type Event int

const (
	// OnStart hooks are executed right after the application is ready,
	// i.e. right after it has parsed the config and initialized the hooks.
	OnStart Event = iota
	// OnStop hooks are executed right before the application exits successfully.
	// This hook is not executed if the application terminates in a failure state.
	OnStop

	// OnSheetViewPre hooks are executed right before the application outputs
	// the contents of a sheet in the view command. Note that the hook is unable to
	// modify the output of the sheet at this point. The sheet-file has already
	// been read into memory.
	OnSheetViewPre
	// OnSheetViewPost hooks are executed after the contents of a sheet have been output
	// via the view command.
	OnSheetViewPost
	// OnSheetEditPre hooks are executed right before the editor is opened to modify a sheet.
	// Note that this happens after a readonly sheet has been copied.
	OnSheetEditPre
	// OnSheetEditPost hooks are executed after the editor has been closed successfully.
	// It will not be executed in case of a failure.
	OnSheetEditPost
	// OnSheetRemovePre hooks are executed before the sheet is removed.
	OnSheetRemovePre
	// OnSheetRemovePost hooks are executed after the sheet is removed.
	// They will still get all available information about the sheet via the command
	// arguments, but the file has already been removed.
	OnSheetRemovePost
)

// eventNames maps types to their names.
var eventNames = map[Event]string{
	OnStart: "OnStart",
	OnStop:  "OnStop",

	OnSheetViewPre:    "OnSheetViewPre",
	OnSheetViewPost:   "OnSheetViewPost",
	OnSheetEditPre:    "OnSheetEditPre",
	OnSheetEditPost:   "OnSheetEditPost",
	OnSheetRemovePre:  "OnSheetRemovePre",
	OnSheetRemovePost: "OnSheetRemovePost",
}

func EventToName(e Event) string {
	return eventNames[e]
}

// NameToEvent looks up the name of an event and returns the
// related event. If the name is not found an error is returned.
func NameToEvent(name string) (Event, error) {
	for t, n := range eventNames {
		if name == n {
			return t, nil
		}
	}
	return 0, errors.New("unknown event \"" + name + "\"")
}

// Hook represents one hook.
//
// The Hook -> Event association is stored inside the HookManager, not inside the Hook.
type Hook struct {
	Name string
	Path string
}

// New creates a new hook.
//
// This function also checks if the given path exists and if
// it is a file. It does *not* check if the file is executable
// by the current user. This will be handled by the Exec call.
func New(name, path string) (*Hook, error) {
	h := &Hook{
		Name: name,
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

// Exec executes a hook with the given arguments and environment.
//
// Note that the environment of the host process is copied before
// the given environment is added. This means that variables form
// the parent environment can be overwritten.
func (h *Hook) Exec(args []string, env map[string]string) error {
	cmd := exec.Command(h.Path, args...)

	environ := os.Environ()
	for key, value := range env {
		environ = append(environ, key+"="+value)
	}
	cmd.Env = environ

	// Output instead of Run, because this collects stderr in the
	// returned error, if there is an error
	_, err := cmd.Output()
	return err
}
