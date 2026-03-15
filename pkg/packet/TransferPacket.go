package packet

type TransferPacket struct {
	Address    string
	Port       uint16
	PacketName string
}

func (*TransferPacket) ID() byte {
	return IDTransferPacket
}

func (pk *TransferPacket) Marshal(w *PacketWriter) {
	w.String(&pk.Address)
	w.Uint16(&pk.Port)
}

func (pk *TransferPacket) Unmarshal(r *PacketReader) {
	r.String(&pk.Address)
	r.Uint16(&pk.Port)
}
