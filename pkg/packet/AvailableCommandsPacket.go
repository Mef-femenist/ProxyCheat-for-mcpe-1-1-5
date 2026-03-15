package packet

type AvailableCommandsPacket struct {
	Commands   string
	Unknown    string
	PacketName string
}

func (*AvailableCommandsPacket) ID() byte {
	return IDAvailableCommandsPacket
}

func (pk *AvailableCommandsPacket) Marshal(w *PacketWriter) {
}

func (pk *AvailableCommandsPacket) Unmarshal(r *PacketReader) {
	r.String(&pk.Commands)
	r.String(&pk.Unknown)
}
