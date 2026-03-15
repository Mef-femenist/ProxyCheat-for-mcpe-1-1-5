package packet

type RiderJumpPacket struct {
	Unknown    int32
	PacketName string
}

func (*RiderJumpPacket) ID() byte {
	return IDRiderJumpPacket
}

func (pk *RiderJumpPacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Unknown)
}

func (pk *RiderJumpPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Unknown)
}
