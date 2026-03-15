package network

import (
	"math"
	"strings"
	"sync"
	"time"

	"mefproxy/module"
	"mefproxy/pkg/math/vec3"
	"mefproxy/proxy"
)

type AntiBot struct {
	module.Base
	mu          sync.Mutex
	knownBots   map[uint32]bool
	joinTimes   map[uint32]time.Time
	moveCount   map[uint32]int
	lastCheck   time.Time
	detectedCnt int
}

func NewAntiBot() *AntiBot {
	m := &AntiBot{
		Base:      module.NewBase("AntiBot", "Определяет и игнорирует ботов-NPC", module.Network),
		knownBots: make(map[uint32]bool),
		joinTimes: make(map[uint32]time.Time),
		moveCount: make(map[uint32]int),
	}
	m.AddSetting(module.NewBool("namecheck", true))
	m.AddSetting(module.NewBool("movecheck", true))
	m.AddSetting(module.NewBool("distcheck", true))
	m.AddSetting(module.NewSlider("maxdist", 0.0, 0.0, 1.0, 0.05))
	m.AddSetting(module.NewBool("notify", true))
	return m
}

func (m *AntiBot) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *AntiBot) Enable() {
	m.Base.Enable()
	m.knownBots = make(map[uint32]bool)
	m.joinTimes = make(map[uint32]time.Time)
	m.moveCount = make(map[uint32]int)
	m.detectedCnt = 0
	go m.checkLoop()
}

func (m *AntiBot) Disable() {
	m.Base.Disable()
	m.mu.Lock()
	m.knownBots = make(map[uint32]bool)
	m.mu.Unlock()
}

func (m *AntiBot) IsBot(eid uint32) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.knownBots[eid]
}

func (m *AntiBot) checkLoop() {
	for m.IsEnabled() {
		time.Sleep(500 * time.Millisecond)
		if !m.Proxy.IsFullyOnline() {
			continue
		}
		players := m.Proxy.Players()
		myLoc := m.Proxy.MyLocation()

		for _, pl := range players {
			if m.IsBot(pl.EntityRuntimeID) {
				continue
			}
			isBot := false
			reason := ""

			if m.Bool("namecheck") {
				nick := pl.Nick
				if isRandomNick(nick) {
					isBot = true
					reason = "random-nick"
				}
			}

			if !isBot && m.Bool("distcheck") {
				dist := vec3.Distance(myLoc, pl.Location)
				maxD := m.Float("maxdist")
				if maxD <= 0 {
					maxD = 0.0
				}
				if dist <= maxD && dist == 0 {
					isBot = true
					reason = "zero-dist"
				}
			}

			if !isBot && m.Bool("movecheck") {
				m.mu.Lock()
				if jt, ok := m.joinTimes[pl.EntityRuntimeID]; ok {
					if time.Since(jt) > 5*time.Second && m.moveCount[pl.EntityRuntimeID] == 0 {
						isBot = true
						reason = "no-movement"
					}
				} else {
					m.joinTimes[pl.EntityRuntimeID] = time.Now()
				}
				m.mu.Unlock()
			}

			if isBot {
				m.mu.Lock()
				m.knownBots[pl.EntityRuntimeID] = true
				m.detectedCnt++
				cnt := m.detectedCnt
				m.mu.Unlock()
				if m.Bool("notify") {
					m.Proxy.NotifyF("§d[AntiBot] §fБот: §e%s §8(%s) §7[%d обнаружено]", pl.Nick, reason, cnt)
				}
			}
		}
	}
}

func (m *AntiBot) TrackMove(eid uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.moveCount[eid]++
}

func isRandomNick(nick string) bool {
	if len(nick) < 6 {
		return false
	}
	upper, lower, digits := 0, 0, 0
	for _, c := range nick {
		switch {
		case c >= 'A' && c <= 'Z':
			upper++
		case c >= 'a' && c <= 'z':
			lower++
		case c >= '0' && c <= '9':
			digits++
		}
	}
	total := float64(upper + lower + digits)
	if total == 0 {
		return false
	}
	digitRatio := float64(digits) / total
	entropy := calcEntropy(nick)
	botPatterns := []string{"Bot", "bot", "NPC", "npc", "fake", "Fake", "dummy"}
	for _, p := range botPatterns {
		if strings.Contains(nick, p) {
			return true
		}
	}
	return digitRatio > 0.4 || entropy > 3.8
}

func calcEntropy(s string) float64 {
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}
	n := float64(len(s))
	h := 0.0
	for _, cnt := range freq {
		p := float64(cnt) / n
		h -= p * math.Log2(p)
	}
	return h
}
