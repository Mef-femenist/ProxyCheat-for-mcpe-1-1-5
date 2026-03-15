package visual

import (
	"fmt"
	"time"

	"mefproxy/module"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type ESP struct {
	module.Base
}

func NewESP() *ESP {
	m := &ESP{
		Base: module.NewBase("ESP", "Показывает ников и дистанцию над игроками", module.Visual),
	}
	m.AddSetting(module.NewSlider("distance", 32.0, 5.0, 128.0, 1.0))
	m.AddSetting(module.NewBool("showdist", true))
	m.AddSetting(module.NewBool("showhp", false))
	return m
}

func (m *ESP) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *ESP) Enable() {
	m.Base.Enable()
	go m.loop()
}

func (m *ESP) Disable() {
	m.Base.Disable()
}

func (m *ESP) loop() {
	for m.IsEnabled() {
		time.Sleep(800 * time.Millisecond)
		if !m.Proxy.IsClientOnline() {
			continue
		}
		myLoc := m.Proxy.MyLocation()
		maxDist := m.Float("distance")
		showDist := m.Bool("showdist")
		for _, pl := range m.Proxy.Players() {
			dist := vec3.Distance(myLoc, pl.Location)
			if dist > maxDist {
				continue
			}
			label := "§e" + pl.Nick
			if showDist {
				label = fmt.Sprintf("§e%s §7%.0fm", pl.Nick, dist)
			}
			m.Proxy.SendToClient(&packet.SetTitlePacket{
				Type:            4,
				Title:           label,
				FadeInDuration:  0,
				Duration:        40,
				FadeOutDuration: 5,
			})
		}
	}
}
