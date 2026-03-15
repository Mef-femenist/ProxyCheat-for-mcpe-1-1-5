package packet

const (
	AnimatePacket_ACTION_SWING_ARM = 1

	AnimatePacket_ACTION_STOP_SLEEP   = 3
	AnimatePacket_ACTION_CRITICAL_HIT = 4
	AnimatePacket_ACTION_ROW_RIGHT    = 128
	AnimatePacket_ACTION_ROW_LEFT     = 129
)

type AnimatePacket struct {
	PacketName string
	Action     int32
	Eid        int32
	Float      float32
}

func (*AnimatePacket) ID() byte {
	return IDAnimatePacket
}

func (pk *AnimatePacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Action)
	w.Varint32(&pk.Eid)
	if ((int64(pk.Float)) & 0x80) > 0 {
		w.Float32(&pk.Float)
	}

}

func (pk *AnimatePacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Action)
	r.Varint32(&pk.Eid)
	pk.Float = 0.0
}
