package visual

import (
	"time"

	"mefproxy/module"
	mgl32 "mefproxy/pkg/math"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Tracer struct {
	module.Base
}

func NewTracer() *Tracer {
	m := &Tracer{
		Base: module.NewBase("Tracer", "Рисует линии частиц к игрокам", module.Visual),
	}
	m.AddSetting(module.NewSlider("distance", 64.0, 10.0, 128.0, 1.0))
	m.AddSetting(module.NewSlider("step", 1.5, 0.5, 4.0, 0.5))
	return m
}

func (m *Tracer) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Tracer) Enable() {
	m.Base.Enable()
	go m.loop()
}

func (m *Tracer) Disable() {
	m.Base.Disable()
}

func (m *Tracer) loop() {
	for m.IsEnabled() {
		time.Sleep(600 * time.Millisecond)
		if !m.Proxy.IsClientOnline() {
			continue
		}
		from := m.Proxy.MyLocation()
		from.Y -= 1.5
		maxDist := m.Float("distance")
		for _, p := range m.Proxy.Players() {
			dist := vec3.Distance(from, p.Location)
			if dist < maxDist {
				m.drawLine(from, p.Location)
			}
		}
	}
}

func (m *Tracer) drawLine(from, to vec3.Vector3) {
	vector := vec3.Subtract(to, from)
	dist := vec3.Distance(from, to)
	step := m.Float("step")
	if step <= 0 {
		step = 1.5
	}
	for i := float32(1); i <= dist; i += step {
		v := vec3.Mult(vector, i)
		cur := vec3.Add(from, v)
		if !m.Proxy.IsClientOnline() {
			return
		}
		m.Proxy.SendToClient(&packet.LevelEventPacket{
			Evid: 0x4000 | 7,
			Data: 0,
			Pos:  &mgl32.Vec3{cur.X, cur.Y, cur.Z},
		})
		vector = vec3.Normalize(vector)
	}
}
