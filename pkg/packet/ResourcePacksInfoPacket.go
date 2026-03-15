package packet

import (
	"mefproxy/pkg/resource"
)

type ResourcePacksInfoPacket struct {
	MustAccept        bool
	BehaviorPackStack []*resource.ResourcePackInfoEntry
	ResourcePackStack []*resource.ResourcePackInfoEntry
}

func (*ResourcePacksInfoPacket) ID() byte {
	return IDResourcePacksInfoPacket
}

func (pk *ResourcePacksInfoPacket) Marshal(w *PacketWriter) {
	w.Bool(&pk.MustAccept)

	bhpcount := uint16(len(pk.BehaviorPackStack))
	w.Uint16(&bhpcount)
	for _, bhp := range pk.BehaviorPackStack {
		w.String(&bhp.PackID)
		w.String(&bhp.Version)
		w.Uint64(&bhp.PackSize)
	}

	rpcount := uint16(len(pk.ResourcePackStack))
	w.Uint16(&rpcount)
	for _, rhp := range pk.ResourcePackStack {
		w.String(&rhp.PackID)
		w.String(&rhp.Version)
		w.Uint64(&rhp.PackSize)
	}
}

func (pk *ResourcePacksInfoPacket) Unmarshal(r *PacketReader) {
	var length uint16
	var packid string
	var version string
	var psize uint64
	r.Bool(&pk.MustAccept)

	r.Uint16(&length)
	for length > uint16(0) {
		length--

		r.String(&packid)
		r.String(&version)
		r.Uint64(&psize)
		pk.BehaviorPackStack = append(pk.BehaviorPackStack, resource.NewResourcePackInfoEntry(packid, version, psize))
	}

	r.Uint16(&length)
	for length > uint16(0) {
		length--

		r.String(&packid)
		r.String(&version)
		r.Uint64(&psize)
		pk.ResourcePackStack = append(pk.ResourcePackStack, resource.NewResourcePackInfoEntry(packid, version, psize))
	}
	
}
