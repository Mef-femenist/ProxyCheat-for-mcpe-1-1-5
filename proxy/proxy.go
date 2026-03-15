package proxy

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"mefproxy/pkg/math/vec3"
	"mefproxy/pkg/packet"
	"mefproxy/pkg/packet/encrypter"
	"mefproxy/pkg/packet/login"
	"mefproxy/pkg/transport"
	"mefproxy/utils"
	"mefproxy/utils/machineid"

	"go.uber.org/atomic"
)

type Proxy struct {
	Config *Config
	State  *State
	hooks  *hookRegistry

	clientMu sync.RWMutex
	client   *clientConn

	serverMu sync.RWMutex
	server   *serverConn

	globalEncrypt *encrypter.Instance

	clientRunning atomic.Bool
	serverRunning atomic.Bool
	bypassRP      atomic.Bool

	localAddress    atomic.String
	currentBindPort atomic.Int32
	clientIDBase    int64

	loginMu        sync.RWMutex
	storedLoginPkt []byte

	reconnecting atomic.Bool

	persistentListener *transport.Listener

	mux sync.Mutex
}

func New(cfg *Config) *Proxy {
	return &Proxy{
		Config:        cfg,
		State:         newState(),
		hooks:         newHookRegistry(),
		globalEncrypt: &encrypter.Instance{IsEnabled: false},
	}
}

func (p *Proxy) Run() {
	if ln, err := net.ListenUDP("udp", &net.UDPAddr{Port: 19132, IP: net.ParseIP("0.0.0.0")}); err != nil {
		log.Fatalln("[mefproxy] Порт 19132 занят. Завершите предыдущую программу: pkill -9 -f mefproxy")
	} else {
		_ = ln.Close()
	}

	rand.Seed(time.Now().Unix())
	p.currentBindPort.Store(int32(rand.Intn(65535-40000) + 40000))
	p.clientIDBase = utils.FixClientIDLen(-rand.Int63())
	p.Config.CurrentClientID = p.clientIDBase

	hwid, err := machineid.Get()
	if err != nil {
		log.Fatalln("[mefproxy] Не удалось получить HWID:", err)
	}

	lanIP := utils.GetLANIP()
	allIPs := utils.GetAllLANIPs()

	cfg := p.Config.Get()
	spoofVer := cfg.GameVersion
	verDisplay := "auto"
	if spoofVer != "" {
		verDisplay = spoofVer
	}

	log.Printf("[mefproxy] HWID: %s | ClientID: %d", hwid.ID, p.clientIDBase)
	log.Printf("[mefproxy] ╔══════════════════════════════════════════════════╗")
	log.Printf("[mefproxy] ║     m e f p r o x y   v 3 . 0                  ║")
	log.Printf("[mefproxy] ║  Multi-version MCPE 1.1 – 1.21.90               ║")
	log.Printf("[mefproxy] ╠══════════════════════════════════════════════════╣")
	log.Printf("[mefproxy] ║  Адрес: %-40s║", lanIP+":19132 ")
	log.Printf("[mefproxy] ║  Сервер: %-39s║", cfg.Address+" ")
	log.Printf("[mefproxy] ║  Версия (спуф): %-32s║", verDisplay+" ")
	log.Printf("[mefproxy] ╚══════════════════════════════════════════════════╝")

	if len(allIPs) > 1 {
		log.Printf("[mefproxy] Все интерфейсы:")
		for _, ip := range allIPs {
			log.Printf("[mefproxy]   • %s", ip)
		}
	}

	ip, port := utils.StringToIPPort(cfg.Address)
	transport.PingServer(ip, port, rand.Intn(65535-40000)+40000)

	for {
		p.runCycle()
		log.Println("[mefproxy] Цикл завершён, перезапуск...")
	}
}

func (p *Proxy) runCycle() {
	defer func() {
		p.setServer(nil)
		p.setClient(nil)
		p.clientRunning.Store(false)
		p.serverRunning.Store(false)
		p.bypassRP.Store(false)
		p.reconnecting.Store(false)
		p.State.reset()
		p.hooks.runDisconnect()
	}()

	p.globalEncrypt = &encrypter.Instance{IsEnabled: false}

	if p.persistentListener == nil {
		p.persistentListener = startPersistentListener(func(buf []byte) {
			p.handleClientBatch(buf)
		})
	}

	lanIP := utils.GetLANIP()
	pongMotd := "§l§5mefproxy §r§7v2.0 §a● §7| " + lanIP + ":19132"
	p.persistentListener.SetPongData(pongMotd, "1", "1")

	cc := newClientConnFromListener(p.persistentListener, func(buf []byte) {
		p.handleClientBatch(buf)
	})
	p.setClient(cc)
	p.clientRunning.Store(true)

	log.Println("[mefproxy] Клиент подключился → прокси")

	p.awaitClientOrServer()

	for p.serverRunning.Load() || p.clientRunning.Load() {
		time.Sleep(300 * time.Millisecond)
	}
}

