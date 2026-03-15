package socks5

import (
	"bytes"
	"io"
	"net"
)

func (r *Request) UDP(c *net.TCPConn, serverAddr *net.UDPAddr) (*net.UDPAddr, error) {
	var clientAddr *net.UDPAddr
	var err error
	if bytes.Compare(r.DstPort, []byte{0x00, 0x00}) == 0 {
		
		clientAddr, err = net.ResolveUDPAddr("udp", c.RemoteAddr().String())
	} else {
		clientAddr, err = net.ResolveUDPAddr("udp", r.Address())
	}
	if err != nil {
		var p *Reply
		if r.Atyp == ATYPIPv4 || r.Atyp == ATYPDomain {
			p = NewReply(RepHostUnreachable, ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = NewReply(RepHostUnreachable, ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(c); err != nil {
			return nil, err
		}
		return nil, err
	}
	a, addr, port, err := ParseAddress(serverAddr.String())
	if err != nil {
		var p *Reply
		if r.Atyp == ATYPIPv4 || r.Atyp == ATYPDomain {
			p = NewReply(RepHostUnreachable, ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = NewReply(RepHostUnreachable, ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(c); err != nil {
			return nil, err
		}
		return nil, err
	}
	p := NewReply(RepSuccess, a, addr, port)
	if _, err := p.WriteTo(c); err != nil {
		return nil, err
	}

	return clientAddr, nil
}

func NewReply(rep byte, atyp byte, bndaddr []byte, bndport []byte) *Reply {
	if atyp == ATYPDomain {
		bndaddr = append([]byte{byte(len(bndaddr))}, bndaddr...)
	}
	return &Reply{
		Ver:     Ver,
		Rep:     rep,
		Rsv:     0x00,
		Atyp:    atyp,
		BndAddr: bndaddr,
		BndPort: bndport,
	}
}

func (r *Reply) WriteTo(w io.Writer) (int64, error) {
	var n int
	i, err := w.Write([]byte{r.Ver, r.Rep, r.Rsv, r.Atyp})
	n = n + i
	if err != nil {
		return int64(n), err
	}
	i, err = w.Write(r.BndAddr)
	n = n + i
	if err != nil {
		return int64(n), err
	}
	i, err = w.Write(r.BndPort)
	n = n + i
	if err != nil {
		return int64(n), err
	}
	return int64(n), nil
}
