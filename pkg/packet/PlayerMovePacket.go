package packet

const PlayerMovePacket_MODE_NORMAL = 0
const PlayerMovePacket_MODE_RESET = 1
const PlayerMovePacket_MODE_ROTATION = 2

type PlayerMovePacket struct {
	PacketName string
	Eid        int32
	X          float32
	Y          float32
	Z          float32
	Yaw        float32
	BodyYaw    float32
	Pitch      float32
	Mode       byte
	OnGround   bool
	Eid2       int32
}

func (*PlayerMovePacket) ID() byte {
	return IDPlayerMovePacket
}

func (pk *PlayerMovePacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Eid)
	w.Float32(&pk.X)
	w.Float32(&pk.Y)
	w.Float32(&pk.Z)
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)
	w.Float32(&pk.BodyYaw)
	w.Uint8(&pk.Mode)
	w.Bool(&pk.OnGround)
	w.Varint32(&pk.Eid2)
}

func (pk *PlayerMovePacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Eid)
	r.Float32(&pk.X)
	r.Float32(&pk.Y)
	r.Float32(&pk.Z)
	r.Float32(&pk.Pitch)
	r.Float32(&pk.Yaw)
	r.Float32(&pk.BodyYaw)
	r.Uint8(&pk.Mode)
	r.Bool(&pk.OnGround)
	r.Varint32(&pk.Eid2)
}
