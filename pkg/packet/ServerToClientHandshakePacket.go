package packet

type ServerToClientHandshakePacket struct {
	PublicKey   string
	ServerToken string
}

func (*ServerToClientHandshakePacket) ID() byte {
	return IDServerToClientHandshakePacket
}

func (pk *ServerToClientHandshakePacket) Marshal(w *PacketWriter) {

}

func (pk *ServerToClientHandshakePacket) Unmarshal(r *PacketReader) {
	r.String(&pk.PublicKey)
	r.String(&pk.ServerToken)
}
