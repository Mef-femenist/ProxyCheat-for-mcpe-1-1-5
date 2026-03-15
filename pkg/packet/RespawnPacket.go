package packet

type RespawnPacket struct {
	X          float32
	Y          float32
	Z          float32
	PacketName string
}

func (*RespawnPacket) ID() byte {
	return IDRespawnPacket
}

func (pk *RespawnPacket) Marshal(w *PacketWriter) {
	w.Float32(&pk.X)
	w.Float32(&pk.Y)
	w.Float32(&pk.Z)
}

func (pk *RespawnPacket) Unmarshal(r *PacketReader) {
	r.Float32(&pk.X)
	r.Float32(&pk.Y)
	r.Float32(&pk.Z)
}
