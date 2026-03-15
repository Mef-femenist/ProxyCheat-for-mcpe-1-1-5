package packet

import "mefproxy/pkg/math"

type LevelEventPacket struct {
	Evid, Data int32
	Pos        *mgl32.Vec3
}

func (*LevelEventPacket) ID() byte {
	return IDLevelEventPacket
}

func (pk *LevelEventPacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Evid)
	w.Vec3(pk.Pos)
	w.Varint32(&pk.Data)
}

func (pk *LevelEventPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Evid)
	r.Vec3(pk.Pos)
	r.Varint32(&pk.Data)
}
