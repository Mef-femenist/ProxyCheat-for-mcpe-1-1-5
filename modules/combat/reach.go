package combat

import (
	"mefproxy/module"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Reach struct {
	module.Base
}

func NewReach() *Reach {
	m := &Reach{
		Base: module.NewBase("Reach", "Авто-атака на увеличенной дистанции", module.Combat),
	}
	m.AddSetting(module.NewSlider("distance", 6.0, 3.0, 12.0, 0.5))
	return m
}

func (m *Reach) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Reach) Enable() {
	m.Base.Enable()
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDInteractPacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		ip, ok := decoded.(*packet.InteractPacket)
		if !ok || ip.Action != 1 {
			return false
		}
		target := m.Proxy.NearestPlayer()
		if target == nil {
			return false
		}
		dist := vec3.Distance(m.Proxy.MyLocation(), target.Location)
		if dist > m.Float("distance") {
			m.Proxy.SendToServer(&packet.InteractPacket{
				Action: 1,
				Target: target.EntityRuntimeID,
			})
		}
		return false
	})
}

func (m *Reach) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
}
