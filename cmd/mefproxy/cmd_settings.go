package main

import (
	"fmt"
	"strings"
)

func cmdSettings(cmd string, args []string) bool {
	switch cmd {
	case "set":
		if len(args) < 3 {
			px.Notify("§c✗ §fИспользование: §e.set §b<модуль> §a<настройка> §e<значение>")
			return true
		}
		modName := args[0]
		settingName := args[1]
		value := strings.Join(args[2:], " ")
		m := mgr.Get(modName)
		if m == nil {
			px.Notify("§c✗ §fМодуль не найден: §e" + modName)
			return true
		}
		s := m.GetSetting(settingName)
		if s == nil {
			px.Notify(fmt.Sprintf("§c✗ §fПараметр §b%s §fне найден в §e%s", settingName, modName))
			px.Notify("§8  └ §7Список: §e.settings §b" + modName)
			return true
		}
		if err := s.SetFromString(value); err != nil {
			px.NotifyF("§c✗ §fОшибка: §7%s", err)
			return true
		}
		px.NotifyF("§a✔ %s§l%s §8│ §7%s §8= %s",
			m.Category().Color(), m.Name(), s.Name(), s.DisplayValue())
		return true

	case "settings":
		if len(args) == 0 {
			px.Notify("§c✗ §fИспользование: §e.settings §b<модуль>")
			return true
		}
		m := mgr.Get(args[0])
		if m == nil {
			px.Notify("§c✗ §fМодуль не найден: §e" + args[0])
			return true
		}
		settings := m.Settings()
		px.Notify(menuTop)
		px.NotifyF("§0§l║  %s§l⚙ %s §r§7— настройки§r §0§l             ║", m.Category().Color(), m.Name())
		px.Notify(menuMid)
		if len(settings) == 0 {
			px.Notify(box("§8нет настроек"))
		} else {
			for _, s := range settings {
				px.NotifyF("§0§l║ §r  §b%-14s §8[§7%s§8]  %s",
					s.Name(), s.Type(), s.DisplayValue())
			}
		}
		px.Notify(menuBot)
		return true
	}
	return false
}
