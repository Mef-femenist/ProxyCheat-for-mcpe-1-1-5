package packet

import "mefproxy/pkg/entity"

type UpdateAttributesPacket struct {
	EntityRuntimeID uint32
	Attributes      []entity.Attribute
}

func (*UpdateAttributesPacket) ID() byte {
	return IDUpdateAttributesPacket
}

func (pk *UpdateAttributesPacket) Marshal(w *PacketWriter) {
	w.Varuint32(&pk.EntityRuntimeID)
	cnt := uint32(len(pk.Attributes))
	w.Varuint32(&cnt)
	for _, attr := range pk.Attributes {
		w.Float32(&attr.MinValue)
		w.Float32(&attr.MaxValue)
		w.Float32(&attr.Value)
		w.Float32(&attr.DefaultValue)
		w.String(&attr.Name)
	}
}

func (pk *UpdateAttributesPacket) Unmarshal(r *PacketReader) {
	r.Varuint32(&pk.EntityRuntimeID)
	var cnt uint32
	r.Varuint32(&cnt)
	for i := 0; i < int(cnt); i++ {
		var attr entity.Attribute
		r.Float32(&attr.MinValue)
		r.Float32(&attr.MaxValue)
		r.Float32(&attr.Value)
		r.Float32(&attr.DefaultValue)
		r.String(&attr.Name)
		pk.Attributes = append(pk.Attributes, attr)
	}

}
