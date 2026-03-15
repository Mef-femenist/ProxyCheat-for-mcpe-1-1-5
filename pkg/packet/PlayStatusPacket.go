package packet

const (
	PlayStatusPacket_LOGIN_SUCCESS               = 0
	PlayStatusPacket_LOGIN_FAILED_CLIENT         = 1
	PlayStatusPacket_LOGIN_FAILED_SERVER         = 2
	PlayStatusPacket_PLAYER_SPAWN                = 3
	PlayStatusPacket_LOGIN_FAILED_INVALID_TENANT = 4
	PlayStatusPacket_LOGIN_FAILED_VANILLA_EDU    = 5
	PlayStatusPacket_LOGIN_FAILED_EDU_VANILLA    = 6
)

type PlayStatusPacket struct {
	Status     int32
	PacketName string
}

func (*PlayStatusPacket) ID() byte {
	return IDPlayStatusPacket
}

func (pk *PlayStatusPacket) Marshal(w *PacketWriter) {
	w.BEInt32(&pk.Status)
}

func (pk *PlayStatusPacket) Unmarshal(r *PacketReader) {
	r.BEInt32(&pk.Status)
}
