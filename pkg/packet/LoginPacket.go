package packet

type LoginPacket struct {
	ClientProtocol int32
	
	ConnectionRequest []byte
}

func (*LoginPacket) ID() byte {
	return IDLoginPacket
}

func (pk *LoginPacket) Marshal(w *PacketWriter) {
	w.BEInt32(&pk.ClientProtocol)
	ged := uint8(0)
	w.Uint8(&ged)
	w.ByteSlice(&pk.ConnectionRequest)
}

func (pk *LoginPacket) Unmarshal(r *PacketReader) {
	r.BEInt32(&pk.ClientProtocol)
	ged := uint8(0)
	r.Uint8(&ged)
	r.ByteSlice(&pk.ConnectionRequest)
}
