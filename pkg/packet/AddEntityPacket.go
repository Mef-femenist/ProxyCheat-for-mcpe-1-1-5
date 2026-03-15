package packet

import (
	"mefproxy/pkg/math"
	"mefproxy/pkg/entity"
)

type AddEntityPacket struct {
	EntityUniqueID  int32
	EntityRuntimeID uint32
	Type            uint32
	Position        mgl32.Vec3
	Speed           mgl32.Vec3
	Yaw, Pitch      float32
	Attributes      []entity.Attribute
	Metadata        map[uint32]interface{}
	Links           []struct {
		EID, UID int32
		Other    byte
	}
}

func (*AddEntityPacket) ID() byte {
	return IDAddEntityPacket
}

func (pk *AddEntityPacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.EntityUniqueID)
	w.Varuint32(&pk.EntityRuntimeID)
	w.Varuint32(&pk.Type)
	w.Vec3(&pk.Position)
	w.Vec3(&pk.Speed)
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)

	var count = uint32(len(pk.Attributes))
	w.Varuint32(&count)
	for _, a := range pk.Attributes {
		w.String(&a.Name)
		w.Float32(&a.MinValue)
		w.Float32(&a.MaxValue)
		w.Float32(&a.DefaultValue)
	}
	w.EntityMetadata(&pk.Metadata)
	count = uint32(len(pk.Links))
	for _, l := range pk.Links {
		w.Varint32(&l.UID)
		w.Varint32(&l.EID)
		w.Uint8(&l.Other)
	}
}

func (pk *AddEntityPacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.EntityUniqueID)
	r.Varuint32(&pk.EntityRuntimeID)
	r.Varuint32(&pk.Type)
	r.Vec3(&pk.Position)
	r.Vec3(&pk.Speed)
	r.Float32(&pk.Pitch)
	r.Float32(&pk.Yaw)
	var count uint32
	r.Varuint32(&count)

	var name string
	var v1, v2, v3 float32
	for i := uint32(0); i < count; i++ {
		r.String(&name)
		r.Float32(&v1)
		r.Float32(&v2)
		r.Float32(&v3)
		pk.Attributes = append(pk.Attributes, entity.NewAttribute(0, name, v1, v2, v3))
	}
	r.EntityMetadata(&pk.Metadata)
	r.Varuint32(&count)
	for i := uint32(0); i < count; i++ {
		var eid, uid int32
		var unk byte
		r.Varint32(&eid)
		r.Varint32(&uid)
		r.Uint8(&unk)
		pk.Links = append(pk.Links, struct {
			EID, UID int32
			Other    byte
		}{eid, uid, unk})
	}
}
