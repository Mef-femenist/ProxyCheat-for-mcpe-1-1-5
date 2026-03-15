
package transport

import (
	"bytes"
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var rakMagic = [16]byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

const (
	idUnconnPing   = 0x01
	idUnconnPong   = 0x1c
	idOpenReq1     = 0x05
	idOpenReply1   = 0x06
	idOpenReq2     = 0x07
	idOpenReply2   = 0x08
	idIncompat     = 0x19
	idConnPing     = 0x00
	idConnPong     = 0x03
	idConnReq      = 0x09
	idConnAccepted = 0x10
	idNewIncoming  = 0x13
	idDisconnect   = 0x15
	flagACK        = 0xC0
	flagNAK        = 0xA0
	flagDatagram   = 0x80
	relReliableOrd = 3
)

func magic() []byte { return rakMagic[:] }

func hasMagic(b []byte) bool {
	if len(b) < 16 {
		return false
	}
	for i := range rakMagic {
		if b[i] != rakMagic[i] {
			return false
		}
	}
	return true
}

func u24le(b []byte) uint32  { return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 }
func pu24le(v uint32) []byte { return []byte{byte(v), byte(v >> 8), byte(v >> 16)} }

func writeAddrIPv4(buf *bytes.Buffer, ip net.IP, port int) {
	ip4 := ip.To4()
	if ip4 == nil {
		ip4 = []byte{127, 0, 0, 1}
	}
	buf.WriteByte(4)
	buf.Write([]byte{^ip4[0], ^ip4[1], ^ip4[2], ^ip4[3]})
	binary.Write(buf, binary.BigEndian, uint16(port))
}

func writeNullAddr(buf *bytes.Buffer) {
	buf.WriteByte(4)
	buf.Write([]byte{0xff, 0xff, 0xff, 0xff})
	buf.Write([]byte{0, 0})
}

func readAddrIPv4(data []byte) (net.IP, int, int) {
	if len(data) < 7 {
		return net.IP{127, 0, 0, 1}, 19132, 0
	}
	if data[0] == 4 {
		ip := net.IP{^data[1], ^data[2], ^data[3], ^data[4]}
		port := int(binary.BigEndian.Uint16(data[5:7]))
		return ip, port, 7
	}
	if len(data) < 29 {
		return net.IP{127, 0, 0, 1}, 19132, 0
	}
	return net.IP{127, 0, 0, 1}, 19132, 29
}

type encapPkt struct {
	reliability  int
	msgIdx       uint32
	orderIdx     uint32
	orderChannel byte
	hasSplit     bool
	splitCount   uint32
	splitID      uint16
	splitIndex   uint32
	payload      []byte
}

func parseDatagram(data []byte) (seqNum uint32, pkts []encapPkt) {
	if len(data) < 4 {
		return
	}
	seqNum = u24le(data[1:4])
	data = data[4:]
	for len(data) >= 3 {
		flags := data[0]
		rel := int((flags >> 5) & 0x07)
		split := flags&0x10 != 0
		lenBits := binary.BigEndian.Uint16(data[1:3])
		lenBytes := int((int(lenBits) + 7) / 8)
		pos := 3
		var pkt encapPkt
		pkt.reliability = rel
		pkt.hasSplit = split
		if rel == 2 || rel == 3 || rel == 4 || rel == 6 || rel == 7 {
			if pos+3 > len(data) {
				return
			}
			pkt.msgIdx = u24le(data[pos:])
			pos += 3
		}
		if rel == 1 || rel == 4 {
			if pos+3 > len(data) {
				return
			}
			pos += 3
		}
		if rel == 3 || rel == 4 || rel == 7 {
			if pos+4 > len(data) {
				return
			}
			pkt.orderIdx = u24le(data[pos:])
			pos += 3
			pkt.orderChannel = data[pos]
			pos++
		}
		if split {
			if pos+10 > len(data) {
				return
			}
			pkt.splitCount = binary.BigEndian.Uint32(data[pos:])
			pos += 4
			pkt.splitID = binary.BigEndian.Uint16(data[pos:])
			pos += 2
			pkt.splitIndex = binary.BigEndian.Uint32(data[pos:])
			pos += 4
		}
		if pos+lenBytes > len(data) {
			break
		}
		pkt.payload = make([]byte, lenBytes)
		copy(pkt.payload, data[pos:pos+lenBytes])
		data = data[pos+lenBytes:]
		pkts = append(pkts, pkt)
	}
	return
}

func buildDatagram(seqNum, msgIdx, orderIdx uint32, payload []byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteByte(0x84)
	buf.Write(pu24le(seqNum))
	buf.WriteByte(byte(relReliableOrd << 5))
	binary.Write(buf, binary.BigEndian, uint16(len(payload)*8))
	buf.Write(pu24le(msgIdx))
	buf.Write(pu24le(orderIdx))
	buf.WriteByte(0)
	buf.Write(payload)
	return buf.Bytes()
}

func buildACK(seqNums []uint32) []byte {
	if len(seqNums) == 0 {
		return nil
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(flagACK)
	binary.Write(buf, binary.BigEndian, uint16(len(seqNums)))
	for _, sn := range seqNums {
		buf.WriteByte(1)
		buf.Write(pu24le(sn))
	}
	return buf.Bytes()
}

type splitBuf struct {
	count  uint32
	pieces map[uint32][]byte
}

func (s *splitBuf) add(index uint32, data []byte) []byte {
	s.pieces[index] = data
	if uint32(len(s.pieces)) < s.count {
		return nil
	}
	var out []byte
	for i := uint32(0); i < s.count; i++ {
		out = append(out, s.pieces[i]...)
	}
	return out
}

type peerState int32

const (
	stateNone      peerState = 0
	stateHandshake peerState = 1
	stateConnected peerState = 2
	stateClosing   peerState = 3
)

type Listener struct {
	conn       *net.UDPConn
	serverGUID int64
	pongStr    string
	version    int

	clientAddr *net.UDPAddr
	clientGUID int64
	clientMTU  int16
	state      int32

	sendMu     sync.Mutex
	sendSeq    uint32
	sendMsgIdx uint32
	sendOrdIdx uint32

	ackMu    sync.Mutex
	pendACKs []uint32

	splitsMu sync.Mutex
	splits   map[uint16]*splitBuf

	cbPayload  func([]byte)
	cbDisconn  func()
	cbLost     func()
	cbConn     func()
	cbShutdown func()

	running  int32
	allDone  int32
	stopCh   chan struct{}
	lastRecv time.Time
	lastMu   sync.Mutex
}

func NewListener(onConnected func(), onShutdown func(), _ bool) *Listener {
	l := &Listener{
		serverGUID: rand.Int63(),
		version:    0,
		splits:     make(map[uint16]*splitBuf),
		stopCh:     make(chan struct{}),
		cbConn:     onConnected,
		cbShutdown: onShutdown,
	}
	l.pongStr = "MCPE;Proxy;110;1.1.5;0;1;" + string(rune(l.serverGUID)) + ";Proxy;Survival;1;19132;19133;"
	return l
}

func (l *Listener) SetPongData(motd, onl, maxl string) {
	l.pongStr = "MCPE;" + motd + ";110;1.1.5;" + onl + ";" + maxl + ";" + "12345678" + ";Proxy;Survival;1;19132;19133;"
}

func (l *Listener) SetVersion(ver int) { l.version = ver }

func (l *Listener) SetClientPayloadReceiveCallback(fn func([]byte)) { l.cbPayload = fn }
func (l *Listener) SetClientDisconnectCallback(fn func())           { l.cbDisconn = fn }
func (l *Listener) SetClientLostConnectionCallback(fn func())       { l.cbLost = fn }

func (l *Listener) StartListen() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 19132})
	if err != nil {
		log.Println("Listener: bind error:", err)
		return
	}
	l.conn = conn
	atomic.StoreInt32(&l.running, 1)
	atomic.StoreInt32(&l.allDone, 0)

	go l.ackSender()
	go l.timeoutMonitor()
	go l.keepAliveSender() 
	l.readLoop()

	atomic.StoreInt32(&l.allDone, 1)
	if l.cbShutdown != nil {
		l.cbShutdown()
	}
}

