package main

import (
	"sort"

	"mefproxy/proxy"
)

func cmdVersion(cmd string, args []string) bool {
	switch cmd {
	case "version":
		if len(args) == 0 {
			ver := px.GetConfig().GameVersion
			px.Notify(menuTop)
			px.Notify("§0§l║  §5§l✦ §d§lВЕРСИЯ СПУФ§r §0§l                    ║")
			px.Notify(menuMid)
			if ver == "" {
				px.Notify(box("§7Статус: §8выключен"))
				px.Notify(box("§7Передаётся версия клиента MCPE"))
			} else {
				proto, found := proxy.GetProtocolForVersion(ver)
				px.Notify(box("§7Статус: §aактивен"))
				px.Notify(box("§7Версия: §e" + ver))
				if found {
					px.NotifyF("§0§l║ §r§7Протокол: §f%d", proto)
				} else {
					px.Notify(box("§7Протокол: §8неизвестен (кастом)"))
				}
			}
			px.Notify(menuMid)
			px.Notify(box("§8.version §71.21.80  §8— установить"))
			px.Notify(box("§8.version off    §8— выключить"))
			px.Notify(box("§8.versions       §8— все версии"))
			px.Notify(menuBot)
			return true
		}
		ver := args[0]
		if ver == "off" || ver == "reset" || ver == "0" {
			px.SetSpoofVersion("")
			px.Notify("§a✔ §fСпуф версии §c§lвыключен §8(след. вход)")
			return true
		}
		px.SetSpoofVersion(ver)
		proto, found := proxy.GetProtocolForVersion(ver)
		if found {
			px.NotifyF("§a✔ §fВерсия: §e%s §8│ §7протокол: §f%d §8│ §8след. вход", ver, proto)
		} else {
			px.NotifyF("§a✔ §fВерсия: §e%s §8│ §7кастом §8│ §8след. вход", ver)
		}
		return true

	case "versions":
		all := proxy.GetAllVersions()
		sort.Strings(all)
		px.Notify(menuTop)
		px.Notify("§0§l║  §5§l✦ §d§lПОДДЕРЖИВАЕМЫЕ ВЕРСИИ§r §0§l            ║")
		px.Notify(menuMid)
		line := ""
		count := 0
		for _, v := range all {
			entry := "§e" + v + "§8  "
			line += entry
			count++
			if len(line) > 180 {
				px.Notify(box(line))
				line = ""
			}
		}
		if line != "" {
			px.Notify(box(line))
		}
		px.Notify(menuMid)
		px.NotifyF("§0§l║ §r§7Всего: §f%d §7│ §7Кастом: §e.version §b2.0.0", count)
		px.Notify(menuBot)
		return true
	}
	return false
}
