package movement

import (
	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type AutoSprint struct {
	module.Base
}

func NewAutoSprint() *AutoSprint {
	return &AutoSprint{Base: module.NewBase("AutoSprint", "Авто-спринт при каждом движении", module.Movement)}
}

func (m *AutoSprint) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *AutoSprint) Enable() {
	m.Base.Enable()
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDPlayerMovePacket {
			return false
		}
		if !m.Proxy.IsServerOnline() {
			return false
		}
		m.Proxy.SendSprint(true)
		return false
	})
}

func (m *AutoSprint) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
	m.Proxy.SendSprint(false)
}
