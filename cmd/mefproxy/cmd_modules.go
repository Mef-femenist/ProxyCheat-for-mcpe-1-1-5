package main

import (
	"mefproxy/module"
	"mefproxy/pkg/math/vec3"
)

const (
	menuTop   = "§0§l╔════════════════════════════════╗"
	menuBot   = "§0§l╚════════════════════════════════╝"
	menuMid   = "§0§l╠════════════════════════════════╣"
	menuSep   = "§0§l║ §r"
	menuTitle = "§0§l║  §5§l✦ §d§l𝗺𝗲𝗳𝗽𝗿𝗼𝘅𝘆 §5§lv2.0§r §0§l            ║"
)

func box(content string) string {
	return menuSep + content
}

func cmdModules(cmd string, args []string) bool {
	switch cmd {
	case "help":
		px.Notify(menuTop)
		px.Notify(menuTitle)
		px.Notify(menuMid)
		px.Notify(box("§e§lКОМАНДЫ"))
		px.Notify(box("§a.toggle §f<модуль>    §7вкл/выкл"))
		px.Notify(box("§a.modules             §7все модули"))
		px.Notify(box("§a.status              §7статус прокси"))
		px.Notify(box("§a.players §f/ §a.near    §7игроки рядом"))
		px.Notify(box("§a.server §f<ip:port>   §7смена сервера"))
		px.Notify(box("§a.nick §f<ник>         §7смена ника"))
		px.Notify(box("§a.set §f<м> §b<п> §e<зн> §7настройка"))
		px.Notify(box("§a.settings §f<модуль>  §7параметры"))
		px.Notify(box("§a.version §f[ver]      §7спуф версии"))
		px.Notify(box("§a.versions            §7список версий"))
		px.Notify(menuBot)
		return true

	case "toggle":
		if len(args) == 0 {
			px.Notify("§c✗ §fИспользование: §e.toggle §b<имя_модуля>")
			return true
		}
		name := args[0]
		if !mgr.Toggle(name) {
			px.Notify("§c✗ §fМодуль не найден: §e" + name)
			return true
		}
		m := mgr.Get(name)
		if m.IsEnabled() {
			px.NotifyF("§a✔ %s§l%s §r§a§lвключён", m.Category().Color(), m.Name())
		} else {
			px.NotifyF("§c✘ %s§l%s §r§c§lвыключен", m.Category().Color(), m.Name())
		}
		return true

	case "modules":
		cats := []module.Category{
			module.Combat, module.Movement, module.Exploit,
			module.Visual, module.Misc, module.Network,
		}
		catIcons := map[module.Category]string{
			module.Combat:   "⚔",
			module.Movement: "🏃",
			module.Exploit:  "⚡",
			module.Visual:   "👁",
			module.Misc:     "⚙",
			module.Network:  "🌐",
		}
		px.Notify(menuTop)
		px.Notify(menuTitle)
		px.Notify(menuMid)
		for _, cat := range cats {
			mods := mgr.ByCategory(cat)
			if len(mods) == 0 {
				continue
			}
			icon := catIcons[cat]
			px.NotifyF("§0§l║ §r%s§l%s %s§r", cat.Color(), icon, cat)
			var line string
			for i, mod := range mods {
				dot := "§c●"
				if mod.IsEnabled() {
					dot = "§a●"
				}
				entry := dot + " " + cat.Color() + mod.Name()
				if i < len(mods)-1 {
					entry += "§8, "
				}
				line += entry
				if len(line) > 160 {
					px.Notify(box("  " + line))
					line = ""
				}
			}
			if line != "" {
				px.Notify(box("  " + line))
			}
		}
		px.Notify(menuBot)
		return true

	case "status":
		px.Notify(menuTop)
		px.Notify(menuTitle)
		px.Notify(menuMid)
		cfg := px.GetConfig()
		px.Notify(box("§b§lСТАТУС СОЕДИНЕНИЯ"))
		px.Notify(box("§7Сервер  §f" + cfg.Address))
		nick := cfg.Nick
		if nick == "" {
			nick = "§8(ник аккаунта)"
		}
		px.Notify(box("§7Ник     §f" + nick))
		ver := cfg.GameVersion
		if ver == "" {
			ver = "§8off"
		} else {
			ver = "§a" + ver
		}
		px.Notify(box("§7Версия  " + ver))
		px.Notify(menuMid)
		px.Notify(box("§e§lАКТИВНЫЕ МОДУЛИ"))
		active := mgr.StatusLine()
		px.Notify(box(active))
		px.Notify(menuMid)
		px.Notify(box("§d§lИГРОКИ"))
		px.NotifyF("§0§l║ §r§7Рядом:  §f%d §7игрок(ов)", px.PlayerCount())
		if np := px.NearestPlayer(); np != nil {
			d := vec3.Distance(px.MyLocation(), np.Location)
			px.NotifyF("§0§l║ §r§7Ближний: §f%s §8(§7%.1fм§8)", np.Nick, d)
		}
		px.Notify(menuBot)
		return true

	case "near":
		np := px.NearestPlayer()
		if np == nil {
			px.Notify("§8✦ §7Рядом нет игроков")
			return true
		}
		d := vec3.Distance(px.MyLocation(), np.Location)
		px.NotifyF("§d✦ §fБлижайший: §e%s §8│ §7%.1fм", np.Nick, d)
		return true

	case "players":
		players := px.Players()
		if len(players) == 0 {
			px.Notify("§8✦ §7Рядом нет игроков")
			return true
		}
		px.Notify(menuTop)
		px.NotifyF("§0§l║  §d§l✦ §f§lИГРОКИ §8(§f%d§8)§r §0§l                  ║", len(players))
		px.Notify(menuMid)
		for i, pl := range players {
			d := vec3.Distance(px.MyLocation(), pl.Location)
			px.NotifyF("§0§l║ §r§e%d. §f%s §8│ §7%.1fм §8│ §8EID:%d", i+1, pl.Nick, d, pl.EntityRuntimeID)
		}
		px.Notify(menuBot)
		return true
	}
	return false
}
