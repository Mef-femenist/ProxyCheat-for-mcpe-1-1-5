package packet

import (
	"bytes"
	"fmt"
	"io"
)

type PacketData struct {
	h       *Header
	full    []byte
	payload *bytes.Buffer
}

type DataPacket interface {
	
	ID() byte
	
	Marshal(w *PacketWriter)
	
	Unmarshal(r *PacketReader)
}

type Header struct {
	PacketID byte
}

func (header *Header) Write(w io.ByteWriter) error {
	return WriteByte(w, header.PacketID)
}

func (header *Header) Read(r io.ByteReader) error {
	var value byte
	if err := ReadByte(r, &value); err != nil {
		return err
	}
	header.PacketID = value
	return nil
}

func ParseDataPacket(data []byte) *PacketData {
	buf := bytes.NewBuffer(data)
	hd, _ := buf.ReadByte()
	return &PacketData{h: &Header{PacketID: hd}, full: data, payload: buf}
}

func (p *PacketData) TryDecodePacket() (pk DataPacket, err error) {
	pkFunc, ok := RegisteredPackets[p.h.PacketID]
	r := NewReader(p.payload, 0) 
	if !ok {
		
		pk = &UnknownPacket{PacketID: p.h.PacketID}
		pk.Unmarshal(r)
		if p.payload.Len() != 0 {
			return pk, nil
		}
		return pk, nil
	}
	pk = pkFunc()

	defer func() {
		if recoveredErr := recover(); recoveredErr != nil {
			err = fmt.Errorf("%T: %w", pk, recoveredErr.(error))
		}
	}()
	pk.Unmarshal(r)
	if p.payload.Len() != 0 {
		return pk, nil
	}
	return pk, nil
}
