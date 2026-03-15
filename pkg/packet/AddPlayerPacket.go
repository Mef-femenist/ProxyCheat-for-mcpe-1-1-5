package packet

import (
	"mefproxy/pkg/math"
	"github.com/google/uuid"
)

type AddPlayerPacket struct {
	UUID            uuid.UUID
	Nick            string
	EntityUniqueID  int32
	EntityRuntimeID uint32
	Position        mgl32.Vec3
}

func (*AddPlayerPacket) ID() byte {
	return IDAddPlayerPacket
}

func (pk *AddPlayerPacket) Marshal(w *PacketWriter) {
	
}

func (pk *AddPlayerPacket) Unmarshal(r *PacketReader) {
	r.UUID(&pk.UUID)
	r.String(&pk.Nick)
	r.Varint32(&pk.EntityUniqueID)
	r.Varuint32(&pk.EntityRuntimeID)
	r.Vec3(&pk.Position)
	
}
