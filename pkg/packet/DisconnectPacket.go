package packet

type DisconnectPacket struct {
	HideDisconnectionScreen bool
	Message                 string
	PacketName              string
}

func (*DisconnectPacket) ID() byte {
	return IDDisconnectPacket
}

func (pk *DisconnectPacket) Marshal(w *PacketWriter) {
	w.Bool(&pk.HideDisconnectionScreen)
	w.String(&pk.Message)
}

func (pk *DisconnectPacket) Unmarshal(r *PacketReader) {
	r.Bool(&pk.HideDisconnectionScreen)
	r.String(&pk.Message)
}
