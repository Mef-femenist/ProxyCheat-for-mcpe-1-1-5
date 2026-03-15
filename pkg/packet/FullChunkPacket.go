package packet

type FullChunkPacket struct {
	ChunkX     int32
	ChunkZ     int32
	Data       string
	PacketName string
}

func (*FullChunkPacket) ID() byte {
	return IDFullChunkPacket
}

func (pk *FullChunkPacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.ChunkX)
	w.Varint32(&pk.ChunkZ)
	w.String(&pk.Data)
}

func (pk *FullChunkPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.ChunkX)
	r.Varint32(&pk.ChunkZ)
	r.String(&pk.Data)
}
