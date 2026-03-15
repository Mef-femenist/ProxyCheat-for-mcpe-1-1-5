package packet

const (
	TextPacket_TextTypeRaw = iota
	TextPacket_TextTypeChat
	TextPacket_TextTypeTranslation
	TextPacket_TextTypePopup
	TextPacket_TextTypeJukeboxPopup
	TextPacket_TextTypeTip
	TextPacket_TextTypeSystem
	TextPacket_TextTypeWhisper
	TextPacket_TextTypeAnnouncement
	TextPacket_TextTypeObject
	TextPacket_TextTypeObjectWhisper
)

type TextPacket struct {
	PacketName string

	TextType byte

	SourceName string

	Message string

	Parameters []string
}

func (*TextPacket) ID() byte {
	return IDTextPacket
}

func (pk *TextPacket) Marshal(w *PacketWriter) {

	w.Uint8(&pk.TextType)
	switch pk.TextType {
	case TextPacket_TextTypeChat, TextPacket_TextTypeWhisper, TextPacket_TextTypeAnnouncement:
		w.String(&pk.SourceName)
		w.String(&pk.Message)
	case TextPacket_TextTypeRaw, TextPacket_TextTypeTip, TextPacket_TextTypeSystem, TextPacket_TextTypeObject, TextPacket_TextTypeObjectWhisper:
		w.String(&pk.SourceName)
		w.String(&pk.Message)
	case TextPacket_TextTypeTranslation, TextPacket_TextTypePopup, TextPacket_TextTypeJukeboxPopup:
		var length uint32

		w.String(&pk.Message)
		w.Varuint32(&length)
		pk.Parameters = make([]string, length)
		for i := uint32(0); i < length; i++ {
			w.String(&pk.Parameters[i])
		}
	}
}

func (pk *TextPacket) Unmarshal(r *PacketReader) {

	r.Uint8(&pk.TextType)
	switch pk.TextType {
	case TextPacket_TextTypeChat, TextPacket_TextTypeWhisper, TextPacket_TextTypeAnnouncement:
		r.String(&pk.SourceName)
		r.String(&pk.Message)
	case TextPacket_TextTypeRaw, TextPacket_TextTypeTip, TextPacket_TextTypeSystem, TextPacket_TextTypeObject, TextPacket_TextTypeObjectWhisper:
		r.String(&pk.SourceName)
		r.String(&pk.Message)
	case TextPacket_TextTypeTranslation, TextPacket_TextTypePopup, TextPacket_TextTypeJukeboxPopup:
		var length uint32

		r.String(&pk.Message)
		r.Varuint32(&length)
		pk.Parameters = make([]string, length)
		for i := uint32(0); i < length; i++ {
			r.String(&pk.Parameters[i])
		}
	}
}
