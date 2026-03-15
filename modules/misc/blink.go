package misc

import (
	"bytes"
	"sync"
	"time"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type Blink struct {
	module.Base
	mu      sync.Mutex
	buffer  [][]byte
	ticking bool
}

func NewBlink() *Blink {
	m := &Blink{
		Base: module.NewBase("Blink", "Буферизует движение, потом отправляет разом", module.Misc),
	}
	m.AddSetting(module.NewInt("duration", 3, 1, 30))
	return m
}

func (m *Blink) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *Blink) Enable() {
	m.Base.Enable()
	m.buffer = nil
	m.Proxy.OnClientPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDPlayerMovePacket {
			return false
		}
		m.mu.Lock()
		cp := make([]byte, len(pk))
		copy(cp, pk)
		m.buffer = append(m.buffer, cp)
		m.mu.Unlock()
		return true
	})
	go m.autoFlush()
}

func (m *Blink) autoFlush() {
	dur := m.Int("duration")
	if dur <= 0 {
		dur = 3
	}
	time.Sleep(time.Duration(dur) * time.Second)
	if m.IsEnabled() {
		m.flush()
	}
}

func (m *Blink) flush() {
	m.mu.Lock()
	buf := m.buffer
	m.buffer = nil
	m.mu.Unlock()
	for _, pk := range buf {
		if !m.Proxy.IsServerOnline() {
			break
		}
		var b bytes.Buffer
		b.Write(pk)
		m.Proxy.SendRawToServer(b.Bytes())
	}
	m.Proxy.Notify("§d[Blink] §fБуфер сброшен: §e" + itoa(len(buf)) + " §fпакетов")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	res := ""
	for n > 0 {
		res = string(rune('0'+n%10)) + res
		n /= 10
	}
	return res
}

func (m *Blink) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
	m.flush()
}
