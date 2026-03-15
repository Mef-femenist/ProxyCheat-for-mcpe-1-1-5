package packet

func Register(id byte, pk func() DataPacket) {
	RegisteredPackets[id] = pk
}

var RegisteredPackets = map[byte]func() DataPacket{}

type PacketPool map[byte]func() DataPacket

var Pool PacketPool

func NewPool() PacketPool {
	p := PacketPool{}
	for id, pk := range RegisteredPackets {
		p[id] = pk
	}
	return p
}

func init() {
	
	packets := map[byte]func() DataPacket{
		IDTextPacket:                       func() DataPacket { return &TextPacket{} },
		IDLoginPacket:                      func() DataPacket { return &LoginPacket{} },
		IDPlayStatusPacket:                 func() DataPacket { return &PlayStatusPacket{} },
		IDResourcePackDataInfoPacket:       func() DataPacket { return &ResourcePackDataInfoPacket{} },
		IDResourcePacksInfoPacket:          func() DataPacket { return &ResourcePacksInfoPacket{} },
		IDResourcePackClientResponsePacket: func() DataPacket { return &ResourcePackClientResponsePacket{} },
		IDResourcePackStackPacket:          func() DataPacket { return &ResourcePackStackPacket{} },
		IDStartGamePacket:                  func() DataPacket { return &StartGamePacket{} },
		IDDisconnectPacket:                 func() DataPacket { return &DisconnectPacket{} },
		IDRequestChunkRadiusPacket:         func() DataPacket { return &RequestChunkRadiusPacket{} },
		IDRespawnPacket:                    func() DataPacket { return &RespawnPacket{} },
		IDAvailableCommandsPacket:          func() DataPacket { return &AvailableCommandsPacket{} },
		IDFullChunkPacket:                  func() DataPacket { return &FullChunkPacket{} },
		IDTransferPacket:                   func() DataPacket { return &TransferPacket{} },
		IDInteractPacket:                   func() DataPacket { return &InteractPacket{} },
		IDAnimatePacket:                    func() DataPacket { return &AnimatePacket{} },
		IDPlayerListPacket:                 func() DataPacket { return &PlayerListPacket{} },
		IDContainerSetSlotPacket:           func() DataPacket { return &ContainerSetSlotPacket{} },
		IDContainerSetContentPacket:        func() DataPacket { return &ContainerSetContentPacket{} },
		IDSetEntityDataPacket:              func() DataPacket { return &SetEntityDataPacket{} },
		IDPlayerActionPacket:               func() DataPacket { return &PlayerActionPacket{} },
		IDMobEquipmentPacket:               func() DataPacket { return &MobEquipmentPacket{} },
		IDSetEntityMotionPacket:            func() DataPacket { return &SetEntityMotionPacket{} },
		IDUpdateBlockPacket:                func() DataPacket { return &UpdateBlockPacket{} },
		IDPlayerMovePacket:                 func() DataPacket { return &PlayerMovePacket{} },
		IDCommandStepPacket:                func() DataPacket { return &CommandStepPacket{} },
		IDRiderJumpPacket:                  func() DataPacket { return &RiderJumpPacket{} },
		IDResourcePackChunkDataPacket:      func() DataPacket { return &ResourcePackChunkDataPacket{} },
		IDResourcePackChunkRequestPacket:   func() DataPacket { return &ResourcePackChunkRequestPacket{} },
		IDClientBoundMapItemDataPacket:     func() DataPacket { return &ClientBoundMapItemDataPacket{} },
		IDServerToClientHandshakePacket:    func() DataPacket { return &ServerToClientHandshakePacket{} },
		IDClientToServerHandshakePacket:    func() DataPacket { return &ClientToServerHandshakePacket{} },
		IDSetTitlePacket:                   func() DataPacket { return &SetTitlePacket{} },
		IDSetPlayerGameTypePacket:          func() DataPacket { return &SetPlayerGameTypePacket{} },
		IDEntityMovePacket:                 func() DataPacket { return &EntityMovePacket{} },
		IDUpdateAttributesPacket:           func() DataPacket { return &UpdateAttributesPacket{} },
		IDAdventureSettingsPacket:          func() DataPacket { return &AdventureSettingsPacket{} },
		IDMobArmorEquipmentPacket:          func() DataPacket { return &MobArmorEquipmentPacket{} },
		IDCraftingEventPacket:              func() DataPacket { return &CraftingEventPacket{} },
		IDBlockEntityDataPacket:            func() DataPacket { return &BlockEntityDataPacket{} },
		IDLevelEventPacket:                 func() DataPacket { return &LevelEventPacket{} },
		IDRemoveEntityPacket:               func() DataPacket { return &RemoveEntityPacket{} },
		IDAddPlayerPacket:                  func() DataPacket { return &AddPlayerPacket{} },
		IDAddEntityPacket:                  func() DataPacket { return &AddEntityPacket{} },
	}
	for id, pk := range packets {
		Register(id, pk)
	}
	Pool = NewPool()
	
}