func (p *Proxy) awaitClientOrServer() {
	cc := p.getClient()
	if cc == nil {
		return
	}
	defer func() { p.clientRunning.Store(false) }()

	select {
	case <-cc.isDisconnected:
		log.Println("[mefproxy] Клиент отключился")
		p.shutdownClientConn(cc, false)
	case <-cc.isLosted:
		log.Println("[mefproxy] Соединение с клиентом потеряно")
		p.shutdownClientConn(cc, true)
	}
}

func (p *Proxy) shutdownClientConn(cc *clientConn, lost bool) {
	p.setClient(nil)
	sc := p.getServer()
	if sc == nil {
		return
	}
	p.setServer(nil)
	go func() { sc.isShutdown <- true }()
	time.Sleep(20 * time.Millisecond)
	sc.connection.DisconnectServer()
	sc.connection.Shutdown()
	for retry := 0; retry < 6; retry++ {
		time.Sleep(500 * time.Millisecond)
		if sc.connection.IsNoRunningTasks() {
			<-connectionClosed
			sc.connection.DestroyAll()
			return
		}
	}
	sc.connection.DestroyAll()
}

func (p *Proxy) awaitServer(sc *serverConn) {
	defer func() { p.serverRunning.Store(false) }()
	select {
	case <-sc.isShutdown:
		return
	case <-sc.makeTransfer:
		p.handleMakeTransfer(sc)
	case <-sc.isLosted:
		log.Println("[mefproxy] Сервер потерян")
		p.hooks.runSrvDisconnect()
		cc := p.getClient()
		if cc != nil && p.clientRunning.Load() {
			p.SendToClient(&packet.DisconnectPacket{
				HideDisconnectionScreen: false,
				Message:                 "Сервер разорвал соединение с прокси",
			})
		}
		p.setServer(nil)
		sc.connection.Shutdown()
		p.drainServerConn(sc)
	case <-sc.isDisconnected:
		p.hooks.runSrvDisconnect()
		p.setServer(nil)
		sc.connection.Shutdown()
		p.drainServerConn(sc)
		cc := p.getClient()
		if cc != nil {
			cc.listener.DisconnectClient()
		}
	}
}

func (p *Proxy) drainServerConn(sc *serverConn) {
	for {
		time.Sleep(10 * time.Millisecond)
		if sc.connection.IsNoRunningTasks() {
			<-connectionClosed
			sc.connection.DestroyAll()
			return
		}
	}
}

func (p *Proxy) handleMakeTransfer(sc *serverConn) {
	cc := p.getClient()
	if cc == nil {
		return
	}
	cfg := p.Config.Get()
	log.Println("[mefproxy] Трансфер на:", cfg.Address)
	p.SendToClient(&packet.TransferPacket{
		Address: strings.Split(p.localAddress.Load(), ":")[0],
		Port:    19132,
	})
	p.setServer(nil)
	sc.connection.DisconnectServer()
	sc.connection.Shutdown()
	p.drainServerConn(sc)
}

func (p *Proxy) makeTransfer() {
	sc := p.getServer()
	if sc == nil || p.getClient() == nil || len(sc.makeTransfer) > 0 {
		return
	}
	go func() { sc.makeTransfer <- true }()
}

func (p *Proxy) startServerConn(loginPkt *packet.LoginPacket) {
	p.currentBindPort.Store(int32(rand.Intn(65535-40000) + 40000))
	log.Println("[mefproxy] Подключение к серверу...")

	cfg := p.Config.Get()
	ip, port := utils.StringToIPPort(cfg.Address)
	sc, err := newServerConn(ip, port, int(p.currentBindPort.Load()))
	if err != nil {
		p.SendToClient(&packet.DisconnectPacket{
			Message: "Сервер не ответил. Проверьте адрес в config.json",
		})
		log.Println("[mefproxy] Ошибка подключения:", err)
		sc.connection.Shutdown()
		go p.drainServerConn(sc)
		return
	}

	log.Println("[mefproxy] Подключён к серверу")
	p.serverRunning.Store(true)
	p.setServer(sc)

	sc.setPayloadCallback(func(buf []byte) {
		p.handleServerBatch(buf)
	})

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(loginPkt.ID())
	loginPkt.Marshal(packet.NewWriter(buf, 2))
	sc.sendPacketsRaw(map[int][]byte{0: buf.Bytes()})

	go p.awaitServer(sc)
}

