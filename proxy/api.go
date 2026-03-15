package proxy

import (
	"bytes"
	"sort"
	"time"

	"mefproxy/pkg/entity"
	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
)

func (p *Proxy) SendToClient(pk packet.DataPacket) {
	cc := p.getClient()
	if cc == nil {
		return
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(pk.ID())
	pk.Marshal(packet.NewWriter(buf, 2))
	cc.sendPacketsRaw(map[int][]byte{0: buf.Bytes()})
}

func (p *Proxy) SendToServer(pk packet.DataPacket) {
	sc := p.getServer()
	if sc == nil {
		return
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(pk.ID())
	pk.Marshal(packet.NewWriter(buf, 2))
	sc.sendPacketsRaw(map[int][]byte{0: buf.Bytes()})
}

func (p *Proxy) SendRawToClient(raw []byte) {
	cc := p.getClient()
	if cc == nil {
		return
	}
	cc.sendPacketsRaw(map[int][]byte{0: raw})
}

func (p *Proxy) SendRawToServer(raw []byte) {
	sc := p.getServer()
	if sc == nil {
		return
	}
	sc.sendPacketsRaw(map[int][]byte{0: raw})
}

func (p *Proxy) SendBatchToClient(pks []packet.DataPacket) {
	cc := p.getClient()
	if cc == nil {
		return
	}
	batch := make(map[int][]byte, len(pks))
	for i, pk := range pks {
		var buf bytes.Buffer
		buf.WriteByte(pk.ID())
		pk.Marshal(packet.NewWriter(&buf, 2))
		batch[i] = buf.Bytes()
	}
	cc.sendPacketsRaw(batch)
}

func (p *Proxy) SendBatchToServer(pks []packet.DataPacket) {
	sc := p.getServer()
	if sc == nil {
		return
	}
	batch := make(map[int][]byte, len(pks))
	for i, pk := range pks {
		var buf bytes.Buffer
		buf.WriteByte(pk.ID())
		pk.Marshal(packet.NewWriter(&buf, 2))
		batch[i] = buf.Bytes()
	}
	sc.sendPacketsRaw(batch)
}

func (p *Proxy) SendChat(msg string) {
	p.SendToServer(&packet.TextPacket{
		TextType:   packet.TextPacket_TextTypeChat,
		SourceName: "",
		Message:    msg,
	})
}

func (p *Proxy) Notify(msg string) {
	cc := p.getClient()
	if cc == nil {
		return
	}
	p.SendToClient(&packet.TextPacket{
		TextType:   packet.TextPacket_TextTypeChat,
		SourceName: "",
		Message:    msg,
	})
}

func (p *Proxy) NotifyF(format string, args ...interface{}) {
	p.Notify(sprintf(format, args...))
}

func (p *Proxy) NotifyTitle(title, subtitle string, fadeIn, stay, fadeOut int) {
	p.SendToClient(&packet.SetTitlePacket{
		Type:            0,
		Title:           title,
		FadeInDuration:  int32(fadeIn),
		Duration:        int32(stay),
		FadeOutDuration: int32(fadeOut),
	})
	if subtitle != "" {
		p.SendToClient(&packet.SetTitlePacket{
			Type:            1,
			Title:           subtitle,
			FadeInDuration:  int32(fadeIn),
			Duration:        int32(stay),
			FadeOutDuration: int32(fadeOut),
		})
	}
}

func (p *Proxy) NotifyActionBar(msg string) {
	p.SendToClient(&packet.SetTitlePacket{
		Type:            4,
		Title:           msg,
		FadeInDuration:  0,
		Duration:        30,
		FadeOutDuration: 5,
	})
}

func (p *Proxy) MyEID() uint32 {
	return p.State.MyEID()
}

func (p *Proxy) MyLocation() vec3.Vector3 {
	return p.State.MyLocation()
}

func (p *Proxy) Players() []*Player {
	return p.State.Players()
}

func (p *Proxy) NearestPlayer() *Player {
	return p.State.NearestPlayer()
}

func (p *Proxy) NearestPlayers(max int) []*Player {
	return p.State.NearestPlayers(max)
}

func (p *Proxy) PlayerByEID(eid uint32) *Player {
	return p.State.PlayerByEID(eid)
}

func (p *Proxy) PlayerByNick(nick string) *Player {
	return p.State.PlayerByNick(nick)
}

func (p *Proxy) PlayerCount() int {
	return p.State.PlayerCount()
}

func (p *Proxy) PlayersInRange(radius float32) []*Player {
	all := p.State.Players()
	myLoc := p.State.MyLocation()
	var out []*Player
	for _, pl := range all {
		if vec3.Distance(myLoc, pl.Location) <= radius {
			out = append(out, pl)
		}
	}
	return out
}

func (p *Proxy) PlayersInRangeSorted(radius float32) []*Player {
	all := p.PlayersInRange(radius)
	myLoc := p.State.MyLocation()
	sort.Slice(all, func(i, j int) bool {
		return vec3.Distance(myLoc, all[i].Location) < vec3.Distance(myLoc, all[j].Location)
	})
	return all
}

func (p *Proxy) IsClientOnline() bool {
	return p.clientRunning.Load()
}

func (p *Proxy) IsServerOnline() bool {
	return p.serverRunning.Load()
}

func (p *Proxy) IsFullyOnline() bool {
	return p.clientRunning.Load() && p.serverRunning.Load()
}

func (p *Proxy) GetConfig() Config {
	return p.Config.Get()
}

func (p *Proxy) SetConfig(fn func(*Config)) {
	p.Config.mu.Lock()
	fn(p.Config)
	p.Config.mu.Unlock()
}

func (p *Proxy) SwitchServer(addr string) {
	p.Config.SetAddress(addr)
	p.makeTransfer()
}

func (p *Proxy) OnClientPacket(key string, fn PacketHook) {
	p.hooks.addClientPacket(key, fn)
}

func (p *Proxy) OnServerPacket(key string, fn PacketHook) {
	p.hooks.addServerPacket(key, fn)
}

func (p *Proxy) OnConnect(key string, fn EventHook) {
	p.hooks.addConnect(key, fn)
}

func (p *Proxy) OnDisconnect(key string, fn EventHook) {
	p.hooks.addDisconnect(key, fn)
}

func (p *Proxy) OnServerDisconnect(key string, fn EventHook) {
	p.hooks.addSrvDisconnect(key, fn)
}

func (p *Proxy) RemoveHooks(key string) {
	p.hooks.remove(key)
}

func (p *Proxy) SetEntityData(eid uint32, meta map[uint32]interface{}) {
	p.SendToClient(&packet.SetEntityDataPacket{
		EntityRuntimeID: eid,
		Metadata:        meta,
	})
}

func (p *Proxy) SetPlayerHitbox(eid uint32, w, h float32) {
	p.SetEntityData(eid, map[uint32]interface{}{
		54: w,
		55: h,
	})
}

func (p *Proxy) SetMySpeed(val float32) {
	p.SendToClient(&packet.UpdateAttributesPacket{
		EntityRuntimeID: p.MyEID(),
		Attributes: []entity.Attribute{
			{
				MinValue:     -99,
				MaxValue:     999,
				Value:        val,
				DefaultValue: 0.1,
				Name:         "minecraft:movement",
			},
		},
	})
}

func (p *Proxy) SetMyHealth(val float32) {
	p.SendToClient(&packet.UpdateAttributesPacket{
		EntityRuntimeID: p.MyEID(),
		Attributes: []entity.Attribute{
			{
				MinValue:     0,
				MaxValue:     20,
				Value:        val,
				DefaultValue: 20,
				Name:         "minecraft:health",
			},
		},
	})
}

func (p *Proxy) SetMyFood(val float32) {
	p.SendToClient(&packet.UpdateAttributesPacket{
		EntityRuntimeID: p.MyEID(),
		Attributes: []entity.Attribute{
			{
				MinValue:     0,
				MaxValue:     20,
				Value:        val,
				DefaultValue: 20,
				Name:         "minecraft:food",
			},
		},
	})
}

func (p *Proxy) SetAdventureFlags(allowFlight, isFlying uint32) {
	p.SendToClient(&packet.AdventureSettingsPacket{
		AllowFlight: allowFlight,
		IsFlying:    isFlying,
	})
}

func (p *Proxy) AttackEntity(eid uint32) {
	p.SendToServer(&packet.InteractPacket{
		Action: 1,
		Target: eid,
	})
	p.SendToServer(&packet.AnimatePacket{
		Action: 1,
		Eid:    int32(p.MyEID()),
	})
}

func (p *Proxy) SwingArm() {
	p.SendToServer(&packet.AnimatePacket{
		Action: 1,
		Eid:    int32(p.MyEID()),
	})
}

func (p *Proxy) SendMove(x, y, z, yaw, pitch float32, onGround bool) {
	p.SendToServer(&packet.PlayerMovePacket{
		X: x, Y: y, Z: z,
		Yaw: yaw, BodyYaw: yaw, Pitch: pitch,
		Mode:     0,
		OnGround: onGround,
	})
}

func (p *Proxy) SendMoveAtCurrentPos(onGround bool) {
	loc := p.MyLocation()
	p.SendMove(loc.X, loc.Y, loc.Z, loc.Yaw, loc.Pitch, onGround)
}

func (p *Proxy) BreakBlock(x, y, z int32, face int32) {
	p.SendToServer(&packet.PlayerActionPacket{
		Eid:    int32(p.MyEID()),
		Action: 5,
		X:      x,
		Y:      uint32(y),
		Z:      z,
		Face:   face,
	})
}

func (p *Proxy) PlaceBlock(x, y, z int32, face int32) {
	p.SendToServer(&packet.PlayerActionPacket{
		Eid:    int32(p.MyEID()),
		Action: 20,
		X:      x,
		Y:      uint32(y),
		Z:      z,
		Face:   face,
	})
}

func (p *Proxy) SetMotion(eid uint32, mx, my, mz float32) {
	p.SendToServer(&packet.SetEntityMotionPacket{
		Eid:     int32(eid),
		MotionX: mx,
		MotionY: my,
		MotionZ: mz,
	})
}

func (p *Proxy) SendSprint(sprinting bool) {
	action := int32(2)
	if sprinting {
		action = int32(1)
	}
	p.SendToServer(&packet.PlayerActionPacket{
		Eid:    int32(p.MyEID()),
		Action: action,
	})
}

func (p *Proxy) DelayedNotify(msg string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		p.Notify(msg)
	}()
}

func (p *Proxy) GetSpoofedVersion() string {
	return p.Config.Get().GameVersion
}

func (p *Proxy) SetSpoofVersion(ver string) {
	p.Config.SetGameVersion(ver)
}

func (p *Proxy) GetVersionInfo() (spoof string, protocol int32, found bool) {
	ver := p.Config.Get().GameVersion
	if ver == "" {
		return "", 0, false
	}
	proto, ok := GetProtocolForVersion(ver)
	return ver, proto, ok
}
