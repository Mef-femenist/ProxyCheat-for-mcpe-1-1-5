package misc

import (
	"fmt"
	"log"
	"os"
	"time"

	"mefproxy/module"
	"mefproxy/proxy"
)

type PacketLogger struct {
	module.Base
	file *os.File
}

func NewPacketLogger() *PacketLogger {
	m := &PacketLogger{
		Base: module.NewBase("PacketLogger", "Логирует все пакеты в файл", module.Misc),
	}
	m.AddSetting(module.NewBool("client", true))
	m.AddSetting(module.NewBool("server", true))
	m.AddSetting(module.NewText("file", "packets.log"))
	return m
}

func (m *PacketLogger) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *PacketLogger) Enable() {
	m.Base.Enable()
	fname := m.Text("file")
	if fname == "" {
		fname = "packets.log"
	}
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("[PacketLogger] Не удалось открыть файл:", err)
		return
	}
	m.file = f
	if m.Bool("client") {
		m.Proxy.OnClientPacket(m.Name()+"_c", func(pk []byte) bool {
			m.write("C→S", pk)
			return false
		})
	}
	if m.Bool("server") {
		m.Proxy.OnServerPacket(m.Name()+"_s", func(pk []byte) bool {
			m.write("S→C", pk)
			return false
		})
	}
}

func (m *PacketLogger) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name() + "_c")
	m.Proxy.RemoveHooks(m.Name() + "_s")
	if m.file != nil {
		_ = m.file.Close()
		m.file = nil
	}
}

func (m *PacketLogger) write(dir string, pk []byte) {
	if m.file == nil || len(pk) == 0 {
		return
	}
	line := fmt.Sprintf("[%s] %s id=0x%02X len=%d\n",
		time.Now().Format("15:04:05.000"),
		dir,
		pk[0],
		len(pk),
	)
	_, _ = m.file.WriteString(line)
}