func (p *Proxy) handleClientBatch(raw []byte) {
	cc := p.getClient()
	if cc == nil {
		return
	}
	pks, err := cc.decodeBatch(raw[1:])
	if err != nil {
		return
	}

	for pos, pk := range pks {
		if len(pk) == 0 {
			continue
		}

		if pk[0] == packet.IDPlayerMovePacket {
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if mv, ok := decoded.(*packet.PlayerMovePacket); ok {
					p.State.SetMyLocation(vec3.Vector3{X: mv.X, Y: mv.Y, Z: mv.Z})
				}
			}
		}

		if pk[0] == packet.IDTextPacket {
			if decoded, ok := packet.ParseDataPacket(pk).TryDecodePacket(); ok == nil {
				if tp, ok := decoded.(*packet.TextPacket); ok {
					if strings.HasPrefix(tp.Message, ".") {
						if p.hooks.runClientPacket(pk) {
							delete(pks, pos)
						}
						continue
					}
				}
			}
		}

		if pk[0] == packet.IDResourcePackClientResponsePacket && p.bypassRP.Load() {
			delete(pks, pos)
			continue
		}

		if p.hooks.runClientPacket(pk) {
			delete(pks, pos)
			continue
		}

		if pk[0] == packet.IDLoginPacket {
			decoded, decErr := packet.ParseDataPacket(pk).TryDecodePacket()
			if decErr != nil {
				continue
			}
			if dp, ok := decoded.(*packet.LoginPacket); ok {
				p.processLogin(dp, pos, pks)
				continue
			}
		}
	}

	sc := p.getServer()
	if len(pks) > 0 && sc != nil && !p.reconnecting.Load() {
		sc.sendPacketsRaw(pks)
	}
}

func (p *Proxy) processLogin(lp *packet.LoginPacket, pos int, pks map[int][]byte) {
	identity, client, _, lerr := login.Parse(lp.ConnectionRequest)
	if client.SkinData == "" {
		log.Println("[mefproxy] Пустой скин:", lerr)
		p.SendToClient(&packet.DisconnectPacket{
			Message: "Ошибка входа: пустые данные скина. Очисти данные MCPE.",
		})
		return
	}

	p.localAddress.Store(client.ServerAddress)
	cfg := p.Config.Get()

	if cfg.DeviceOS != 0 {
		client.DeviceOS = cfg.DeviceOS
	}
	if cfg.UUID != "" {
		identity.Identity = cfg.UUID
	}
	client.ClientRandomID = cfg.CurrentClientID
	if cfg.DeviceModel != "" {
		client.DeviceModel = cfg.DeviceModel
	}
	client.ServerAddress = cfg.Address
	if cfg.InputMode != 0 {
		client.CurrentInputMode = cfg.InputMode
	}
	if cfg.DInputMode != 0 {
		client.DefaultInputMode = cfg.DInputMode
	}
	client.TenantId = ""
	if cfg.Nick != "" {
		identity.DisplayName = cfg.Nick
	}
	client.LanguageCode = "ru_RU"
	client.GuiScale = 0
	if cfg.SkinData != "" {
		client.SkinData = cfg.SkinData
	}
	if cfg.SkinID != "" {
		client.SkinID = cfg.SkinID
	}
	client.ADRole = 2
	if cfg.UIProfile != 0 {
		client.UIProfile = cfg.UIProfile
	}

	if cfg.GameVersion != "" {
		client.GameVersion = cfg.GameVersion
		log.Printf("[mefproxy] Спуф версии активен: %s (оригинал: %s)", cfg.GameVersion, client.GameVersion)
	}

	encoded, key := login.GetLoginEncodedBytes(identity, client)
	p.globalEncrypt.BotKeyPair = key
	lp.ConnectionRequest = encoded

	delete(pks, pos)

	var rawLogin bytes.Buffer
	rawLogin.WriteByte(lp.ID())
	lp.Marshal(packet.NewWriter(&rawLogin, 2))
	p.loginMu.Lock()
	p.storedLoginPkt = make([]byte, rawLogin.Len())
	copy(p.storedLoginPkt, rawLogin.Bytes())
	p.loginMu.Unlock()

	go p.startServerConn(lp)
}

