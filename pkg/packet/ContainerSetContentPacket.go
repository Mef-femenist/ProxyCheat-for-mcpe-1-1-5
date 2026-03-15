package packet

import (
	"mefproxy/pkg/item"
)

const ContainerSetContentPacket_SPECIAL_INVENTORY = 0
const ContainerSetContentPacket_SPECIAL_ARMOR = 0x78
const ContainerSetContentPacket_SPECIAL_CREATIVE = 0x79
const ContainerSetContentPacket_SPECIAL_HOTBAR = 0x7a
const ContainerSetContentPacket_SPECIAL_FIXED_INVENTORY = 0x7b

type ContainerSetContentPacket struct {
	WindowID  uint32                `json:"WindowID"`
	TargetEid int32                 `json:"TargetEid"` 
	Slots     map[uint32]*item.Item `json:"Slots"`
	Hotbar    map[uint32]int32      `json:"Hotbar"`
}

func (*ContainerSetContentPacket) ID() byte {
	return IDContainerSetContentPacket
}

func (pk *ContainerSetContentPacket) Marshal(w *PacketWriter) {
	w.Varuint32(&pk.WindowID)
	w.Varint32(&pk.TargetEid)
	count := uint32(len(pk.Slots))
	w.Varuint32(&count)
	for s := uint32(0); s < count; s++ {
		w.Item(pk.Slots[s])
	}
	if pk.WindowID == ContainerSetContentPacket_SPECIAL_INVENTORY {
		w.Varuint32(&count)
		for s := uint32(0); s < count; s++ {
			g := pk.Hotbar[s]
			w.Varint32(&g)
		}
	}
}

func (pk *ContainerSetContentPacket) Unmarshal(r *PacketReader) {
	pk.Slots = make(map[uint32]*item.Item)
	pk.Hotbar = make(map[uint32]int32)
	r.Varuint32(&pk.WindowID)
	r.Varint32(&pk.TargetEid)
	var count uint32
	r.Varuint32(&count)
	for s := uint32(0); s < count; s++ {
		pk.Slots[s] = r.Item()
	}
	if pk.WindowID == ContainerSetContentPacket_SPECIAL_INVENTORY {
		r.Varuint32(&count)
		for s := uint32(0); s < count; s++ {
			var g int32
			r.Varint32(&g)
			pk.Hotbar[s] = g
		}
	}
}
