package packet

type ClientToServerHandshakePacket struct {
}

func (*ClientToServerHandshakePacket) ID() byte {
	return IDClientToServerHandshakePacket
}

func (pk *ClientToServerHandshakePacket) Marshal(w *PacketWriter) {
}

func (pk *ClientToServerHandshakePacket) Unmarshal(r *PacketReader) {
}