func (p *Proxy) handleServerBatch(raw []byte) {
	sc := p.getServer()
	if sc == nil {
		return
	}
	pks := sc.decodeBatch(raw[1:])
	if pks == nil {
		return
	}

	cfg := p.Config.Get()

	for pos, pk := range pks {
		if len(pk) == 0 {
			continue
		}

		if p.hooks.runServerPacket(pk) {
			delete(pks, pos)
			continue
		}

		if pk[0] == packet.IDResourcePacksInfoPacket && cfg.BypassRPDownload {
			p.bypassRP.Store(true)
			cc := p.getClient()
			if cc != nil {
				cc.sendPacketsRaw(map[int][]byte{0: pk})
			}
			go func() {
				time.Sleep(50 * time.Millisecond)
				p.sendRawRPResponse(packet.ResourcePackClientResponsePacket_STATUS_HAVE_ALL_PACKS)
			}()
			delete(pks, pos)
			continue
		}

		if pk[0] == packet.IDResourcePackStackPacket && cfg.BypassRPDownload && p.bypassRP.Load() {
			cc := p.getClient()
			if cc != nil {
				cc.sendPacketsRaw(map[int][]byte{0: pk})
			}
			go func() {
				time.Sleep(50 * time.Millisecond)
				p.sendRawRPResponse(packet.ResourcePackClientResponsePacket_STATUS_COMPLETED)
			}()
			delete(pks, pos)
			continue
		}

		if p.bypassRP.Load() && (pk[0] == packet.IDResourcePackDataInfoPacket || pk[0] == packet.IDResourcePackChunkDataPacket) {
			delete(pks, pos)
			continue
		}

		switch pk[0] {
		case packet.IDStartGamePacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if sgp, ok := decoded.(*packet.StartGamePacket); ok {
					p.State.SetMyEID(sgp.EntityRuntimeID)
					p.bypassRP.Store(false)
					p.reconnecting.Store(false)
					p.State.SetMyLocation(vec3.Vector3{X: sgp.X, Y: sgp.Y, Z: sgp.Z})
					p.hooks.runConnect()
					log.Printf("[mefproxy] StartGame EID=%d pos=(%.1f,%.1f,%.1f)", sgp.EntityRuntimeID, sgp.X, sgp.Y, sgp.Z)
					go func() {
						time.Sleep(500 * time.Millisecond)
						p.Notify("§0§l╔══════════════════════════════════╗")
						p.Notify("§0§l║  §5§l✦ §d§lmefproxy §fv2.0 §a§l● ONLINE§r §0§l      ║")
						p.Notify("§0§l╠══════════════════════════════════╣")
						time.Sleep(80 * time.Millisecond)
						cfg2 := p.Config.Get()
						p.Notify("§0§l║ §r §7Сервер §8▸ §f" + cfg2.Address)
						time.Sleep(80 * time.Millisecond)
						verStr := cfg2.GameVersion
						if verStr == "" {
							verStr = "§8авто"
						} else {
							verStr = "§e" + verStr
						}
						p.Notify("§0§l║ §r §7Версия §8▸ " + verStr)
						time.Sleep(80 * time.Millisecond)
						p.Notify("§0§l╠══════════════════════════════════╣")
						p.Notify("§0§l║ §r §8.help §7│ §8.modules §7│ §8.version")
						p.Notify("§0§l╚══════════════════════════════════╝")
					}()
				}
			}

		case packet.IDServerToClientHandshakePacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if hsp, ok := decoded.(*packet.ServerToClientHandshakePacket); ok {
					p.globalEncrypt.CreateSession(hsp.PublicKey, hsp.ServerToken)
					p.sendHandshakeEncrypted(sc)
					delete(pks, pos)
					continue
				}
			}

		case packet.IDDisconnectPacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if dp, ok := decoded.(*packet.DisconnectPacket); ok {
					log.Println("[mefproxy] Сервер отключил:", dp.Message)
				}
			}

		case packet.IDAddPlayerPacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if app, ok := decoded.(*packet.AddPlayerPacket); ok && app.Nick != "" && app.EntityRuntimeID != p.State.MyEID() {
					p.State.AddPlayer(&Player{
						Nick:            app.Nick,
						EntityRuntimeID: app.EntityRuntimeID,
						EntityUniqueID:  app.EntityUniqueID,
						UUID:            app.UUID,
						Location:        vec3.Vector3{X: app.Position.X(), Y: app.Position.Y(), Z: app.Position.Z()},
					})
				}
			}

		case packet.IDPlayerListPacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if plp, ok := decoded.(*packet.PlayerListPacket); ok {
					if plp.Type == packet.PlayerListPacket_TYPE_ADD {
						for uid, ent := range plp.Entries {
							p.State.CacheSkin(uid, ent[3].(string))
						}
					} else {
						for uid := range plp.Entries {
							p.State.RemovePlayerByUUID(uid)
						}
					}
				}
			}

		case packet.IDRemoveEntityPacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if rep, ok := decoded.(*packet.RemoveEntityPacket); ok {
					p.State.RemovePlayer(rep.EntityRuntimeID)
				}
			}

		case packet.IDEntityMovePacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if emp, ok := decoded.(*packet.EntityMovePacket); ok && emp.EntityRuntimeID != p.State.MyEID() {
					p.State.UpdatePlayerLocation(emp.EntityRuntimeID, vec3.Vector3{
						X: emp.X, Y: emp.Y, Z: emp.Z,
						Pitch: emp.Pitch, Yaw: emp.Yaw, HeadYaw: emp.HeadYaw,
					})
				}
			}

		case packet.IDTransferPacket:
			if decoded, err := packet.ParseDataPacket(pk).TryDecodePacket(); err == nil {
				if tp, ok := decoded.(*packet.TransferPacket); ok {
					newAddr := tp.Address + ":" + strconv.Itoa(int(tp.Port))
					if tp.Address == "" || tp.Address == "127.0.0.1" || tp.Address == "0.0.0.0" {
						cfg2 := p.Config.Get()
						newAddr = strings.Split(cfg2.Address, ":")[0] + ":" + strconv.Itoa(int(tp.Port))
					}
					log.Println("[mefproxy] Трансфер на:", newAddr)
					p.Config.SetAddress(newAddr)
					p.State.resetPlayers()
					lanIP := utils.GetLANIP()
					cc := p.getClient()
					if cc != nil {
						p.SendToClient(&packet.TransferPacket{
							Address: lanIP,
							Port:    19132,
						})
					}
					delete(pks, pos)
					continue
				}
			}
		}
	}

	cc := p.getClient()
	if len(pks) > 0 && cc != nil {
		cc.sendPacketsRaw(pks)
	}
}