func (l *Listener) readLoop() {
	buf := make([]byte, 4096) 
	for atomic.LoadInt32(&l.running) == 1 {
		l.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, addr, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		l.handlePacket(data, addr)
	}
}

func (l *Listener) handlePacket(data []byte, addr *net.UDPAddr) {
	if len(data) == 0 {
		return
	}
	id := data[0]

	if id == idUnconnPing && hasMagic(data[9:]) {
		l.sendPong(data, addr)
		return
	}

	if id == idOpenReq1 {
		l.handleOpenReq1(data, addr)
		return
	}

	if id == idOpenReq2 {
		l.handleOpenReq2(data, addr)
		return
	}

	if l.clientAddr == nil || addr.String() != l.clientAddr.String() {
		return
	}

	l.lastMu.Lock()
	l.lastRecv = time.Now()
	l.lastMu.Unlock()

	if id == flagACK {
		return
	}
	if id == flagNAK {
		return
	}

	if id&flagDatagram != 0 {
		l.handleDatagram(data, addr)
		return
	}
}

func (l *Listener) sendPong(ping []byte, addr *net.UDPAddr) {
	var pingTime int64
	if len(ping) >= 9 {
		pingTime = int64(binary.BigEndian.Uint64(ping[1:9]))
	}
	motdData := []byte(l.pongStr)
	buf := &bytes.Buffer{}
	buf.WriteByte(idUnconnPong)
	binary.Write(buf, binary.BigEndian, pingTime)
	binary.Write(buf, binary.BigEndian, l.serverGUID)
	buf.Write(magic())
	binary.Write(buf, binary.BigEndian, uint16(len(motdData)))
	buf.Write(motdData)
	l.conn.WriteToUDP(buf.Bytes(), addr)
}

