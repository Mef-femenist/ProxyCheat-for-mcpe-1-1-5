package network

import (
	"sync"
	"time"

	"mefproxy/module"
	"mefproxy/proxy"
)

type PacketRateLimiter struct {
	module.Base
	mu       sync.Mutex
	counters map[byte]*rateEntry
}

type rateEntry struct {
	count    int
	window   time.Time
	dropped  int64
	total    int64
}

func NewPacketRateLimiter() *PacketRateLimiter {
	m := &PacketRateLimiter{
		Base:     module.NewBase("RateLimiter", "Ограничивает частоту пакетов клиента", module.Network),
		counters: make(map[byte]*rateEntry),
	}
	m.AddSetting(module.NewInt("maxpps", 40, 5, 200))
	m.AddSetting(module.NewInt("window", 1000, 100, 5000))
	m.AddSetting(module.NewBool("notify", false))
	return m
}

func (m *PacketRateLimiter) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *PacketRateLimiter) Enable() {
	m.Base.Enable()
	m.counters = make(map[byte]*rateEntry)
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if len(pk) == 0 {
			return false
		}
		id := pk[0]
		maxPPS := m.Int("maxpps")
		windowMs := m.Int("window")
		if maxPPS <= 0 {
			maxPPS = 40
		}
		if windowMs <= 0 {
			windowMs = 1000
		}
		window := time.Duration(windowMs) * time.Millisecond

		m.mu.Lock()
		entry, ok := m.counters[id]
		if !ok {
			entry = &rateEntry{window: time.Now()}
			m.counters[id] = entry
		}
		if time.Since(entry.window) > window {
			entry.count = 0
			entry.window = time.Now()
		}
		entry.total++
		if entry.count >= maxPPS {
			entry.dropped++
			m.mu.Unlock()
			if m.Bool("notify") {
				m.Proxy.NotifyActionBar("§c[RateLimit] §fdropped 0x" + byteHex(id))
			}
			return true
		}
		entry.count++
		m.mu.Unlock()
		return false
	})
}

func (m *PacketRateLimiter) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
	m.mu.Lock()
	m.counters = make(map[byte]*rateEntry)
	m.mu.Unlock()
}

func byteHex(b byte) string {
	const hex = "0123456789ABCDEF"
	return string([]byte{hex[b>>4], hex[b&0xf]})
}