func (p *Proxy) sendHandshakeEncrypted(sc *serverConn) {
	var hsBuf bytes.Buffer
	hsBuf.WriteByte(packet.IDClientToServerHandshakePacket)
	hspk := &packet.ClientToServerHandshakePacket{}
	hspk.Marshal(packet.NewWriter(&hsBuf, 2))
	hsRaw := hsBuf.Bytes()

	var b bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&b, 7)
	l := make([]byte, 5)
	_ = utils.WriteVarUInt32(zw, uint32(len(hsRaw)), l)
	_, _ = zw.Write(hsRaw)
	_ = zw.Close()

	var bb bytes.Buffer
	bb.WriteByte(0xfe)
	bb.Write(p.globalEncrypt.EncodePayloadOnce(b.Bytes()))

	p.globalEncrypt.IsEnabled = true
	sc.encrypt = p.globalEncrypt
	sc.sendPayloadDirect(bb.Bytes())
}

func (p *Proxy) sendRawRPResponse(status byte) {
	sc := p.getServer()
	if sc == nil {
		return
	}
	raw := []byte{
		packet.IDResourcePackClientResponsePacket,
		status,
		0x00, 0x00,
	}
	sc.sendPayloadDirect(sc.encodeBatch(map[int][]byte{0: raw}))
}

func (p *Proxy) getClient() *clientConn {
	p.clientMu.RLock()
	defer p.clientMu.RUnlock()
	return p.client
}

func (p *Proxy) setClient(cc *clientConn) {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()
	p.client = cc
}

func (p *Proxy) getServer() *serverConn {
	p.serverMu.RLock()
	defer p.serverMu.RUnlock()
	return p.server
}

func (p *Proxy) setServer(sc *serverConn) {
	p.serverMu.Lock()
	defer p.serverMu.Unlock()
	p.server = sc
}

func sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
