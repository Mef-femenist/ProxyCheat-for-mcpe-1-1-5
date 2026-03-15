package movement

import (
	"mefproxy/module"
	"mefproxy/proxy"
)

type Speed struct {
	module.Base
}

func NewSpeed() *Speed {
	m := &Speed{
		Base: module.NewBase("Speed", "Устанавливает скорость передвижения", module.Movement),
	}
	m.AddSetting(module.NewSlider("value", 1.0, 0.1, 10.0, 0.1))
	return m
}

func (m *Speed) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Speed) Enable() {
	m.Base.Enable()
	m.Proxy.SetMySpeed(m.Float("value"))
}

func (m *Speed) Disable() {
	m.Base.Disable()
	m.Proxy.SetMySpeed(0.1)
}
