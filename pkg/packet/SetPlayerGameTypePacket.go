package packet

type SetPlayerGameTypePacket struct {
	Type int32
}

func (*SetPlayerGameTypePacket) ID() byte {
	return IDSetPlayerGameTypePacket
}

func (pk *SetPlayerGameTypePacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Type)
}

func (pk *SetPlayerGameTypePacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Type)
}
