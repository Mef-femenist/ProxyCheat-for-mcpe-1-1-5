package nbt

import (
	"io"
)

type offsetReader struct {
	io.Reader
	off int64

	ReadByte func() (byte, error)
	
	Next func(n int) []byte
}

func newOffsetReader(r io.Reader) *offsetReader {
	reader := &offsetReader{Reader: r}
	if byteReader, ok := r.(io.ByteReader); ok {
		reader.ReadByte = func() (byte, error) {
			reader.off++
			return byteReader.ReadByte()
		}
	} else {
		reader.ReadByte = func() (byte, error) {
			data := make([]byte, 1)
			_, err := io.ReadAtLeast(reader, data, 1)
			return data[0], err
		}
	}
	if r, ok := r.(interface {
		Next(n int) []byte
	}); ok {
		reader.Next = func(n int) []byte {
			data := r.Next(n)
			reader.off += int64(len(data))
			return data
		}
	} else {
		reader.Next = func(n int) []byte {
			data := make([]byte, n)
			_, _ = io.ReadAtLeast(reader, data, n)
			return data
		}
	}
	return reader
}

func (b *offsetReader) Read(p []byte) (n int, err error) {
	n, err = io.ReadAtLeast(b.Reader, p, len(p))
	b.off += int64(n)
	return
}
