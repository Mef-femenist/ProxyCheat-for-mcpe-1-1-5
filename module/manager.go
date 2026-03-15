package module

import (
	"fmt"
	"strings"
	"sync"

	"mefproxy/proxy"
)

type Manager struct {
	mu      sync.RWMutex
	modules map[string]Module
	proxy   *proxy.Proxy
}

func NewManager(p *proxy.Proxy) *Manager {
	return &Manager{
		modules: make(map[string]Module),
		proxy:   p,
	}
}

func (m *Manager) Register(mod Module) {
	mod.Init(m.proxy)
	m.mu.Lock()
	m.modules[strings.ToLower(mod.Name())] = mod
	m.mu.Unlock()
}

func (m *Manager) Get(name string) Module {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.modules[strings.ToLower(name)]
}

func (m *Manager) Toggle(name string) bool {
	mod := m.Get(name)
	if mod == nil {
		return false
	}
	if mod.IsEnabled() {
		mod.Disable()
	} else {
		mod.Enable()
	}
	return true
}

func (m *Manager) All() []Module {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Module, 0, len(m.modules))
	for _, mod := range m.modules {
		out = append(out, mod)
	}
	return out
}

func (m *Manager) ByCategory(cat Category) []Module {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Module
	for _, mod := range m.modules {
		if mod.Category() == cat {
			out = append(out, mod)
		}
	}
	return out
}

func (m *Manager) DisableAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, mod := range m.modules {
		if mod.IsEnabled() {
			mod.Disable()
		}
	}
}

func (m *Manager) SetSetting(modName, settingName, value string) error {
	mod := m.Get(modName)
	if mod == nil {
		return fmt.Errorf("модуль не найден: %s", modName)
	}
	s := mod.GetSetting(settingName)
	if s == nil {
		return fmt.Errorf("настройка не найдена: %s", settingName)
	}
	return s.SetFromString(value)
}

func (m *Manager) StatusLine() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var active []string
	for _, mod := range m.modules {
		if mod.IsEnabled() {
			active = append(active, mod.Category().Color()+mod.Name())
		}
	}
	if len(active) == 0 {
		return "§7нет активных модулей"
	}
	return strings.Join(active, " §7| ")
}

func (m *Manager) FormatList() []string {
	cats := []Category{Combat, Movement, Exploit, Visual, Misc, Network}
	var lines []string
	for _, cat := range cats {
		mods := m.ByCategory(cat)
		if len(mods) == 0 {
			continue
		}
		var names []string
		for _, mod := range mods {
			state := "§c✗"
			if mod.IsEnabled() {
				state = "§a✓"
			}
			names = append(names, fmt.Sprintf("%s %s%s", state, cat.Color(), mod.Name()))
		}
		lines = append(lines, fmt.Sprintf("%s§l[%s]§r %s", cat.Color(), cat, strings.Join(names, " §7| ")))
	}
	return lines
}
