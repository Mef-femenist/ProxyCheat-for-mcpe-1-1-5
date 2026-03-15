package socks5

import (
	"io"
	"net"
)

func (r *Request) Connect(w io.Writer) (*net.TCPConn, error) {
	tmp, err := Dial.Dial("tcp", r.Address())
	if err != nil {
		var p *Reply
		if r.Atyp == ATYPIPv4 || r.Atyp == ATYPDomain {
			p = NewReply(RepHostUnreachable, ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = NewReply(RepHostUnreachable, ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(w); err != nil {
			return nil, err
		}
		return nil, err
	}
	rc := tmp.(*net.TCPConn)

	a, addr, port, err := ParseAddress(rc.LocalAddr().String())
	if err != nil {
		var p *Reply
		if r.Atyp == ATYPIPv4 || r.Atyp == ATYPDomain {
			p = NewReply(RepHostUnreachable, ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = NewReply(RepHostUnreachable, ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(w); err != nil {
			return nil, err
		}
		return nil, err
	}
	p := NewReply(RepSuccess, a, addr, port)
	if _, err := p.WriteTo(w); err != nil {
		return nil, err
	}

	return rc, nil
}
