package network

import (
	"time"

	"mefproxy/module"
	"mefproxy/proxy"
)

type ConnectionSpoof struct {
	module.Base
}

func NewConnectionSpoof() *ConnectionSpoof {
	m := &ConnectionSpoof{
		Base: module.NewBase("ConnSpoof", "Подделывает параметры подключения", module.Network),
	}
	m.AddSetting(module.NewInt("deviceos", 7, 1, 14))
	m.AddSetting(module.NewText("model", "Samsung SM-G998B"))
	m.AddSetting(module.NewInt("inputmode", 2, 1, 4))
	m.AddSetting(module.NewBool("autoapply", true))
	return m
}

func (m *ConnectionSpoof) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *ConnectionSpoof) Enable() {
	m.Base.Enable()
	if m.Bool("autoapply") {
		m.apply()
	}
	m.Proxy.OnConnect(m.Name(), func() {
		if m.IsEnabled() {
			go func() {
				time.Sleep(300 * time.Millisecond)
				m.Proxy.NotifyF("§d[ConnSpoof] §fOS:§e%d §fModel:§e%s §fInput:§e%d",
					m.Int("deviceos"), m.Text("model"), m.Int("inputmode"))
			}()
		}
	})
}

func (m *ConnectionSpoof) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name())
	m.Proxy.SetConfig(func(c *proxy.Config) {
		c.DeviceOS = 0
		c.DeviceModel = ""
		c.InputMode = 0
	})
}

func (m *ConnectionSpoof) apply() {
	m.Proxy.SetConfig(func(c *proxy.Config) {
		c.DeviceOS = m.Int("deviceos")
		c.DeviceModel = m.Text("model")
		c.InputMode = m.Int("inputmode")
	})
}
