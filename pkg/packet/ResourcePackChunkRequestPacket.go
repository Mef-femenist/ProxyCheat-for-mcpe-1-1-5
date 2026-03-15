package packet

type ResourcePackChunkRequestPacket struct {
	PacketName string
	PackID     string
	ChunkIndex int32
}

func (*ResourcePackChunkRequestPacket) ID() byte {
	return IDResourcePackChunkRequestPacket
}

func (pk *ResourcePackChunkRequestPacket) Marshal(w *PacketWriter) {
	w.String(&pk.PackID)
	w.Int32(&pk.ChunkIndex)
}

func (pk *ResourcePackChunkRequestPacket) Unmarshal(r *PacketReader) {

}
