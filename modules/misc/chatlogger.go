package misc

import (
	"fmt"
	"log"
	"os"
	"time"

	"mefproxy/module"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type ChatLogger struct {
	module.Base
	file *os.File
}

func NewChatLogger() *ChatLogger {
	m := &ChatLogger{Base: module.NewBase("ChatLogger", "Сохраняет весь чат в файл", module.Misc)}
	m.AddSetting(module.NewText("file", "chat.log"))
	return m
}

func (m *ChatLogger) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *ChatLogger) Enable() {
	m.Base.Enable()
	fname := m.Text("file")
	if fname == "" {
		fname = "chat.log"
	}
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("[ChatLogger] Не удалось открыть файл:", err)
		return
	}
	m.file = f
	m.Proxy.OnServerPacket(m.Name(), func(pk []byte) bool {
		if pk[0] != packet.IDTextPacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		tp, ok := decoded.(*packet.TextPacket)
		if !ok {
			return false
		}
		line := fmt.Sprintf("[%s] <%s> %s\n",
			time.Now().Format("15:04:05"),
			tp.SourceName,
			tp.Message,
		)
		if m.file != nil {
			_, _ = m.file.WriteString(line)
		}
		return false
	})
}

func (m *ChatLogger) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
	if m.file != nil {
		_ = m.file.Close()
		m.file = nil
	}
}
