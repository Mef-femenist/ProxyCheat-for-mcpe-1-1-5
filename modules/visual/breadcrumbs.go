package visual

import (
	"time"

	"mefproxy/module"
	mgl32 "mefproxy/pkg/math"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Breadcrumbs struct {
	module.Base
	trail   []vec3.Vector3
	lastPos vec3.Vector3
}

func NewBreadcrumbs() *Breadcrumbs {
	m := &Breadcrumbs{
		Base: module.NewBase("Breadcrumbs", "Оставляет след частиц за игроком", module.Visual),
	}
	m.AddSetting(module.NewInt("maxlen", 100, 10, 500))
	return m
}

func (m *Breadcrumbs) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Breadcrumbs) Enable() {
	m.Base.Enable()
	m.trail = nil
	m.lastPos = m.Proxy.MyLocation()
	go m.loop()
}

func (m *Breadcrumbs) Disable() {
	m.Base.Disable()
	m.trail = nil
}

func (m *Breadcrumbs) loop() {
	for m.IsEnabled() {
		time.Sleep(200 * time.Millisecond)
		if !m.Proxy.IsClientOnline() {
			continue
		}
		cur := m.Proxy.MyLocation()
		maxLen := m.Int("maxlen")
		if vec3.Distance(cur, m.lastPos) > 0.5 {
			m.trail = append(m.trail, cur)
			if len(m.trail) > maxLen {
				m.trail = m.trail[1:]
			}
			m.lastPos = cur
		}
		for _, pos := range m.trail {
			m.Proxy.SendToClient(&packet.LevelEventPacket{
				Evid: 0x4000 | 2,
				Data: 0,
				Pos:  &mgl32.Vec3{pos.X, pos.Y, pos.Z},
			})
		}
	}
}
