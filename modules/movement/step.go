package movement

import (
	"mefproxy/module"
	"mefproxy/proxy"
)

type Step struct {
	module.Base
}

func NewStep() *Step {
	m := &Step{
		Base: module.NewBase("Step", "Увеличивает высоту шага", module.Movement),
	}
	m.AddSetting(module.NewSlider("height", 1.5, 0.5, 3.0, 0.1))
	return m
}

func (m *Step) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Step) Enable() {
	m.Base.Enable()
	m.Proxy.SetEntityData(m.Proxy.MyEID(), map[uint32]interface{}{
		39: m.Float("height"),
	})
}

func (m *Step) Disable() {
	m.Base.Disable()
	m.Proxy.SetEntityData(m.Proxy.MyEID(), map[uint32]interface{}{
		39: float32(0.5),
	})
}
