package combat

import (
	"bytes"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Velocity struct {
	module.Base
}

func NewVelocity() *Velocity {
	m := &Velocity{
		Base: module.NewBase("Velocity", "Снижает или обнуляет отбрасывание", module.Combat),
	}
	m.AddSetting(module.NewSlider("multiplier", 0.0, 0.0, 1.0, 0.05))
	return m
}

func (m *Velocity) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Velocity) Enable() {
	m.Base.Enable()
	m.Proxy.OnServerPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDSetEntityMotionPacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		emp, ok := decoded.(*packet.SetEntityMotionPacket)
		if !ok || uint32(emp.Eid) != m.Proxy.MyEID() {
			return false
		}
		mul := m.Float("multiplier")
		emp.MotionX *= mul
		emp.MotionY *= mul
		emp.MotionZ *= mul
		var buf bytes.Buffer
		buf.WriteByte(emp.ID())
		emp.Marshal(packet.NewWriter(&buf, 2))
		m.Proxy.SendRawToClient(buf.Bytes())
		return true
	})
}

func (m *Velocity) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
}
