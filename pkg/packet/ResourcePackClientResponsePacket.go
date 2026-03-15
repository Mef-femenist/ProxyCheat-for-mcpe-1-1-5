package packet

const (
	ResourcePackClientResponsePacket_STATUS_REFUSED        = 1
	ResourcePackClientResponsePacket_STATUS_SEND_PACKS     = 2
	ResourcePackClientResponsePacket_STATUS_HAVE_ALL_PACKS = 3
	ResourcePackClientResponsePacket_STATUS_COMPLETED      = 4
)

type ResourcePackClientResponsePacket struct {
	Status     byte
	PackIDs    []string
	PacketName string
}

func (*ResourcePackClientResponsePacket) ID() byte {
	return IDResourcePackClientResponsePacket
}

func (pk *ResourcePackClientResponsePacket) Marshal(w *PacketWriter) {
	w.Uint8(&pk.Status)
	ln := uint16(len(pk.PackIDs))
	w.Uint16(&ln)
	for _, pack := range pk.PackIDs {
		w.String(&pack)
	}
}

func (pk *ResourcePackClientResponsePacket) Unmarshal(r *PacketReader) {
	
	r.Uint8(&pk.Status)
	var ln uint16
	r.Uint16(&ln)
	pk.PackIDs = make([]string, ln)
	for i := uint16(0); i < ln; i++ {
		r.String(&pk.PackIDs[i])
	}
}
