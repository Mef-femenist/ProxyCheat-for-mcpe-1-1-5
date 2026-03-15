package packet

const ()

type ResourcePackDataInfoPacket struct {
	PackID             string
	MaxChunkSize       int32
	ChunkCount         int32
	CompressedPackSize int64
	Sha512             string
	PacketName         string
}

func (*ResourcePackDataInfoPacket) ID() byte {
	return IDResourcePackDataInfoPacket
}

func (pk *ResourcePackDataInfoPacket) Marshal(w *PacketWriter) {
	
}

func (pk *ResourcePackDataInfoPacket) Unmarshal(r *PacketReader) {
	r.String(&pk.PackID)
	r.Int32(&pk.MaxChunkSize)
	r.Int32(&pk.ChunkCount)
	r.Int64(&pk.CompressedPackSize)
	r.String(&pk.Sha512)
}
