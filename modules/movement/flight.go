package movement

import (
	"bytes"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Flight struct {
	module.Base
}

func NewFlight() *Flight {
	m := &Flight{
		Base: module.NewBase("Flight", "Полёт через патч флагов приключений", module.Movement),
	}
	m.AddSetting(module.NewSlider("speed", 1.0, 0.1, 5.0, 0.1))
	return m
}

func (m *Flight) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Flight) Enable() {
	m.Base.Enable()
	m.setFlags(true)
	spd := m.Float("speed")
	if spd != 1.0 {
		m.Proxy.SetMySpeed(spd)
	}
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDPlayerMovePacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		mv, ok := decoded.(*packet.PlayerMovePacket)
		if !ok {
			return false
		}
		mv.OnGround = false
		var buf bytes.Buffer
		buf.WriteByte(mv.ID())
		mv.Marshal(packet.NewWriter(&buf, 2))
		m.Proxy.SendRawToServer(buf.Bytes())
		return true
	})
}

func (m *Flight) Disable() {
	m.Base.Disable()
	m.setFlags(false)
	m.Proxy.SetMySpeed(0.1)
	m.Proxy.RemoveHooks(m.Name())
}

func (m *Flight) setFlags(fly bool) {
	v := uint32(0)
	if fly {
		v = 1
	}
	m.Proxy.SendToClient(&packet.AdventureSettingsPacket{
		AllowFlight: v,
		IsFlying:    v,
	})
}
