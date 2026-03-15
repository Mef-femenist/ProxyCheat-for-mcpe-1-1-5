package packet

type RemoveEntityPacket struct {
	EntityRuntimeID uint32
}

func (*RemoveEntityPacket) ID() byte {
	return IDRemoveEntityPacket
}

func (pk *RemoveEntityPacket) Marshal(w *PacketWriter) {
	w.Varuint32(&pk.EntityRuntimeID)
}

func (pk *RemoveEntityPacket) Unmarshal(r *PacketReader) {
	r.Varuint32(&pk.EntityRuntimeID)
}
