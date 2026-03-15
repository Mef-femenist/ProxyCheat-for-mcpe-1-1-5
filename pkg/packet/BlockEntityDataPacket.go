package packet

import (
	"mefproxy/pkg/nbt"
)

type BlockEntityDataPacket struct {
	X   int32
	Y   uint32
	Z   int32
	NBT map[string]interface{}
}

func (*BlockEntityDataPacket) ID() byte {
	return IDBlockEntityDataPacket
}

func (pk *BlockEntityDataPacket) Marshal(w *PacketWriter) {
	w.BlockCoords(&pk.X, &pk.Y, &pk.Z)
	w.NBT(&pk.NBT, nbt.NetworkLittleEndian)
}

func (pk *BlockEntityDataPacket) Unmarshal(r *PacketReader) {
	pk.NBT = make(map[string]interface{})
	r.BlockCoords(&pk.X, &pk.Y, &pk.Z)
	
	r.NBT(&pk.NBT, nbt.NetworkLittleEndian)
}
