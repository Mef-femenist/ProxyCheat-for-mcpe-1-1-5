package main

import (
	"log"
	"runtime"

	"mefproxy/module"
	"mefproxy/modules/combat"
	"mefproxy/modules/exploit"
	"mefproxy/modules/misc"
	"mefproxy/modules/movement"
	"mefproxy/modules/network"
	"mefproxy/modules/visual"
	"mefproxy/proxy"
	"mefproxy/utils"
)

var (
	px  *proxy.Proxy
	mgr *module.Manager
)

func main() {
	log.SetFlags(log.Ltime)
	printBanner()

	cfg := proxy.LoadConfig("config.json")
	px = proxy.New(cfg)
	mgr = module.NewManager(px)

	registerModules()
	registerCommandHook()

	runtime.LockOSThread()
	px.Run()
}

func registerModules() {
	mods := []module.Module{
		combat.NewHitbox(),
		combat.NewKillAura(),
		combat.NewCriticals(),
		combat.NewVelocity(),
		combat.NewReach(),
		combat.NewTPAura(),

		movement.NewNoFall(),
		movement.NewAutoSprint(),
		movement.NewSpeed(),
		movement.NewFlight(),
		movement.NewStep(),

		misc.NewAntiAFK(),
		misc.NewBypass(),
		misc.NewChatLogger(),
		misc.NewAutoReconnect(),
		misc.NewPacketLogger(),
		misc.NewBlink(),

		exploit.NewPhase(),
		exploit.NewNuker(),
		exploit.NewScaffold(),
		exploit.NewEntitySpeed(),
		exploit.NewInventoryMove(),

		visual.NewTracer(),
		visual.NewESP(),
		visual.NewBreadcrumbs(),
		visual.NewChestFinder(),
		visual.NewNametagHUD(),

		network.NewPacketFlooder(),
		network.NewAntiBot(),
		network.NewConnectionSpoof(),
		network.NewPacketRateLimiter(),
	}
	for _, m := range mods {
		mgr.Register(m)
	}
	log.Printf("[mefproxy] Загружено модулей: %d", len(mods))
}

func printBanner() {
	lanIP := utils.GetLANIP()
	log.Printf("\n" +
		"  ╔══════════════════════════════════════════════════════╗\n" +
		"  ║                                                      ║\n" +
		"  ║   ███╗   ███╗███████╗███████╗██████╗ ██████╗ ██╗  ██╗║\n" +
		"  ║   ████╗ ████║██╔════╝██╔════╝██╔══██╗██╔══██╗╚██╗██╔╝║\n" +
		"  ║   ██╔████╔██║█████╗  █████╗  ██████╔╝██████╔╝ ╚███╔╝ ║\n" +
		"  ║   ██║╚██╔╝██║██╔══╝  ██╔══╝  ██╔═══╝ ██╔══██╗ ██╔██╗ ║\n" +
		"  ║   ██║ ╚═╝ ██║███████╗██║     ██║      ██║  ██║██╔╝ ██╗║\n" +
		"  ║   ╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝      ╚═╝  ╚═╝╚═╝  ╚═╝║\n" +
		"  ║                                                      ║\n" +
		"  ║          v 3 . 0   —   M C P E   P r o x y          ║\n" +
		"  ║       Multi-version 1.1.0 — 1.21.90 · RakNet        ║\n" +
		"  ╠══════════════════════════════════════════════════════╣\n" +
		"  ║  Подключись в MCPE → Добавить сервер:               ║\n" +
		"  ║  >>  " + padRight(lanIP+":19132", 46) + "  ||  \n" +
		"  ╚══════════════════════════════════════════════════════╝\n")
}

func padRight(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}
