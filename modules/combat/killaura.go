package combat

import (
	"time"

	"mefproxy/module"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type KillAura struct {
	module.Base
	lastHit time.Time
}

func NewKillAura() *KillAura {
	m := &KillAura{
		Base: module.NewBase("KillAura", "Авто-атака ближайших игроков", module.Combat),
	}
	m.AddSetting(module.NewSlider("range", 4.0, 1.0, 10.0, 0.5))
	m.AddSetting(module.NewInt("cps", 10, 1, 20))
	m.AddSetting(module.NewInt("targets", 1, 1, 5))
	m.AddSetting(module.NewBool("rotate", false))
	return m
}

func (m *KillAura) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *KillAura) Enable() {
	m.Base.Enable()
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDPlayerMovePacket {
			return false
		}
		cps := m.Int("cps")
		if cps <= 0 {
			cps = 1
		}
		if time.Since(m.lastHit) < time.Duration(1000/cps)*time.Millisecond {
			return false
		}
		if !m.Proxy.IsFullyOnline() {
			return false
		}
		maxTargets := m.Int("targets")
		targets := m.Proxy.NearestPlayers(maxTargets)
		if len(targets) == 0 {
			return false
		}
		attackRange := m.Float("range")
		hit := false
		for _, target := range targets {
			if vec3.Distance(m.Proxy.MyLocation(), target.Location) > attackRange {
				continue
			}
			m.Proxy.AttackEntity(target.EntityRuntimeID)
			hit = true
		}
		if hit {
			m.lastHit = time.Now()
		}
		return false
	})
}

func (m *KillAura) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
}
