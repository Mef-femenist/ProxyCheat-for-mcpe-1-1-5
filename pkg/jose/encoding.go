package jose

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/binary"
	"io"
	"math/big"
	"strings"
	"unicode"

	"mefproxy/pkg/jose/json"
)

func mustSerializeJSON(value interface{}) []byte {
	out, err := json.Marshal(value)
	out = []byte(string(out) + "\n")
	if err != nil {
		panic(err)
	}
	
	if string(out) == "null" {
		panic("Tried to serialize a nil pointer.")
	}
	return out
}

func stripWhitespace(data string) string {
	buf := strings.Builder{}
	buf.Grow(len(data))
	for _, r := range data {
		if !unicode.IsSpace(r) {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func compress(algorithm CompressionAlgorithm, input []byte) ([]byte, error) {
	switch algorithm {
	case DEFLATE:
		return deflate(input)
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

func decompress(algorithm CompressionAlgorithm, input []byte) ([]byte, error) {
	switch algorithm {
	case DEFLATE:
		return inflate(input)
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

func deflate(input []byte) ([]byte, error) {
	output := new(bytes.Buffer)

	writer, _ := flate.NewWriter(output, 1)
	_, _ = io.Copy(writer, bytes.NewBuffer(input))

	err := writer.Close()
	return output.Bytes(), err
}

func inflate(input []byte) ([]byte, error) {
	output := new(bytes.Buffer)
	reader := flate.NewReader(bytes.NewBuffer(input))

	_, err := io.Copy(output, reader)
	if err != nil {
		return nil, err
	}

	err = reader.Close()
	return output.Bytes(), err
}

type byteBuffer struct {
	data []byte
}

func newBuffer(data []byte) *byteBuffer {
	if data == nil {
		return nil
	}
	return &byteBuffer{
		data: data,
	}
}

func newFixedSizeBuffer(data []byte, length int) *byteBuffer {
	if len(data) > length {
		panic("square/go-jose: invalid call to newFixedSizeBuffer (len(data) > length)")
	}
	pad := make([]byte, length-len(data))
	return newBuffer(append(pad, data...))
}

func newBufferFromInt(num uint64) *byteBuffer {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, num)
	return newBuffer(bytes.TrimLeft(data, "\x00"))
}

func (b *byteBuffer) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.base64())
}

func (b *byteBuffer) UnmarshalJSON(data []byte) error {
	var encoded string
	err := json.Unmarshal(data, &encoded)
	if err != nil {
		return err
	}

	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	*b = *newBuffer(decoded)

	return nil
}

func (b *byteBuffer) base64() string {
	return base64.RawURLEncoding.EncodeToString(b.data)
}

func (b *byteBuffer) bytes() []byte {
	
	if b == nil {
		return nil
	}
	return b.data
}

func (b byteBuffer) bigInt() *big.Int {
	return new(big.Int).SetBytes(b.data)
}

func (b byteBuffer) toInt() int {
	return int(b.bigInt().Int64())
}
