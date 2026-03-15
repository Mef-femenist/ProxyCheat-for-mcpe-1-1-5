package packet

import (
	"mefproxy/utils/color"
)

type ClientBoundMapItemDataPacket struct {
	MapID        int32
	Type         uint32
	Eids         []int32
	Scale        byte
	Decoreations map[int32]map[string]interface{}
	Width        int32
	Height       int32
	XOffset      int32
	YOffset      int32
	Colors       map[ClientBoundMapItemDataPacket_Position]color.Color
}

type ClientBoundMapItemDataPacket_Position struct {
	X int32
	Y int32
}

const ClientBoundMapItemDataPacket_BITFLAG_TEXTURE_UPDATE = 0x02
const ClientBoundMapItemDataPacket_BITFLAG_DECORATION_UPDATE = 0x04
const ClientBoundMapItemDataPacket_BITFLAG_ENTITY_UPDATE = 0x08

func (*ClientBoundMapItemDataPacket) ID() byte {
	return IDClientBoundMapItemDataPacket
}

func (pk *ClientBoundMapItemDataPacket) Marshal(w *PacketWriter) {
	
}

func (pk *ClientBoundMapItemDataPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.MapID)
	r.Varuint32(&pk.Type)

	if (pk.Type & ClientBoundMapItemDataPacket_BITFLAG_ENTITY_UPDATE) != 0 {
		var count uint32
		r.Varuint32(&count)
		for i := uint32(0); i < count; i++ {
			var eid int32
			r.Varint32(&eid)
			pk.Eids = append(pk.Eids, eid)
		}
	}

	if (pk.Type & (ClientBoundMapItemDataPacket_BITFLAG_DECORATION_UPDATE | ClientBoundMapItemDataPacket_BITFLAG_TEXTURE_UPDATE)) != 0 {
		r.Uint8(&pk.Scale)
	}
	if (pk.Type & ClientBoundMapItemDataPacket_BITFLAG_DECORATION_UPDATE) != 0 {
		pk.Decoreations = map[int32]map[string]interface{}{}
		var count int32
		r.Varint32(&count)
		for i := int32(0); i < count; i++ {
			var weird int32
			r.Varint32(&weird)
			pk.Decoreations[i]["rot"] = weird & 0x0f
			pk.Decoreations[i]["img"] = weird >> 4
			var offs byte
			r.Uint8(&offs)
			pk.Decoreations[i]["xOffset"] = offs
			r.Uint8(&offs)
			pk.Decoreations[i]["yOffset"] = offs
			var label string
			r.String(&label)
			pk.Decoreations[i]["label"] = label
			var clr int32
			r.Int32(&clr)
			pk.Decoreations[i]["color"] = color.FromARGB(clr)
		}
	}
	if (pk.Type & ClientBoundMapItemDataPacket_BITFLAG_TEXTURE_UPDATE) != 0 {
		pk.Colors = map[ClientBoundMapItemDataPacket_Position]color.Color{}
		r.Varint32(&pk.Width)
		r.Varint32(&pk.Height)
		r.Varint32(&pk.XOffset)
		r.Varint32(&pk.YOffset)
		
		for y := int32(0); y < pk.Height; y++ {
			for x := int32(0); x < pk.Width; x++ {

				var rgba uint32
				r.Varuint32(&rgba)
				
				pk.Colors[ClientBoundMapItemDataPacket_Position{y, x}] = color.FromARGB(int32(rgba))
			}
		}
	}
}