func (l *Listener) handleOpenReq1(data []byte, addr *net.UDPAddr) {
	if len(data) < 19 {
		return
	}
	if !hasMagic(data[1:]) {
		return
	}
	proto := data[17]
	mtu := int16(len(data) + 28)

	if l.version != 0 && int(proto) != l.version {
		buf := &bytes.Buffer{}
		buf.WriteByte(idIncompat)
		buf.Write(magic())
		binary.Write(buf, binary.BigEndian, l.serverGUID)
		l.conn.WriteToUDP(buf.Bytes(), addr)
		return
	}

	buf := &bytes.Buffer{}
	buf.WriteByte(idOpenReply1)
	buf.Write(magic())
	binary.Write(buf, binary.BigEndian, l.serverGUID)
	buf.WriteByte(0)
	binary.Write(buf, binary.BigEndian, mtu)
	l.conn.WriteToUDP(buf.Bytes(), addr)
	l.clientMTU = mtu
}

func (l *Listener) handleOpenReq2(data []byte, addr *net.UDPAddr) {
	if len(data) < 26 {
		return
	}
	if !hasMagic(data[1:]) {
		return
	}
	pos := 17
	_, _, addrLen := readAddrIPv4(data[pos:])
	pos += addrLen
	if pos+10 > len(data) {
		return
	}
	mtu := int16(binary.BigEndian.Uint16(data[pos:]))
	pos += 2
	guid := int64(binary.BigEndian.Uint64(data[pos:]))

	l.clientAddr = addr
	l.clientGUID = guid
	l.clientMTU = mtu
	atomic.StoreInt32(&l.state, int32(stateHandshake))

	buf := &bytes.Buffer{}
	buf.WriteByte(idOpenReply2)
	buf.Write(magic())
	binary.Write(buf, binary.BigEndian, l.serverGUID)
	writeAddrIPv4(buf, addr.IP, addr.Port)
	binary.Write(buf, binary.BigEndian, mtu)
	buf.WriteByte(0)
	l.conn.WriteToUDP(buf.Bytes(), addr)
}

func (l *Listener) handleDatagram(data []byte, addr *net.UDPAddr) {
	seqNum, pkts := parseDatagram(data)
	l.ackMu.Lock()
	l.pendACKs = append(l.pendACKs, seqNum)
	l.ackMu.Unlock()

	for _, pkt := range pkts {
		payload := pkt.payload
		if pkt.hasSplit {
			l.splitsMu.Lock()
			sb, ok := l.splits[pkt.splitID]
			if !ok {
				sb = &splitBuf{count: pkt.splitCount, pieces: make(map[uint32][]byte)}
				l.splits[pkt.splitID] = sb
			}
			assembled := sb.add(pkt.splitIndex, pkt.payload)
			if assembled != nil {
				delete(l.splits, pkt.splitID)
				l.splitsMu.Unlock()
				payload = assembled
			} else {
				l.splitsMu.Unlock()
				continue
			}
		}
		if len(payload) == 0 {
			continue
		}
		l.handleInnerPacket(payload, addr)
	}
}

func (l *Listener) handleInnerPacket(payload []byte, addr *net.UDPAddr) {
	if len(payload) == 0 {
		return
	}
	id := payload[0]

	switch id {
	case idConnPing:
		if len(payload) < 9 {
			return
		}
		t := int64(binary.BigEndian.Uint64(payload[1:9]))
		l.sendConnPong(t)

	case idConnReq:
		l.sendConnAccepted(payload, addr)

	case idNewIncoming:
		if atomic.LoadInt32(&l.state) < int32(stateConnected) {
			atomic.StoreInt32(&l.state, int32(stateConnected))
			l.lastMu.Lock()
			l.lastRecv = time.Now()
			l.lastMu.Unlock()
			log.Println("RakNet Listener: client connected from", addr)
			if l.cbConn != nil {
				go l.cbConn()
			}
		}

	case idDisconnect:
		if atomic.LoadInt32(&l.state) == int32(stateConnected) {
			atomic.StoreInt32(&l.state, int32(stateClosing))
			if l.cbDisconn != nil {
				go l.cbDisconn()
			}
		}

	case 0xfe:
		if atomic.LoadInt32(&l.state) == int32(stateConnected) && l.cbPayload != nil {
			cp := make([]byte, len(payload))
			copy(cp, payload)
			go l.cbPayload(cp)
		}
	}
}

