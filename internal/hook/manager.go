package hook

import (
	"fmt"
	"strings"

	"github.com/cheat/cheat/internal/config"
	"github.com/cheat/cheat/internal/sheet"
)

type Manager struct {
	conf *config.Config

	hooks map[Type][]*Hook
}

func NewManager(conf config.Config) (*Manager, error) {
	m := &Manager{
		conf: &conf,

		hooks: make(map[Type][]*Hook),
	}

	err := m.createHooksFromConfig()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) RunOnStartHooks() error {
	return m.runHooksOfType(OnStart)
}

func (m *Manager) RunOnStopHooks() error {
	return m.runHooksOfType(OnStop)
}

func (m *Manager) RunOnSheetViewPreHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetViewPre, &sheet)
}

func (m *Manager) RunOnSheetViewPostHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetViewPost, &sheet)
}

func (m *Manager) RunOnSheetEditPreHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetEditPre, &sheet)
}

func (m *Manager) RunOnSheetEditPostHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetEditPost, &sheet)
}

func (m *Manager) RunOnSheetRemovePreHooks(sheet sheet.Sheet) error {
	return m.runHooksOfTypeWithSheet(OnSheetViewPre, &sheet)
}
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
			tName, err := FindTypeFromName(t)
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

func (m *Manager) runHooksOfType(t Type, args ...string) error {
	// prepend type
	args = append([]string{TypeNames[t]}, args...)

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

func (m *Manager) runHooksOfTypeWithSheet(t Type, sheet *sheet.Sheet) error {
	var rdonly string
	if sheet.ReadOnly {
		rdonly = "rdonly"
	} else {
		rdonly = "rw"
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
	// We might want to move this into the internal/config package
	// so it is not overlooked when the config is changed.

	env := map[string]string{
		"CHEAT_CONF_EDITOR":           m.conf.Editor,
		"CHEAT_CONF_FORMATTER":        m.conf.Formatter,
		"CHEAT_CONF_STYLE":            m.conf.Style,
		"CHEAT_CONF_CHEATPATHS_COUNT": fmt.Sprint(len(m.conf.Cheatpaths)),
		"CHEAT_CONF_HOOKS_COUNT":      fmt.Sprint(len(m.conf.Hooks)),
	}

	for i, cp := range m.conf.Cheatpaths {
		env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_PATH", i)] = cp.Path
		env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_NAME", i)] = cp.Name
		if cp.ReadOnly {
			env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_READONLY", i)] = "true"
		} else {
			env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_READONLY", i)] = "false"
		}
		env[fmt.Sprintf("CHEAT_CONF_CHEATPATHS_%d_PATH", i)] = strings.Join(cp.Tags, ",")
	}

	for i, h := range m.conf.Hooks {
		env[fmt.Sprintf("CHEAT_CONF_HOOKS_%d_PATH", i)] = h.Path
		env[fmt.Sprintf("CHEAT_CONF_HOOKS_%d_TYPES", i)] = strings.Join(h.Events, ",")
	}

	return env
}
