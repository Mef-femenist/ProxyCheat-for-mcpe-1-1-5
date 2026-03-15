package combat

import (
	"bytes"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Criticals struct {
	module.Base
}

func NewCriticals() *Criticals {
	return &Criticals{
		Base: module.NewBase("Criticals", "Добавляет крит-прыжок перед атакой", module.Combat),
	}
}

func (m *Criticals) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Criticals) Enable() {
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
		loc := m.Proxy.MyLocation()
		sendMove := func(dy float32, onGround bool) {
			var buf bytes.Buffer
			mv := &packet.PlayerMovePacket{
				X: loc.X, Y: loc.Y + dy, Z: loc.Z,
				Yaw: loc.Yaw, BodyYaw: loc.Yaw, Pitch: loc.Pitch,
				Mode:     0,
				OnGround: onGround,
			}
			buf.WriteByte(mv.ID())
			mv.Marshal(packet.NewWriter(&buf, 2))
			m.Proxy.SendRawToServer(buf.Bytes())
		}
		sendMove(0.08, false)
		sendMove(0.0, false)
		return false
	})
}

func (m *Criticals) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
}
