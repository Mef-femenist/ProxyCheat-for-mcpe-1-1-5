package packet

type StartGamePacket struct {
	EntityUniqueID          int32
	EntityRuntimeID         uint32
	PlayerGamemode          int32
	X                       float32
	Y                       float32
	Z                       float32
	Pitch                   float32
	Yaw                     float32
	Seed                    int32
	Dimension               int32
	Generator               int32
	WorldGamemode           int32
	Difficulty              int32
	SpawnX                  int32
	SpawnY                  uint32
	SpawnZ                  int32
	HasAchievementsDisabled bool
	DayCycleStopTime        int32
	EduMode                 bool
	RainLevel               float32
	LightningLevel          float32
	CommandsEnabled         bool
	IsTexturePacksRequired  bool
	Gamerules               uint32
	LevelID                 string
	WorldName               string
	PremiumWorldTemplateID  string
	PacketName              string
}

func (*StartGamePacket) ID() byte {
	return IDStartGamePacket
}

func (pk *StartGamePacket) Marshal(w *PacketWriter) {
	w.Varint32(&pk.EntityUniqueID)
	w.Varuint32(&pk.EntityRuntimeID)
	w.Varint32(&pk.PlayerGamemode)
	w.Float32(&pk.X)
	w.Float32(&pk.Y)
	w.Float32(&pk.Z)
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)
	w.Varint32(&pk.Seed)
	w.Varint32(&pk.Dimension)
	w.Varint32(&pk.Generator)
	w.Varint32(&pk.WorldGamemode)
	w.Varint32(&pk.Difficulty)
	w.Varint32(&pk.SpawnX)
	w.Varuint32(&pk.SpawnY)
	w.Varint32(&pk.SpawnZ)
	w.Bool(&pk.HasAchievementsDisabled)
	w.Varint32(&pk.DayCycleStopTime)
	w.Bool(&pk.EduMode)
	w.Float32(&pk.RainLevel)
	w.Float32(&pk.LightningLevel)
	w.Bool(&pk.CommandsEnabled)
	w.Bool(&pk.IsTexturePacksRequired)
	w.Varuint32(&pk.Gamerules)
	w.String(&pk.LevelID)
	w.String(&pk.WorldName)
	w.String(&pk.PremiumWorldTemplateID)
}

func (pk *StartGamePacket) Unmarshal(r *PacketReader) {
	r.Varint32(&pk.EntityUniqueID)
	r.Varuint32(&pk.EntityRuntimeID)
	r.Varint32(&pk.PlayerGamemode)
	r.Float32(&pk.X)
	r.Float32(&pk.Y)
	r.Float32(&pk.Z)
	r.Float32(&pk.Pitch)
	r.Float32(&pk.Yaw)
	r.Varint32(&pk.Seed)
	r.Varint32(&pk.Dimension)
	r.Varint32(&pk.Generator)
	r.Varint32(&pk.WorldGamemode)
	r.Varint32(&pk.Difficulty)
	r.Varint32(&pk.SpawnX)
	r.Varuint32(&pk.SpawnY)
	r.Varint32(&pk.SpawnZ)
	r.Bool(&pk.HasAchievementsDisabled)
	r.Varint32(&pk.DayCycleStopTime)
	r.Bool(&pk.EduMode)
	r.Float32(&pk.RainLevel)
	r.Float32(&pk.LightningLevel)
	r.Bool(&pk.CommandsEnabled)
	r.Bool(&pk.IsTexturePacksRequired)
	r.Varuint32(&pk.Gamerules)
	r.String(&pk.LevelID)
	r.String(&pk.WorldName)
	r.String(&pk.PremiumWorldTemplateID)
}
