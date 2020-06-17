package hook

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cheat/cheat/internal/config"
	"github.com/cheat/cheat/internal/sheet"
)

// Manager manages hook creation and execution.
type Manager struct {
	conf *config.Config

	hooks map[Event][]*Hook
}

// NewManager create a new hook Manager.
//
// During the creation of the manager, all hooks are created as well.
// An error is returned if one of the hooks, defined by the given
// config, either have an invalid path or invalid events.
func NewManager(conf config.Config) (*Manager, error) {
	m := &Manager{
		conf: &conf,

		hooks: make(map[Event][]*Hook),
	}

	err := m.createHooksFromConfig()
	if err != nil {
		return nil, err
	}

	return m, nil
}

// RunOnStartHooks executes any registered OnStart hooks.
func (m *Manager) RunOnStartHooks() {
	m.runHooksOfType(OnStart)
}

// RunOnStopHooks executes any registered OnStart hooks.
func (m *Manager) RunOnStopHooks() {
	m.runHooksOfType(OnStop)
}

// RunOnSheetViewPreHooks executes any registered OnSheetViewPre hooks.
func (m *Manager) RunOnSheetViewPreHooks(sheet sheet.Sheet) {
	m.runHooksOfTypeWithSheet(OnSheetViewPre, &sheet)
}

// RunOnSheetViewPostHooks executes any registered OnSheetViewPost hooks.
func (m *Manager) RunOnSheetViewPostHooks(sheet sheet.Sheet) {
	m.runHooksOfTypeWithSheet(OnSheetViewPost, &sheet)
}

// RunOnSheetEditPreHooks executes any registered OnSheetEditPre hooks.
func (m *Manager) RunOnSheetEditPreHooks(sheet sheet.Sheet) {
	m.runHooksOfTypeWithSheet(OnSheetEditPre, &sheet)
}

// RunOnSheetEditPostHooks executes any registered OnSheetEditPost hooks.
func (m *Manager) RunOnSheetEditPostHooks(sheet sheet.Sheet) {
	m.runHooksOfTypeWithSheet(OnSheetEditPost, &sheet)
}

// RunOnSheetRemovePreHooks executes any registered OnSheetViewPre hooks.
func (m *Manager) RunOnSheetRemovePreHooks(sheet sheet.Sheet) {
	m.runHooksOfTypeWithSheet(OnSheetViewPre, &sheet)
}

// RunOnSheetRemovePostHooks executes any registered OnSheetViewPost hooks.
func (m *Manager) RunOnSheetRemovePostHooks(sheet sheet.Sheet) {
	m.runHooksOfTypeWithSheet(OnSheetViewPost, &sheet)
}

func (m *Manager) createHooksFromConfig() error {
	for _, h := range m.conf.Hooks {
		hook, err := New(h.Name, h.Path)
		if err != nil {
			return err
		}

		for _, t := range h.Events {
			tName, err := NameToEvent(t)
			if err != nil {
				return err
			}

			_, exists := m.hooks[tName]
			if !exists {
				m.hooks[tName] = make([]*Hook, 0)
			}

			m.hooks[tName] = append(m.hooks[tName], hook)
		}
	}
	return nil
}

func (m *Manager) runHooksOfType(t Event, args ...string) error {
	// prepend type
	args = append([]string{eventNames[t]}, args...)

	// The hook environment contains the config values
	env := m.buildHookEnv()

	for _, h := range m.hooks[t] {
		// It's debatable whether or not we want to continue with other hooks if one fails.
		// Assuming that the order of hooks might matter and they might depend on oneanother
		// it's safer to stop after the first error.
		err := h.Exec(args, env)
		if err != nil {
			m.handleExecError(t, h, err)
			return err
		}
	}

	return nil
}

func (m *Manager) runHooksOfTypeWithSheet(t Event, sheet *sheet.Sheet) error {
	var rdonly string
	if sheet.ReadOnly {
		rdonly = "true"
	} else {
		rdonly = "false"
	}

	// Reasoning for these call parameters (and order)
	// - Make all sheet information available without forcing the hook to parse the sheet
	// - Most relevant info first
	// 		1. Hook type
	// 		2. Path first (interesting for git stuff, as well as content of the sheet)
	//		3. Is it rw/rdonly second because of git stuff (only push on rw sheets)
	//		4. Syntax
	//		5. Tags (separated by ',' without spaces)

	return m.runHooksOfType(t, sheet.Path, rdonly, sheet.Title, sheet.Syntax, strings.Join(sheet.Tags, ","))
}

func (m *Manager) buildHookEnv() map[string]string {
	return config.ConfigToEnvironment(m.conf)
}

func (m *Manager) handleExecError(t Event, h *Hook, err error) {
	if err == nil {
		return
	}

	fatal := false

	exitErr, ok := err.(*exec.ExitError)
	if ok {
		fmt.Fprintf(os.Stderr, "Hook \"%s\" exited with code: %d\n", h.Name, exitErr.ExitCode())
		fmt.Fprintf(os.Stderr, "------ STDERR ------\n%s--------------------\n", string(exitErr.Stderr))
		// Maybe make this dependant on a specific exit code of the hook.
		// E.g. if the hook exists with 42, we consider this fatal and any
		// other non-zero exit code continues execution.
		fatal = true
	} else {
		fmt.Fprintf(os.Stderr, "Hook \"%s\" could not be executed: %v\n", h.Name, err)
		// This happens e.g. if the file exists but is not executable by the current user.
		// Not sure if this should be fatal or not.
		// Ideally handle the case of a non-executable file in the creation (New), in which
		// case this should always be treated as fatal.
		fatal = true
	}

	if fatal {
		os.Exit(2)
	}
}
