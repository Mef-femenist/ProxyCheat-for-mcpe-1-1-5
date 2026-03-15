package packet

type RequestChunkRadiusPacket struct {
	Radius     int32
	PacketName string
}

func (*RequestChunkRadiusPacket) ID() byte {
	return IDRequestChunkRadiusPacket
}

func (pk *RequestChunkRadiusPacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Radius)
}

func (pk *RequestChunkRadiusPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Radius)
}
