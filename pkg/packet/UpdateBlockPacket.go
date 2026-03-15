package packet

const UpdateBlockPacket_FLAG_NONE = 0b0000
const UpdateBlockPacket_FLAG_NEIGHBORS = 0b0001
const UpdateBlockPacket_FLAG_NETWORK = 0b0010
const UpdateBlockPacket_FLAG_NOGRAPHIC = 0b0100
const UpdateBlockPacket_FLAG_PRIORITY = 0b1000

type UpdateBlockPacket struct {
	PacketName string
	X          int32
	Y          uint32
	Z          int32
	BlockID    uint32
	Flags      uint32
}

func (*UpdateBlockPacket) ID() byte {
	return IDUpdateBlockPacket
}

func (pk *UpdateBlockPacket) Marshal(w *PacketWriter) {
	w.BlockCoords(&pk.X, &pk.Y, &pk.Z)
	w.Varuint32(&pk.BlockID)
	w.Varuint32(&pk.Flags)
}

func (pk *UpdateBlockPacket) Unmarshal(r *PacketReader) {
	r.BlockCoords(&pk.X, &pk.Y, &pk.Z)
	r.Varuint32(&pk.BlockID)
	r.Varuint32(&pk.Flags)
}
