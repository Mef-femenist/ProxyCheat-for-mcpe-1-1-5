package packet

type AdventureSettingsPacket struct {
	WorldImmutable uint32
	NoPVP          uint32
	NoMVP          uint32
	NoPVM          uint32

	AutoJump     uint32
	AllowFlight  uint32
	NoClip       uint32
	WorldBuilder uint32
	IsFlying     uint32
	Muted        uint32

	Flags          uint32
	UserPermission uint32
}

const (
	AdventureSettingsPacket_PERMISSION_NORMAL     = 0
	AdventureSettingsPacket_PERMISSION_OPERATOR   = 1
	AdventureSettingsPacket_PERMISSION_HOST       = 2
	AdventureSettingsPacket_PERMISSION_AUTOMATION = 3
	AdventureSettingsPacket_PERMISSION_ADMIN      = 4
)

func (*AdventureSettingsPacket) ID() byte {
	return IDAdventureSettingsPacket
}

func (pk *AdventureSettingsPacket) Marshal(w *PacketWriter) {
	pk.Flags |= pk.WorldImmutable
	pk.Flags |= pk.NoPVP << 1
	pk.Flags |= pk.NoPVM << 2
	pk.Flags |= pk.NoMVP << 3
	
	pk.Flags |= pk.AutoJump << 5
	pk.Flags |= pk.AllowFlight << 6
	pk.Flags |= pk.NoClip << 7
	pk.Flags |= pk.WorldBuilder << 8
	pk.Flags |= pk.IsFlying << 9
	pk.Flags |= pk.Muted << 10

	w.Varuint32(&pk.Flags)
	w.Varuint32(&pk.UserPermission)
}

func (pk *AdventureSettingsPacket) Unmarshal(r *PacketReader) {

	r.Varuint32(&pk.Flags)
	r.Varuint32(&pk.UserPermission)

	pk.WorldImmutable = pk.Flags & 1
	pk.NoPVP = pk.Flags & (1 << 1)
	pk.NoPVM = pk.Flags & (1 << 2)
	pk.NoMVP = pk.Flags & (1 << 3)

	pk.AutoJump = pk.Flags & (1 << 5)
	pk.AllowFlight = pk.Flags & (1 << 6)
	pk.NoClip = pk.Flags & (1 << 7)
	pk.WorldBuilder = pk.Flags & (1 << 8)
	pk.IsFlying = pk.Flags & (1 << 9)
	pk.Muted = pk.Flags & (1 << 9)
}
