package packet

import uuid2 "github.com/google/uuid"

const (
	PlayerListPacket_TYPE_ADD    = 0
	PlayerListPacket_TYPE_REMOVE = 1
)

type PlayerListPacket struct {
	Entries    map[string][]interface{}
	Type       byte
	PacketName string
}

func (*PlayerListPacket) ID() byte {
	return IDPlayerListPacket
}

func (pk *PlayerListPacket) Marshal(w *PacketWriter) {
	w.Uint8(&pk.Type)
	count := uint32(len(pk.Entries))
	w.Varuint32(&count)
	for uuid, ent := range pk.Entries {
		u, _ := uuid2.Parse(uuid)
		if pk.Type == PlayerListPacket_TYPE_ADD {
			eid := ent[0].(int32)
			s1 := ent[1].(string)
			s2 := ent[2].(string)
			s3 := ent[3].(string)
			w.UUID(&u)
			w.Varint32(&eid)
			w.String(&s1)
			w.String(&s2)
			w.String(&s3)
		} else {
			w.UUID(&u)
		}
	}
}

func (pk *PlayerListPacket) Unmarshal(r *PacketReader) {
	pk.Entries = make(map[string][]interface{})

	r.Uint8(&pk.Type)
	var count uint32
	r.Varuint32(&count)
	for i := 0; i < int(count); i++ {
		if pk.Type == PlayerListPacket_TYPE_ADD {
			var uuid uuid2.UUID
			r.UUID(&uuid)
			var eid int32
			r.Varint32(&eid)
			var s1 string
			var s2 string
			var s3 string
			r.String(&s1)
			r.String(&s2)
			r.String(&s3)
			pk.Entries[uuid.String()] = append(pk.Entries[uuid.String()], eid)
			pk.Entries[uuid.String()] = append(pk.Entries[uuid.String()], s1)
			pk.Entries[uuid.String()] = append(pk.Entries[uuid.String()], s2)
			pk.Entries[uuid.String()] = append(pk.Entries[uuid.String()], s3)
		} else {
			var uuid uuid2.UUID
			r.UUID(&uuid)
			pk.Entries[uuid.String()] = nil
		}

	}
}
