package hook

import (
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
func (m *Manager) RunOnStartHooks() error {
	return m.runHooksOfType(OnStart)
}

// RunOnStopHooks executes any registered OnStart hooks.
func (m *Manager) RunOnStopHooks() error {
	return m.runHooksOfType(OnStop)
}

// RunOnSheetViewPreHooks executes any registered OnSheetViewPre hooks.
func (m *Manager) RunOnSheetViewPreHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetViewPre, &sheet)
}

// RunOnSheetViewPostHooks executes any registered OnSheetViewPost hooks.
func (m *Manager) RunOnSheetViewPostHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetViewPost, &sheet)
}

// RunOnSheetEditPreHooks executes any registered OnSheetEditPre hooks.
func (m *Manager) RunOnSheetEditPreHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetEditPre, &sheet)
}

// RunOnSheetEditPostHooks executes any registered OnSheetEditPost hooks.
func (m *Manager) RunOnSheetEditPostHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetEditPost, &sheet)
}

// RunOnSheetRemovePreHooks executes any registered OnSheetViewPre hooks.
func (m *Manager) RunOnSheetRemovePreHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetViewPre, &sheet)
}

// RunOnSheetRemovePostHooks executes any registered OnSheetViewPost hooks.
func (m *Manager) RunOnSheetRemovePostHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetViewPost, &sheet)
}

func (m *Manager) createHooksFromConfig() error {
	for _, h := range m.conf.Hooks {
		hook, err := New(h.Path)
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
