package packet

import "mefproxy/pkg/resource"

type ResourcePackStackPacket struct {
	MustAccept        bool
	PacketName        string
	BehaviorPackStack []*resource.ResourcePackInfoEntry
	ResourcePackStack []*resource.ResourcePackInfoEntry
}

func (*ResourcePackStackPacket) ID() byte {
	return IDResourcePackStackPacket
}

func (pk *ResourcePackStackPacket) Marshal(w *PacketWriter) {
	w.Bool(&pk.MustAccept)
	bhpcount := int16(len(pk.BehaviorPackStack))
	w.Int16(&bhpcount)
	for _, bhp := range pk.BehaviorPackStack {
		if bhpcount <= 0 {
			break
		}
		bhpcount--
		w.String(&bhp.PackID)
		w.String(&bhp.Version)
	}

	rpcount := int16(len(pk.ResourcePackStack))
	w.Int16(&rpcount)
	for _, rhp := range pk.ResourcePackStack {

		if rpcount <= 0 {
			break
		}
		rpcount--
		w.String(&rhp.PackID)
		w.String(&rhp.Version)
	}
}

func (pk *ResourcePackStackPacket) Unmarshal(r *PacketReader) {
	
	r.Bool(&pk.MustAccept)
	var bhpcount int16
	r.Int16(&bhpcount)
	for {

		if bhpcount <= 0 {
			break
		}
		bhpcount--
		var packid string
		r.String(&packid)
		var version string
		r.String(&version)
		pk.BehaviorPackStack = append(pk.BehaviorPackStack, resource.NewResourcePackInfoEntry(packid, version, 0))
	}

	var rpcount int16
	r.Int16(&rpcount)
	for {

		if rpcount <= 0 {
			break
		}
		rpcount--
		var packid string
		r.String(&packid)
		var version string
		r.String(&version)
		pk.ResourcePackStack = append(pk.ResourcePackStack, resource.NewResourcePackInfoEntry(packid, version, 0))
	}
}