func (l *Listener) sendConnPong(t int64) {
	buf := &bytes.Buffer{}
	buf.WriteByte(idConnPong)
	binary.Write(buf, binary.BigEndian, t)
	binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
	l.sendReliableOrdered(buf.Bytes())
}

func (l *Listener) sendConnAccepted(req []byte, addr *net.UDPAddr) {
	var clientTime int64
	if len(req) >= 9 {
		clientTime = int64(binary.BigEndian.Uint64(req[1:9]))
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(idConnAccepted)
	writeAddrIPv4(buf, addr.IP, addr.Port)
	binary.Write(buf, binary.BigEndian, uint16(0))
	for i := 0; i < 10; i++ {
		writeNullAddr(buf)
	}
	binary.Write(buf, binary.BigEndian, clientTime)
	binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
	l.sendReliableOrdered(buf.Bytes())
}

func (l *Listener) sendReliableOrdered(payload []byte) {
	if l.conn == nil || l.clientAddr == nil {
		return
	}
	l.sendMu.Lock()
	seq := l.sendSeq
	msgIdx := l.sendMsgIdx
	ordIdx := l.sendOrdIdx
	l.sendSeq++
	l.sendMsgIdx++
	l.sendOrdIdx++
	l.sendMu.Unlock()
	dgram := buildDatagram(seq, msgIdx, ordIdx, payload)
	l.conn.WriteToUDP(dgram, l.clientAddr)
}

func (l *Listener) SendPayload(payload []byte, _ int) {
	if l.conn == nil || l.clientAddr == nil {
		return
	}
	mtu := int(l.clientMTU)
	if mtu <= 0 {
		mtu = 1400
	}
	maxPayload := mtu - 60
	if maxPayload < 100 {
		maxPayload = 100
	}

	if len(payload) <= maxPayload {
		l.sendReliableOrdered(payload)
		return
	}

	splitID := uint16(rand.Uint32())
	splitCount := uint32((len(payload) + maxPayload - 1) / maxPayload)
	l.sendMu.Lock()
	baseSeq := l.sendSeq
	baseMsgIdx := l.sendMsgIdx
	baseOrdIdx := l.sendOrdIdx
	l.sendSeq += splitCount
	l.sendMsgIdx += splitCount
	l.sendOrdIdx++
	l.sendMu.Unlock()

	for i := uint32(0); i < splitCount; i++ {
		start := int(i) * maxPayload
		end := start + maxPayload
		if end > len(payload) {
			end = len(payload)
		}
		chunk := payload[start:end]

		buf := &bytes.Buffer{}
		buf.WriteByte(0x84)
		buf.Write(pu24le(baseSeq + i))
		flags := byte(relReliableOrd<<5) | 0x10
		buf.WriteByte(flags)
		binary.Write(buf, binary.BigEndian, uint16(len(chunk)*8))
		buf.Write(pu24le(baseMsgIdx + i))
		buf.Write(pu24le(baseOrdIdx))
		buf.WriteByte(0)
		binary.Write(buf, binary.BigEndian, splitCount)
		binary.Write(buf, binary.BigEndian, splitID)
		binary.Write(buf, binary.BigEndian, i)
		buf.Write(chunk)
		l.conn.WriteToUDP(buf.Bytes(), l.clientAddr)
	}
}

func (l *Listener) DisconnectClient() {
	if l.conn == nil || l.clientAddr == nil {
		return
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(idDisconnect)
	l.sendReliableOrdered(buf.Bytes())
	time.Sleep(100 * time.Millisecond)
}

func (l *Listener) ReceivePayload() []byte { return nil }

func (l *Listener) Shutdown() {
	atomic.StoreInt32(&l.running, 0)
	if l.conn != nil {
		l.conn.Close()
	}
}

func (l *Listener) IsNoRunningTasks() bool {
	return atomic.LoadInt32(&l.allDone) == 1
}

func (l *Listener) DestroyAll() {}

func (l *Listener) Reset() {
	l.sendMu.Lock()
	l.sendSeq = 0
	l.sendMsgIdx = 0
	l.sendOrdIdx = 0
	l.sendMu.Unlock()

	l.ackMu.Lock()
	l.pendACKs = nil
	l.ackMu.Unlock()

	l.splitsMu.Lock()
	l.splits = make(map[uint16]*splitBuf)
	l.splitsMu.Unlock()

	l.clientAddr = nil
	l.clientGUID = 0
	l.clientMTU = 1400
	atomic.StoreInt32(&l.state, int32(stateNone))
	l.lastMu.Lock()
	l.lastRecv = time.Now()
	l.lastMu.Unlock()
}

func (l *Listener) SetOnConnected(fn func()) {
	l.cbConn = fn
}

func (l *Listener) ResetIfNeeded() {
	st := atomic.LoadInt32(&l.state)
	if st == int32(stateNone) || st == int32(stateHandshake) {
		
		return
	}
	
	l.sendMu.Lock()
	l.sendSeq = 0
	l.sendMsgIdx = 0
	l.sendOrdIdx = 0
	l.sendMu.Unlock()
	l.ackMu.Lock()
	l.pendACKs = nil
	l.ackMu.Unlock()
	l.splitsMu.Lock()
	l.splits = make(map[uint16]*splitBuf)
	l.splitsMu.Unlock()
	l.clientAddr = nil
	l.clientGUID = 0
	l.clientMTU = 1400
	atomic.StoreInt32(&l.state, int32(stateNone))
	l.lastMu.Lock()
	l.lastRecv = time.Now()
	l.lastMu.Unlock()
}

func (l *Listener) ackSender() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for atomic.LoadInt32(&l.running) == 1 {
		<-ticker.C
		l.ackMu.Lock()
		if len(l.pendACKs) > 0 && l.clientAddr != nil {
			ack := buildACK(l.pendACKs)
			l.pendACKs = nil
			l.ackMu.Unlock()
			l.conn.WriteToUDP(ack, l.clientAddr)
		} else {
			l.ackMu.Unlock()
		}
	}
}

func (l *Listener) timeoutMonitor() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for atomic.LoadInt32(&l.running) == 1 {
		<-ticker.C
		if atomic.LoadInt32(&l.state) != int32(stateConnected) {
			continue
		}
		l.lastMu.Lock()
		since := time.Since(l.lastRecv)
		l.lastMu.Unlock()
		if since > 10*time.Second {
			log.Println("RakNet Listener: client timeout")
			atomic.StoreInt32(&l.state, int32(stateClosing))
			if l.cbLost != nil {
				go l.cbLost()
			}
			return
		}
	}
}

