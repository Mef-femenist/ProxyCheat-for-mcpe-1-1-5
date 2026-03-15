package packet

type SetEntityDataPacket struct {
	EntityRuntimeID uint32
	Metadata        map[uint32]interface{}
}

func (*SetEntityDataPacket) ID() byte {
	return IDSetEntityDataPacket
}

func (pk *SetEntityDataPacket) Marshal(w *PacketWriter) {
	w.Varuint32(&pk.EntityRuntimeID)
	w.EntityMetadata(&pk.Metadata)
}

func (pk *SetEntityDataPacket) Unmarshal(r *PacketReader) {
	r.Varuint32(&pk.EntityRuntimeID)
	r.EntityMetadata(&pk.Metadata)
}
