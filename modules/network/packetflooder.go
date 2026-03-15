package network

import (
	"bytes"
	"sync/atomic"
	"time"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type PacketFlooder struct {
	module.Base
	running int32
	stats   struct {
		sent    int64
		dropped int64
	}
}

func NewPacketFlooder() *PacketFlooder {
	m := &PacketFlooder{
		Base: module.NewBase("PacketFlooder", "Флудит сервер пакетами анимации/движения", module.Network),
	}
	m.AddSetting(module.NewInt("pps", 100, 10, 2000))
	m.AddSetting(module.NewBool("animate", true))
	m.AddSetting(module.NewBool("move", false))
	m.AddSetting(module.NewBool("showstats", true))
	return m
}

func (m *PacketFlooder) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *PacketFlooder) Enable() {
	m.Base.Enable()
	atomic.StoreInt32(&m.running, 1)
	m.stats.sent = 0
	m.stats.dropped = 0
	go m.loop()
	if m.Bool("showstats") {
		go m.statsLoop()
	}
}

func (m *PacketFlooder) Disable() {
	m.Base.Disable()
	atomic.StoreInt32(&m.running, 0)
}

func (m *PacketFlooder) loop() {
	for atomic.LoadInt32(&m.running) == 1 {
		pps := m.Int("pps")
		if pps <= 0 {
			pps = 100
		}
		if !m.Proxy.IsFullyOnline() {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		interval := time.Duration(1000/pps) * time.Millisecond

		if m.Bool("animate") {
			var buf bytes.Buffer
			pk := &packet.AnimatePacket{
				Action: 1,
				Eid:    int32(m.Proxy.MyEID()),
			}
			buf.WriteByte(pk.ID())
			pk.Marshal(packet.NewWriter(&buf, 2))
			m.Proxy.SendRawToServer(buf.Bytes())
			atomic.AddInt64(&m.stats.sent, 1)
		}

		if m.Bool("move") {
			loc := m.Proxy.MyLocation()
			m.Proxy.SendMove(loc.X, loc.Y, loc.Z, loc.Yaw, loc.Pitch, true)
			atomic.AddInt64(&m.stats.sent, 1)
		}

		time.Sleep(interval)
	}
}

func (m *PacketFlooder) statsLoop() {
	for atomic.LoadInt32(&m.running) == 1 {
		time.Sleep(3 * time.Second)
		if !m.Proxy.IsClientOnline() {
			continue
		}
		sent := atomic.LoadInt64(&m.stats.sent)
		m.Proxy.NotifyActionBar("§d[Flooder] §fОтправлено: §e" + formatInt64(sent) + " §fпакетов")
	}
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + formatInt64(-n)
	}
	res := ""
	for n > 0 {
		res = string(rune('0'+n%10)) + res
		n /= 10
	}
	return res
}