func (l *Listener) keepAliveSender() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for atomic.LoadInt32(&l.running) == 1 {
		<-ticker.C
		if atomic.LoadInt32(&l.state) != int32(stateConnected) {
			continue
		}
		buf := &bytes.Buffer{}
		buf.WriteByte(idConnPing)
		binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
		l.sendReliableOrdered(buf.Bytes())
	}
}

type Connector struct {
	conn       *net.UDPConn
	serverAddr *net.UDPAddr
	bindPort   int
	clientGUID int64
	serverGUID int64
	mtu        int16
	version    int

	state int32

	sendMu     sync.Mutex
	sendSeq    uint32
	sendMsgIdx uint32
	sendOrdIdx uint32

	ackMu    sync.Mutex
	pendACKs []uint32

	splitsMu sync.Mutex
	splits   map[uint16]*splitBuf

	cbPayload  func([]byte)
	cbDisconn  func()
	cbLost     func()
	cbTimeout  func()
	cbShutdown func()

	running        int32
	allDone        int32
	stopCh         chan struct{}
	lastRecv       time.Time
	lastMu         sync.Mutex
	handshakePktCh chan []byte
	innerPktCh     chan []byte
	payloadCh      chan []byte
}

func NewConnector(onShutdown func(), _ bool) *Connector {
	return &Connector{
		clientGUID:     rand.Int63(),
		version:        8,
		mtu:            1400,
		splits:         make(map[uint16]*splitBuf),
		stopCh:         make(chan struct{}),
		cbShutdown:     onShutdown,
		handshakePktCh: make(chan []byte, 8),
		innerPktCh:     make(chan []byte, 8),
		payloadCh:      make(chan []byte, 2048), 
	}
}

func (c *Connector) SetVersion(ver int)                              { c.version = ver }
func (c *Connector) SetServerPayloadReceiveCallback(fn func([]byte)) { c.cbPayload = fn }
func (c *Connector) SetServerDisconnectCallback(fn func())           { c.cbDisconn = fn }
func (c *Connector) SetServerLostConnectionCallback(fn func())       { c.cbLost = fn }
func (c *Connector) SetServerConnectionTimeoutCallback(fn func())    { c.cbTimeout = fn }

