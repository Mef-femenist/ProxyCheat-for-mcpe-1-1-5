package packet

type PlayerActionPacket struct {
	Eid        int32
	Action     int32
	X          int32
	Y          uint32
	Z          int32
	Face       int32
	PacketName string
}

func (*PlayerActionPacket) ID() byte {
	return IDPlayerActionPacket
}

func (pk *PlayerActionPacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Eid)
	w.Varint32(&pk.Action)
	w.BlockCoords(&pk.X, &pk.Y, &pk.Z)
	w.Varint32(&pk.Face)
}

func (pk *PlayerActionPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Eid)
	r.Varint32(&pk.Action)
	r.BlockCoords(&pk.X, &pk.Y, &pk.Z)
	r.Varint32(&pk.Face)
}
