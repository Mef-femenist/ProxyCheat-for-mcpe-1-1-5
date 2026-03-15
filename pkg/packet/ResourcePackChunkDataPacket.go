package packet

type ResourcePackChunkDataPacket struct {
	PacketName string
	PackID     string
	ChunkIndex int32
	Progress   int64
	Data       string
}

func (*ResourcePackChunkDataPacket) ID() byte {
	return IDResourcePackChunkDataPacket
}

func (pk *ResourcePackChunkDataPacket) Marshal(w *PacketWriter) {
	
}

func (pk *ResourcePackChunkDataPacket) Unmarshal(r *PacketReader) {
	r.String(&pk.PackID)
	r.Int32(&pk.ChunkIndex)
	r.Int64(&pk.Progress)
	var leng int32
	r.Int32(&leng)
	bbuf := make([]byte, leng)
	r.BytesLength(bbuf)
	pk.Data = string(bbuf)
}
