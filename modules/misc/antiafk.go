package misc

import (
	"time"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type AntiAFK struct {
	module.Base
	lastActivity time.Time
}

func NewAntiAFK() *AntiAFK {
	m := &AntiAFK{
		Base:         module.NewBase("AntiAFK", "Шевелит ротацией чтобы не кикнуло", module.Misc),
		lastActivity: time.Now(),
	}
	m.AddSetting(module.NewInt("interval", 30, 5, 300))
	return m
}

func (m *AntiAFK) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *AntiAFK) Enable() {
	m.Base.Enable()
	m.lastActivity = time.Now()
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if pk[0] == packet.IDPlayerMovePacket {
			m.lastActivity = time.Now()
		}
		return false
	})
	go m.loop()
}

func (m *AntiAFK) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
}

func (m *AntiAFK) loop() {
	for m.IsEnabled() {
		time.Sleep(1 * time.Second)
		if !m.Proxy.IsFullyOnline() {
			continue
		}
		interval := m.Int("interval")
		if interval <= 0 {
			interval = 30
		}
		if time.Since(m.lastActivity) < time.Duration(interval)*time.Second {
			continue
		}
		loc := m.Proxy.MyLocation()
		yaw := loc.Yaw + 0.5
		m.Proxy.SendToServer(&packet.PlayerMovePacket{
			X: loc.X, Y: loc.Y, Z: loc.Z,
			Yaw: yaw, BodyYaw: yaw, Pitch: loc.Pitch,
			Mode:     0,
			OnGround: true,
		})
		m.lastActivity = time.Now()
	}
}