func resolveHost(host string) net.IP {
	if ip := net.ParseIP(host); ip != nil {
		if v4 := ip.To4(); v4 != nil {
			return v4
		}
		return ip
	}
	addrs, err := net.LookupHost(host)
	if err != nil || len(addrs) == 0 {
		return nil
	}
	for _, a := range addrs {
		if ip := net.ParseIP(a); ip != nil {
			if v4 := ip.To4(); v4 != nil {
				return v4
			}
		}
	}
	if ip := net.ParseIP(addrs[0]); ip != nil {
		return ip
	}
	return nil
}

func (c *Connector) SetInfo(ip string, port, localport int) {
	resolved := resolveHost(ip)
	c.serverAddr = &net.UDPAddr{IP: resolved, Port: port}
	c.bindPort = localport
}

func (c *Connector) Connect(onConnected func()) {
	if c.serverAddr == nil || c.serverAddr.IP == nil {
		log.Println("Connector: не удалось разрешить адрес сервера. Проверьте адрес в config.json")
		if c.cbTimeout != nil {
			go c.cbTimeout()
		}
		return
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: c.bindPort})
	if err != nil {
		log.Println("Connector: bind error:", err)
		if c.cbTimeout != nil {
			go c.cbTimeout()
		}
		return
	}
	c.conn = conn
	atomic.StoreInt32(&c.running, 1)
	atomic.StoreInt32(&c.allDone, 0)

	go c.ackSender()
	go c.timeoutMonitor()
	go c.payloadWorker()
	go c.readLoop()
	go c.keepAliveSender()

	if !c.doHandshake() {
		atomic.StoreInt32(&c.running, 0)
		if c.cbTimeout != nil {
			go c.cbTimeout()
		}
		return
	}

	if onConnected != nil {
		go onConnected()
	}
}

func (c *Connector) doHandshake() bool {
	mtuSizes := []int{1492, 1400, 1200, 576}
	var useMTU int16

	for _, mtu := range mtuSizes {
		if err := c.sendOpenReq1(mtu); err != nil {
			continue
		}
		reply, err := c.waitPacket(idOpenReply1, 3*time.Second)
		if err != nil {
			continue
		}
		if len(reply) < 28 || !hasMagic(reply[1:]) {
			continue
		}
		c.serverGUID = int64(binary.BigEndian.Uint64(reply[17:25]))
		useMTU = int16(binary.BigEndian.Uint16(reply[26:28]))
		break
	}
	if useMTU == 0 {
		return false
	}
	c.mtu = useMTU

	if err := c.sendOpenReq2(int(useMTU)); err != nil {
		return false
	}
	if _, err := c.waitPacket(idOpenReply2, 3*time.Second); err != nil {
		return false
	}

	c.sendConnReq()
	connAccept, err := c.waitInnerPacket(idConnAccepted, 7*time.Second)
	if err != nil {
		log.Println("Connector: no ConnectionRequestAccepted:", err)
		return false
	}
	_ = connAccept
	c.sendNewIncoming()
	atomic.StoreInt32(&c.state, int32(stateConnected))
	c.lastMu.Lock()
	c.lastRecv = time.Now()
	c.lastMu.Unlock()
	return true
}

func (c *Connector) waitPacket(id byte, timeout time.Duration) ([]byte, error) {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case pkt := <-c.handshakePktCh:
			if len(pkt) > 0 && pkt[0] == id {
				return pkt, nil
			}
		case <-deadline.C:
			return nil, net.ErrClosed
		}
	}
}

func (c *Connector) waitInnerPacket(id byte, timeout time.Duration) ([]byte, error) {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case pkt := <-c.innerPktCh:
			if len(pkt) > 0 && pkt[0] == id {
				return pkt, nil
			}
		case <-deadline.C:
			return nil, net.ErrClosed
		}
	}
}

func (c *Connector) sendOpenReq1(mtu int) error {
	buf := &bytes.Buffer{}
	buf.WriteByte(idOpenReq1)
	buf.Write(magic())
	buf.WriteByte(byte(c.version))
	padding := mtu - 1 - 16 - 1 - 28
	if padding < 0 {
		padding = 0
	}
	buf.Write(make([]byte, padding))
	_, err := c.conn.WriteToUDP(buf.Bytes(), c.serverAddr)
	return err
}

func (c *Connector) sendOpenReq2(mtu int) error {
	buf := &bytes.Buffer{}
	buf.WriteByte(idOpenReq2)
	buf.Write(magic())
	writeAddrIPv4(buf, c.serverAddr.IP, c.serverAddr.Port)
	binary.Write(buf, binary.BigEndian, int16(mtu))
	binary.Write(buf, binary.BigEndian, c.clientGUID)
	_, err := c.conn.WriteToUDP(buf.Bytes(), c.serverAddr)
	return err
}

func (c *Connector) sendConnReq() {
	buf := &bytes.Buffer{}
	buf.WriteByte(idConnReq)
	binary.Write(buf, binary.BigEndian, c.clientGUID)
	binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
	buf.WriteByte(0)
	c.sendReliableOrdered(buf.Bytes())
}

