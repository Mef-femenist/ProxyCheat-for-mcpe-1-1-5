package proxy

import (
	"bytes"
	"compress/zlib"
	"io"
	"sort"

	"mefproxy/pkg/transport"
	"mefproxy/utils"
)

type clientConn struct {
	listener       *transport.Listener
	isShutdown     chan bool
	isLosted       chan bool
	isDisconnected chan bool
}

func startPersistentListener(onPayload func([]byte)) *transport.Listener {
	connectedCh := make(chan bool, 1)
	list := transport.NewListener(
		func() { connectedCh <- true },
		func() {},
		false,
	)
	list.SetVersion(0)
	list.SetClientPayloadReceiveCallback(onPayload)
	go list.StartListen()
	<-connectedCh
	return list
}

func newClientConnFromListener(list *transport.Listener, onPayload func([]byte)) *clientConn {
	isDisconnected := make(chan bool, 1)
	isLosted := make(chan bool, 1)
	connected := make(chan bool, 1)

	list.SetOnConnected(func() {
		select {
		case connected <- true:
		default:
		}
	})
	list.SetClientDisconnectCallback(func() {
		select {
		case isDisconnected <- true:
		default:
		}
	})
	list.SetClientLostConnectionCallback(func() {
		select {
		case isLosted <- true:
		default:
		}
	})
	list.SetClientPayloadReceiveCallback(onPayload)
	list.Reset()
	<-connected

	return &clientConn{
		listener:       list,
		isShutdown:     make(chan bool, 1),
		isLosted:       isLosted,
		isDisconnected: isDisconnected,
	}
}

func (c *clientConn) setPayloadCallback(fn func([]byte)) {
	c.listener.SetClientPayloadReceiveCallback(fn)
}

func (c *clientConn) setPongData(motd, players, maxPlayers string) {
	c.listener.SetPongData(motd, players, maxPlayers)
}

func (c *clientConn) sendPacketsRaw(pks map[int][]byte) {
	c.listener.SendPayload(c.encodeBatch(pks), 0)
}

func (c *clientConn) sendPacketPriority(pks map[int][]byte) {
	c.listener.SendPayload(c.encodeBatch(pks), 3)
}

func (c *clientConn) decodeBatch(buf []byte) (map[int][]byte, error) {
	var out bytes.Buffer
	r, err := zlib.NewReader(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(&out, r); err != nil {
		return nil, err
	}
	r.Close()
	packets := map[int][]byte{}
	for out.Len() != 0 {
		packets[len(packets)] = utils.ReadBytesLen(&out)
	}
	return packets, nil
}

func (c *clientConn) encodeBatch(pks map[int][]byte) []byte {
	var b bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&b, 7)
	l := make([]byte, 5)
	keys := make([]int, 0, len(pks))
	for k := range pks {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		_ = utils.WriteVarUInt32(zw, uint32(len(pks[k])), l)
		_, _ = zw.Write(pks[k])
	}
	_ = zw.Close()
	var bb bytes.Buffer
	bb.WriteByte(0xfe)
	bb.Write(b.Bytes())
	return bb.Bytes()
}
