package combat

import (
	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Hitbox struct {
	module.Base
}

func NewHitbox() *Hitbox {
	m := &Hitbox{
		Base: module.NewBase("Hitbox", "Расширяет хитбоксы игроков", module.Combat),
	}
	m.AddSetting(module.NewSlider("width", 1.9, 0.6, 5.0, 0.1))
	m.AddSetting(module.NewSlider("height", 1.9, 1.0, 5.0, 0.1))
	return m
}

func (m *Hitbox) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Hitbox) Enable() {
	m.Base.Enable()
	m.Proxy.OnServerPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDEntityMovePacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		emp, ok := decoded.(*packet.EntityMovePacket)
		if !ok || emp.EntityRuntimeID == m.Proxy.MyEID() {
			return false
		}
		if !m.Proxy.State.IsPlayer(emp.EntityRuntimeID) {
			return false
		}
		m.Proxy.SetPlayerHitbox(emp.EntityRuntimeID, m.Float("width"), m.Float("height"))
		return false
	})
}

func (m *Hitbox) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
	for _, p := range m.Proxy.Players() {
		m.Proxy.SetPlayerHitbox(p.EntityRuntimeID, 0.6, 1.8)
	}
}