func (c *Connector) sendNewIncoming() {
	buf := &bytes.Buffer{}
	buf.WriteByte(idNewIncoming)
	writeAddrIPv4(buf, c.serverAddr.IP, c.serverAddr.Port)
	for i := 0; i < 10; i++ {
		writeNullAddr(buf)
	}
	binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
	binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
	c.sendReliableOrdered(buf.Bytes())
}

func (c *Connector) readLoop() {
	buf := make([]byte, 4096) 
	for atomic.LoadInt32(&c.running) == 1 {
		c.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		c.handlePacket(data)
	}
	atomic.StoreInt32(&c.allDone, 1)
	if c.cbShutdown != nil {
		c.cbShutdown()
	}
}

func (c *Connector) handlePacket(data []byte) {
	if len(data) == 0 {
		return
	}
	id := data[0]

	if id == idOpenReply1 || id == idOpenReply2 || id == idIncompat {
		select {
		case c.handshakePktCh <- data:
		default:
		}
		return
	}

	if id == flagACK {
		return
	}
	if id == flagNAK {
		return
	}

	if id&flagDatagram != 0 {
		c.lastMu.Lock()
		c.lastRecv = time.Now()
		c.lastMu.Unlock()
		c.handleDatagram(data)
		return
	}
}

func (c *Connector) handleDatagram(data []byte) {
	seqNum, pkts := parseDatagram(data)
	c.ackMu.Lock()
	c.pendACKs = append(c.pendACKs, seqNum)
	c.ackMu.Unlock()

	for _, pkt := range pkts {
		payload := pkt.payload
		if pkt.hasSplit {
			c.splitsMu.Lock()
			sb, ok := c.splits[pkt.splitID]
			if !ok {
				sb = &splitBuf{count: pkt.splitCount, pieces: make(map[uint32][]byte)}
				c.splits[pkt.splitID] = sb
			}
			assembled := sb.add(pkt.splitIndex, pkt.payload)
			if assembled != nil {
				delete(c.splits, pkt.splitID)
				c.splitsMu.Unlock()
				payload = assembled
			} else {
				c.splitsMu.Unlock()
				continue
			}
		}
		if len(payload) == 0 {
			continue
		}
		c.handleInnerPacket(payload)
	}
}

func (c *Connector) handleInnerPacket(payload []byte) {
	if len(payload) == 0 {
		return
	}
	id := payload[0]

	switch id {
	case idConnPing:
		if len(payload) < 9 {
			return
		}
		t := int64(binary.BigEndian.Uint64(payload[1:9]))
		buf := &bytes.Buffer{}
		buf.WriteByte(idConnPong)
		binary.Write(buf, binary.BigEndian, t)
		binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
		c.sendReliableOrdered(buf.Bytes())

	case idConnAccepted, idNewIncoming:
		select {
		case c.innerPktCh <- payload:
		default:
		}

	case idDisconnect:
		if atomic.CompareAndSwapInt32(&c.state, int32(stateConnected), int32(stateClosing)) {
			if c.cbDisconn != nil {
				go c.cbDisconn()
			}
		}

	case 0xfe:
		if atomic.LoadInt32(&c.state) == int32(stateConnected) && c.cbPayload != nil {
			cp := make([]byte, len(payload))
			copy(cp, payload)
			select {
			case c.payloadCh <- cp:
			default:
				
				<-c.payloadCh
				c.payloadCh <- cp
			}
		}
	}
}

func (c *Connector) sendReliableOrdered(payload []byte) {
	if c.conn == nil {
		return
	}
	c.sendMu.Lock()
	seq := c.sendSeq
	msgIdx := c.sendMsgIdx
	ordIdx := c.sendOrdIdx
	c.sendSeq++
	c.sendMsgIdx++
	c.sendOrdIdx++
	c.sendMu.Unlock()
	dgram := buildDatagram(seq, msgIdx, ordIdx, payload)
	c.conn.WriteToUDP(dgram, c.serverAddr)
}

