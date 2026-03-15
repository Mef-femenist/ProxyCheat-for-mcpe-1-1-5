package packet

import "io"

func ReadByte(r io.ByteReader, out *byte) error {
	g, err  := r.ReadByte()
	if err != nil{
		return err
	}
	out = &g
	return nil
}

func WriteByte(w io.ByteWriter, in byte) error {
	return w.WriteByte(in)
}