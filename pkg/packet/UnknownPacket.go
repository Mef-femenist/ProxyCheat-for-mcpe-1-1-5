package packet

import (
	"fmt"
)

type UnknownPacket struct {
	
	PacketID byte
	
	Payload    []byte
	PacketName string
}

func (pk *UnknownPacket) ID() byte {
	return pk.PacketID
}

func (pk *UnknownPacket) Marshal(w *PacketWriter) {
	w.Bytes(&pk.Payload)
}

func (pk *UnknownPacket) Unmarshal(r *PacketReader) {
	r.Bytes(&pk.Payload)
}

func (pk *UnknownPacket) String() string {
	return fmt.Sprintf("{ID:0x%x Payload:0x%x}", pk.PacketID, pk.Payload)
}