func (c *Connector) SendPayload(payload []byte, _ int) {
	if c.conn == nil {
		return
	}
	mtu := int(c.mtu)
	if mtu <= 0 {
		mtu = 1400
	}
	maxPayload := mtu - 60
	if maxPayload < 100 {
		maxPayload = 100
	}

	if len(payload) <= maxPayload {
		c.sendReliableOrdered(payload)
		return
	}

	splitID := uint16(rand.Uint32())
	splitCount := uint32((len(payload) + maxPayload - 1) / maxPayload)
	c.sendMu.Lock()
	baseSeq := c.sendSeq
	baseMsgIdx := c.sendMsgIdx
	baseOrdIdx := c.sendOrdIdx
	c.sendSeq += splitCount
	c.sendMsgIdx += splitCount
	c.sendOrdIdx++
	c.sendMu.Unlock()

	for i := uint32(0); i < splitCount; i++ {
		start := int(i) * maxPayload
		end := start + maxPayload
		if end > len(payload) {
			end = len(payload)
		}
		chunk := payload[start:end]

		buf := &bytes.Buffer{}
		buf.WriteByte(0x84)
		buf.Write(pu24le(baseSeq + i))
		flags := byte(relReliableOrd<<5) | 0x10
		buf.WriteByte(flags)
		binary.Write(buf, binary.BigEndian, uint16(len(chunk)*8))
		buf.Write(pu24le(baseMsgIdx + i))
		buf.Write(pu24le(baseOrdIdx))
		buf.WriteByte(0)
		binary.Write(buf, binary.BigEndian, splitCount)
		binary.Write(buf, binary.BigEndian, splitID)
		binary.Write(buf, binary.BigEndian, i)
		buf.Write(chunk)
		c.conn.WriteToUDP(buf.Bytes(), c.serverAddr)
	}
}

func (c *Connector) DisconnectServer() {
	if c.conn == nil {
		return
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(idDisconnect)
	c.sendReliableOrdered(buf.Bytes())
	time.Sleep(100 * time.Millisecond)
}

func (c *Connector) Shutdown() {
	atomic.StoreInt32(&c.running, 0)
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Connector) IsNoRunningTasks() bool { return atomic.LoadInt32(&c.allDone) == 1 }
func (c *Connector) DestroyAll()            {}

func (c *Connector) payloadWorker() {
	for {
		select {
		case data, ok := <-c.payloadCh:
			if !ok {
				return
			}
			if c.cbPayload != nil {
				c.cbPayload(data)
			}
		}
		if atomic.LoadInt32(&c.running) == 0 && len(c.payloadCh) == 0 {
			return
		}
	}
}

func (c *Connector) ackSender() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for atomic.LoadInt32(&c.running) == 1 {
		<-ticker.C
		c.ackMu.Lock()
		if len(c.pendACKs) > 0 {
			ack := buildACK(c.pendACKs)
			c.pendACKs = nil
			c.ackMu.Unlock()
			c.conn.WriteToUDP(ack, c.serverAddr)
		} else {
			c.ackMu.Unlock()
		}
	}
}

func (c *Connector) keepAliveSender() {
	ticker := time.NewTicker(8 * time.Second)
	defer ticker.Stop()
	for atomic.LoadInt32(&c.running) == 1 {
		<-ticker.C
		if atomic.LoadInt32(&c.state) != int32(stateConnected) {
			continue
		}
		buf := &bytes.Buffer{}
		buf.WriteByte(idConnPing)
		binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
		c.sendReliableOrdered(buf.Bytes())
	}
}

func (c *Connector) timeoutMonitor() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for atomic.LoadInt32(&c.running) == 1 {
		<-ticker.C
		if atomic.LoadInt32(&c.state) != int32(stateConnected) {
			continue
		}
		c.lastMu.Lock()
		since := time.Since(c.lastRecv)
		c.lastMu.Unlock()
		if since > 10*time.Second {
			log.Println("Connector: server timeout")
			if atomic.CompareAndSwapInt32(&c.state, int32(stateConnected), int32(stateClosing)) {
				if c.cbLost != nil {
					go c.cbLost()
				}
			}
			return
		}
	}
}

func PingServer(ip string, port, lport int) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: lport})
	if err != nil {
		return
	}
	defer conn.Close()
	resolved := resolveHost(ip)
	if resolved == nil {
		log.Println("PingServer: не удалось разрешить адрес:", ip)
		return
	}
	addr := &net.UDPAddr{IP: resolved, Port: port}
	buf := &bytes.Buffer{}
	buf.WriteByte(idUnconnPing)
	binary.Write(buf, binary.BigEndian, time.Now().UnixMilli())
	buf.Write(magic())
	binary.Write(buf, binary.BigEndian, rand.Int63())
	conn.WriteToUDP(buf.Bytes(), addr)

	tmp := make([]byte, 2048)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, _, err := conn.ReadFromUDP(tmp)
	if err != nil {
		return
	}
	pong := tmp[:n]
	if len(pong) < 35 || pong[0] != idUnconnPong {
		return
	}
	dataLen := int(binary.BigEndian.Uint16(pong[33:35]))
	if 35+dataLen > len(pong) {
		return
	}
	motd := string(pong[35 : 35+dataLen])
	log.Println("Сервер отвечает:", motd)
}
