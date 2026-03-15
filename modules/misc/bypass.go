package misc

import (
	"mefproxy/module"
	"mefproxy/proxy"
)

type Bypass struct {
	module.Base
}

func NewBypass() *Bypass {
	return &Bypass{Base: module.NewBase("Bypass", "Пропускает загрузку ресурспаков", module.Misc)}
}

func (m *Bypass) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Bypass) Enable() {
	m.Base.Enable()
	m.Proxy.SetConfig(func(c *proxy.Config) {
		c.BypassRPDownload = true
	})
}

func (m *Bypass) Disable() {
	m.Base.Disable()
	m.Proxy.SetConfig(func(c *proxy.Config) {
		c.BypassRPDownload = false
	})
}
