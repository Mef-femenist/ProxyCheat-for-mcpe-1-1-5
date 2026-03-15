package main

import (
	"strings"

	"mefproxy/proxy"
)

func cmdServer(cmd string, args []string) bool {
	switch cmd {
	case "server":
		if len(args) == 0 {
			px.Notify("§8│ §7Текущий сервер: §e" + px.GetConfig().Address)
			return true
		}
		addr := args[0]
		if !strings.Contains(addr, ":") {
			px.Notify("§c✗ §fФормат: §e.server §b<ip:порт>")
			return true
		}
		px.Notify("§a✔ §fПереключение → §e" + addr)
		px.SwitchServer(addr)
		return true

	case "nick":
		if len(args) == 0 {
			nick := px.GetConfig().Nick
			if nick == "" {
				px.Notify("§8│ §7Ник не задан §8(ник аккаунта)")
			} else {
				px.Notify("§8│ §7Текущий ник: §f" + nick)
			}
			return true
		}
		nick := strings.Join(args, " ")
		px.SetConfig(func(c *proxy.Config) { c.Nick = nick })
		px.NotifyF("§a✔ §fНик: §e%s §8(следующий вход)", nick)
		return true
	}
	return false
}
