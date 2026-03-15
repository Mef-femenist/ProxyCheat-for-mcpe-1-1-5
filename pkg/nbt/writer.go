package nbt

import "io"

type offsetWriter struct {
	io.Writer
	off int64

	WriteByte func(byte) error
}

func (w *offsetWriter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	w.off += int64(n)
	return
}
