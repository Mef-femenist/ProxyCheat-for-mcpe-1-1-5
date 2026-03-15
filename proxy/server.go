package proxy

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"sort"

	"mefproxy/pkg/packet/encrypter"
	"mefproxy/pkg/transport"
	"mefproxy/utils"
)

type serverConn struct {
	connection     *transport.Connector
	isShutdown     chan bool
	isLosted       chan bool
	isDisconnected chan bool
	makeTransfer   chan bool
	encrypt        *encrypter.Instance
}

var connectionClosed = make(chan bool, 4)

func newServerConn(ip string, port, bindPort int) (*serverConn, error) {
	conn := transport.NewConnector(func() {
		connectionClosed <- true
	}, false)

	conn.SetVersion(8)
	conn.SetInfo(ip, port, bindPort)

	isDisconnected := make(chan bool)
	isLosted := make(chan bool)
	isShutdown := make(chan bool)
	done := make(chan bool)
	var connErr error

	conn.SetServerConnectionTimeoutCallback(func() {
		connErr = errors.New("server connection timeout")
		done <- true
	})
	conn.SetServerDisconnectCallback(func() {
		isDisconnected <- true
		close(isDisconnected)
	})
	conn.SetServerLostConnectionCallback(func() {
		isLosted <- true
		close(isLosted)
	})

	go conn.Connect(func() { done <- true })
	<-done
	close(done)

	return &serverConn{
		connection:     conn,
		isShutdown:     isShutdown,
		isLosted:       isLosted,
		isDisconnected: isDisconnected,
		makeTransfer:   make(chan bool),
		encrypt:        &encrypter.Instance{IsEnabled: false},
	}, connErr
}

func (s *serverConn) setPayloadCallback(fn func([]byte)) {
	s.connection.SetServerPayloadReceiveCallback(fn)
}

func (s *serverConn) sendPacketsRaw(pks map[int][]byte) {
	s.connection.SendPayload(s.encodeBatch(pks), 0)
}

func (s *serverConn) sendPayloadDirect(data []byte) {
	s.connection.SendPayload(data, 0)
}

func (s *serverConn) decodeBatch(buf []byte) map[int][]byte {
	var out bytes.Buffer
	if s.encrypt.IsEnabled {
		buf = s.encrypt.DecodePayload(buf)
	}
	r, err := zlib.NewReader(bytes.NewReader(buf))
	if err != nil {
		return nil
	}
	_, _ = io.Copy(&out, r)
	r.Close()
	packets := make(map[int][]byte)
	for out.Len() != 0 {
		packets[len(packets)] = utils.ReadBytesLen(&out)
	}
	return packets
}

func (s *serverConn) encodeBatch(pks map[int][]byte) []byte {
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
	if s.encrypt != nil && s.encrypt.IsEnabled {
		bb.Write(s.encrypt.EncodePayload(b.Bytes()))
	} else {
		bb.Write(b.Bytes())
	}
	return bb.Bytes()
}
