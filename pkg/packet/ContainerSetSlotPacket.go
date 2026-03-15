package packet

import "mefproxy/pkg/item"

type ContainerSetSlotPacket struct {
	WindowID   byte
	Slot       int32
	HotbarSlot int32
	Item       *item.Item
	PacketName string
}

func (*ContainerSetSlotPacket) ID() byte {
	return IDContainerSetSlotPacket
}

func (pk *ContainerSetSlotPacket) Marshal(w *PacketWriter) {
	w.Uint8(&pk.WindowID)
	w.Varint32(&pk.Slot)
	w.Varint32(&pk.HotbarSlot)
	w.Item(pk.Item)
	var unk byte = 0x00
	w.Uint8(&unk)
}

func (pk *ContainerSetSlotPacket) Unmarshal(r *PacketReader) {
	r.Uint8(&pk.WindowID)
	r.Varint32(&pk.Slot)
	r.Varint32(&pk.HotbarSlot)
	pk.Item = r.Item()
	var unk byte
	r.Uint8(&unk)
}
