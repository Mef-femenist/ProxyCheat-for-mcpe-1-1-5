package misc

import (
	"log"
	"time"

	"mefproxy/module"
	"mefproxy/proxy"
)

type AutoReconnect struct {
	module.Base
}

func NewAutoReconnect() *AutoReconnect {
	m := &AutoReconnect{
		Base: module.NewBase("AutoReconnect", "Авто-переподключение при дисконнекте", module.Misc),
	}
	m.AddSetting(module.NewInt("delay", 5, 1, 60))
	return m
}

func (m *AutoReconnect) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *AutoReconnect) Enable() {
	m.Base.Enable()
	m.Proxy.OnServerDisconnect(m.Name(), func() {
		if !m.IsEnabled() {
			return
		}
		delay := m.Int("delay")
		log.Printf("[AutoReconnect] Переподключение через %ds...", delay)
		go func() {
			time.Sleep(time.Duration(delay) * time.Second)
			if !m.IsEnabled() {
				return
			}
			cfg := m.Proxy.GetConfig()
			m.Proxy.SwitchServer(cfg.Address)
		}()
	})
}

func (m *AutoReconnect) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
}
