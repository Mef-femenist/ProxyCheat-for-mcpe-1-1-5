package packet

type SetEntityMotionPacket struct {
	PacketName string
	Eid        int32
	MotionX    float32
	MotionY    float32
	MotionZ    float32
}

func (*SetEntityMotionPacket) ID() byte {
	return IDSetEntityMotionPacket
}

func (pk *SetEntityMotionPacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Eid)
	w.Float32(&pk.MotionX)
	w.Float32(&pk.MotionY)
	w.Float32(&pk.MotionZ)
}

func (pk *SetEntityMotionPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Eid)
	r.Float32(&pk.MotionX)
	r.Float32(&pk.MotionY)
	r.Float32(&pk.MotionZ)
}
