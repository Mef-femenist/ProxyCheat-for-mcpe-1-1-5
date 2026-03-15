package packet

type InteractPacket struct {
	Action byte
	Target uint32
}

func (*InteractPacket) ID() byte {
	return IDInteractPacket
}

func (pk *InteractPacket) Marshal(w *PacketWriter) {
	w.Uint8(&pk.Action)
	w.Varuint32(&pk.Target)
}

func (pk *InteractPacket) Unmarshal(r *PacketReader) {
	r.Uint8(&pk.Action)
	r.Varuint32(&pk.Target)
}
