package packet

type EntityMovePacket struct {
	EntityRuntimeID uint32
	X               float32
	Y               float32
	Z               float32
	Yaw             float32
	HeadYaw         float32
	Pitch           float32
	OnGround        bool
	Teleported      bool
}

func (*EntityMovePacket) ID() byte {
	return IDEntityMovePacket
}

func (pk *EntityMovePacket) Marshal(w *PacketWriter) {
	w.Varuint32(&pk.EntityRuntimeID)
	w.Float32(&pk.X)
	w.Float32(&pk.Y)
	w.Float32(&pk.Z)
	w.ByteRotation(&pk.Pitch)
	w.ByteRotation(&pk.Yaw)
	w.ByteRotation(&pk.HeadYaw)
	w.Bool(&pk.OnGround)
	w.Bool(&pk.Teleported)
}

func (pk *EntityMovePacket) Unmarshal(r *PacketReader) {
	r.Varuint32(&pk.EntityRuntimeID)
	r.Float32(&pk.X)
	r.Float32(&pk.Y)
	r.Float32(&pk.Z)
	r.ByteRotation(&pk.Pitch)
	r.ByteRotation(&pk.Yaw)
	r.ByteRotation(&pk.HeadYaw)
	r.Bool(&pk.OnGround)
	r.Bool(&pk.Teleported)
}
