package socks5

const (
	
	Ver byte = 0x05

	MethodNone byte = 0x00

	MethodUsernamePassword byte = 0x02 

	UserPassVer byte = 0x01
	
	UserPassStatusSuccess byte = 0x00

	CmdConnect byte = 0x01

	CmdUDP byte = 0x03

	ATYPIPv4 byte = 0x01 
	
	ATYPDomain byte = 0x03 
	
	ATYPIPv6 byte = 0x04 

	RepSuccess byte = 0x00

	RepHostUnreachable byte = 0x04
)

type NegotiationRequest struct {
	Ver      byte
	NMethods byte
	Methods  []byte 
}

type NegotiationReply struct {
	Ver    byte
	Method byte
}

type UserPassNegotiationRequest struct {
	Ver    byte
	Ulen   byte
	Uname  []byte 
	Plen   byte
	Passwd []byte 
}

type UserPassNegotiationReply struct {
	Ver    byte
	Status byte
}

type Request struct {
	Ver     byte
	Cmd     byte
	Rsv     byte 
	Atyp    byte
	DstAddr []byte
	DstPort []byte 
}

type Reply struct {
	Ver  byte
	Rep  byte
	Rsv  byte 
	Atyp byte
	
	BndAddr []byte
	
	BndPort []byte 
}

type Datagram struct {
	Rsv     []byte 
	Frag    byte
	Atyp    byte
	DstAddr []byte
	DstPort []byte 
	Data    []byte
}
