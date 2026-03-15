package movement

import (
	"bytes"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type NoFall struct {
	module.Base
}

func NewNoFall() *NoFall {
	return &NoFall{Base: module.NewBase("NoFall", "Обнуляет урон от падения", module.Movement)}
}

func (m *NoFall) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *NoFall) Enable() {
	m.Base.Enable()
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDPlayerMovePacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		mv, ok := decoded.(*packet.PlayerMovePacket)
		if !ok || mv.OnGround {
			return false
		}
		mv.OnGround = true
		var buf bytes.Buffer
		buf.WriteByte(mv.ID())
		mv.Marshal(packet.NewWriter(&buf, 2))
		m.Proxy.SendRawToServer(buf.Bytes())
		return true
	})
}

func (m *NoFall) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
}
