package packet

const (
	SetTitlePacket_TYPE_CLEAR      = 0
	SetTitlePacket_TYPE_RESET      = 1
	SetTitlePacket_TYPE_TITLE      = 2
	SetTitlePacket_TYPE_SUB_TITLE  = 3
	SetTitlePacket_TYPE_ACTION_BAR = 4
	SetTitlePacket_TYPE_TIMES      = 5
)

type SetTitlePacket struct {
	Type            int32
	Title           string
	FadeInDuration  int32
	Duration        int32
	FadeOutDuration int32
}

func (*SetTitlePacket) ID() byte {
	return IDSetTitlePacket
}

func (pk *SetTitlePacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.Type)
	w.String(&pk.Title)
	w.Varint32(&pk.FadeInDuration)
	w.Varint32(&pk.Duration)
	w.Varint32(&pk.FadeOutDuration)
}

func (pk *SetTitlePacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.Type)
	r.String(&pk.Title)
	r.Varint32(&pk.FadeInDuration)
	r.Varint32(&pk.Duration)
	r.Varint32(&pk.FadeOutDuration)
}
