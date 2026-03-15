package packet

type CommandStepPacket struct {
	Command     string
	Overload    string
	Uvarint1    uint32
	CurrentStep uint32
	Done        bool
	ClientID    uint64
	InputJson   string
	OutputJson  string
}

func (*CommandStepPacket) ID() byte {
	return IDCommandStepPacket
}

func (pk *CommandStepPacket) Marshal(w *PacketWriter) {
	w.String(&pk.Command)
	w.String(&pk.Overload)
	w.Varuint32(&pk.Uvarint1)
	w.Varuint32(&pk.CurrentStep)
	w.Bool(&pk.Done)
	w.Varuint64(&pk.ClientID)
	w.String(&pk.InputJson)
	w.String(&pk.OutputJson)
	
}

func (pk *CommandStepPacket) Unmarshal(r *PacketReader) {
	r.String(&pk.Command)
	r.String(&pk.Overload)
	r.Varuint32(&pk.Uvarint1)
	r.Varuint32(&pk.CurrentStep)
	r.Bool(&pk.Done)
	r.Varuint64(&pk.ClientID)
	r.String(&pk.InputJson)
	r.String(&pk.OutputJson)
}
