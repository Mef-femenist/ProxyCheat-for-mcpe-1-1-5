package module

import (
	"strings"

	"mefproxy/proxy"
)

type Module interface {
	Name() string
	Category() Category
	Description() string
	IsEnabled() bool
	Enable()
	Disable()
	Init(p *proxy.Proxy)
	Settings() []Setting
	GetSetting(name string) Setting
	SetSetting(name, value string) error
}

type Base struct {
	name        string
	description string
	category    Category
	enabled     bool
	Proxy       *proxy.Proxy
	settings    []Setting
	settingsMap map[string]Setting
}

func NewBase(name, description string, category Category) Base {
	return Base{
		name:        name,
		description: description,
		category:    category,
		settingsMap: make(map[string]Setting),
	}
}

func (b *Base) Name() string        { return b.name }
func (b *Base) Description() string { return b.description }
func (b *Base) Category() Category  { return b.category }
func (b *Base) IsEnabled() bool     { return b.enabled }
func (b *Base) SetEnabled(v bool)   { b.enabled = v }
func (b *Base) Init(p *proxy.Proxy) { b.Proxy = p }
func (b *Base) Enable()             { b.enabled = true }
func (b *Base) Disable()            { b.enabled = false }

func (b *Base) AddSetting(s Setting) {
	b.settings = append(b.settings, s)
	b.settingsMap[strings.ToLower(s.Name())] = s
}

func (b *Base) Settings() []Setting { return b.settings }

func (b *Base) GetSetting(name string) Setting {
	return b.settingsMap[strings.ToLower(name)]
}

func (b *Base) SetSetting(name, value string) error {
	s := b.GetSetting(name)
	if s == nil {
		return nil
	}
	return s.SetFromString(value)
}

func (b *Base) Bool(name string) bool {
	s := b.GetSetting(name)
	if s == nil {
		return false
	}
	if bs, ok := s.(*BoolSetting); ok {
		return bs.Value
	}
	return false
}

func (b *Base) Float(name string) float32 {
	s := b.GetSetting(name)
	if s == nil {
		return 0
	}
	switch v := s.(type) {
	case *FloatSetting:
		return v.Value
	case *SliderSetting:
		return v.Value
	}
	return 0
}

func (b *Base) Int(name string) int {
	s := b.GetSetting(name)
	if s == nil {
		return 0
	}
	if is, ok := s.(*IntSetting); ok {
		return is.Value
	}
	return 0
}

func (b *Base) Text(name string) string {
	s := b.GetSetting(name)
	if s == nil {
		return ""
	}
	if ts, ok := s.(*TextSetting); ok {
		return ts.Value
	}
	return ""
}
