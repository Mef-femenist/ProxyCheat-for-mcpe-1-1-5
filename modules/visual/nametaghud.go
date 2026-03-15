package visual

import (
	"fmt"
	"strings"
	"time"

	"mefproxy/module"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type NametagHUD struct {
	module.Base
}

func NewNametagHUD() *NametagHUD {
	m := &NametagHUD{
		Base: module.NewBase("NametagHUD", "ActionBar с именами и дистанцией ближайших", module.Visual),
	}
	m.AddSetting(module.NewInt("maxplayers", 5, 1, 20))
	return m
}

func (m *NametagHUD) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *NametagHUD) Enable() {
	m.Base.Enable()
	go m.loop()
}

func (m *NametagHUD) Disable() {
	m.Base.Disable()
	m.Proxy.SendToClient(&packet.SetTitlePacket{
		Type:  5,
		Title: "",
	})
}

func (m *NametagHUD) loop() {
	for m.IsEnabled() {
		time.Sleep(500 * time.Millisecond)
		if !m.Proxy.IsClientOnline() {
			continue
		}
		players := m.Proxy.Players()
		myLoc := m.Proxy.MyLocation()
		maxP := m.Int("maxplayers")

		type entry struct {
			nick string
			dist float32
		}
		var entries []entry
		for _, pl := range players {
			d := vec3.Distance(myLoc, pl.Location)
			entries = append(entries, entry{pl.Nick, d})
		}
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[j].dist < entries[i].dist {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
		if len(entries) > maxP {
			entries = entries[:maxP]
		}

		var parts []string
		for _, e := range entries {
			parts = append(parts, fmt.Sprintf("§c%s §7%.0fm", e.nick, e.dist))
		}

		text := strings.Join(parts, " §8| ")
		if text == "" {
			text = "§7нет игроков рядом"
		}

		m.Proxy.SendToClient(&packet.SetTitlePacket{
			Type:            4,
			Title:           text,
			FadeInDuration:  0,
			Duration:        25,
			FadeOutDuration: 5,
		})
	}
}
