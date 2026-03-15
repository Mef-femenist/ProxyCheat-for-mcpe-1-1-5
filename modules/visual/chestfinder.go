package visual

import (
	"time"

	"mefproxy/module"
	mgl32 "mefproxy/pkg/math"
	"mefproxy/pkg/packet"
	"mefproxy/proxy"
)

type ChestFinder struct {
	module.Base
	chests []chestPos
}

type chestPos struct {
	x, y, z float32
}

func NewChestFinder() *ChestFinder {
	return &ChestFinder{Base: module.NewBase("ChestFinder", "Подсвечивает блок-энтити частицами", module.Visual)}
}

func (m *ChestFinder) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *ChestFinder) Enable() {
	m.Base.Enable()
	m.Proxy.OnServerPacket(m.Name()+"_data", func(pk []byte) bool {
		if pk[0] != packet.IDBlockEntityDataPacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		bedp, ok := decoded.(*packet.BlockEntityDataPacket)
		if !ok {
			return false
		}
		m.chests = append(m.chests, chestPos{
			x: float32(bedp.X),
			y: float32(bedp.Y),
			z: float32(bedp.Z),
		})
		return false
	})
	go m.render()
}

func (m *ChestFinder) Disable() {
	m.Base.Disable()
	m.Proxy.RemoveHooks(m.Name() + "_data")
	m.chests = nil
}

func (m *ChestFinder) render() {
	for m.IsEnabled() {
		time.Sleep(500 * time.Millisecond)
		if !m.Proxy.IsClientOnline() {
			continue
		}
		for _, c := range m.chests {
			m.Proxy.SendToClient(&packet.LevelEventPacket{
				Evid: 0x4000 | 11,
				Data: 0,
				Pos:  &mgl32.Vec3{c.x + 0.5, c.y + 1.0, c.z + 0.5},
			})
		}
	}
}
