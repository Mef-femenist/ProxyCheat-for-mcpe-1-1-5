package combat

import (
	"time"

	"mefproxy/module"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type TPAura struct {
	module.Base
}

func NewTPAura() *TPAura {
	m := &TPAura{
		Base: module.NewBase("TPAura", "Телепортируется к цели и атакует", module.Combat),
	}
	m.AddSetting(module.NewSlider("range", 8.0, 2.0, 20.0, 0.5))
	m.AddSetting(module.NewInt("cps", 6, 1, 20))
	m.AddSetting(module.NewBool("nofall", true))
	return m
}

func (m *TPAura) Init(p *proxy.Proxy) { m.Base.Init(p) }

var tpLastHit time.Time

func (m *TPAura) Enable() {
	m.Base.Enable()
	go m.loop()
}

func (m *TPAura) loop() {
	for m.IsEnabled() {
		cps := m.Int("cps")
		if cps <= 0 {
			cps = 6
		}
		time.Sleep(time.Duration(1000/cps) * time.Millisecond)
		if !m.Proxy.IsFullyOnline() {
			continue
		}
		target := m.Proxy.NearestPlayer()
		if target == nil {
			continue
		}
		dist := vec3.Distance(m.Proxy.MyLocation(), target.Location)
		if dist > m.Float("range") {
			continue
		}
		loc := target.Location
		nofall := m.Bool("nofall")
		m.Proxy.SendToServer(&packet.PlayerMovePacket{
			X:        loc.X,
			Y:        loc.Y,
			Z:        loc.Z,
			Yaw:      loc.Yaw,
			BodyYaw:  loc.Yaw,
			Pitch:    loc.Pitch,
			Mode:     0,
			OnGround: nofall,
		})
		m.Proxy.AttackEntity(target.EntityRuntimeID)
	}
}

func (m *TPAura) Disable() {
	m.Base.Disable()
}
