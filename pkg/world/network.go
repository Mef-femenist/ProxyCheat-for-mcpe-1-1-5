package world

import (
	"bytes"
	"mefproxy/pkg/packet"
	"sync"
)

var SavedChunks []*Chunk
var Mux sync.Mutex

func init() {

}

func ResetWorld() {
	Mux.Lock()
	defer Mux.Unlock()
	SavedChunks = []*Chunk{}
}

func DecodeNetwork(x int32, z int32, all []byte) {
	jnk := make([]byte, 1)
	
	bb := bytes.NewBuffer(all)
	r := packet.NewReader(bb, 2)

	var subChunkCount byte
	r.Uint8(&subChunkCount)
	
	var subChunks []*SubChunk

	for y := 0; y < int(subChunkCount); y++ {

		r.BytesLength(jnk)
		a1 := make([]byte, 4096)
		r.BytesLength(a1)
		a2 := make([]byte, 2048)
		r.BytesLength(a2)
		a3 := make([]byte, 2048)
		r.BytesLength(a3)
		a4 := make([]byte, 2048)
		r.BytesLength(a4)
		
		subChunks = append(subChunks, NewSubChunk(a1, a2))
	}

	g1 := make([]byte, 512)
	r.ByteSlice(&g1)
	biomeIds := make([]byte, 256)
	r.BytesLength(biomeIds)

	Mux.Lock()
	defer Mux.Unlock()
	SavedChunks = append(SavedChunks, &Chunk{
		X:         x,
		Z:         z,
		SubChunks: subChunks,
		Height:    [256]int16{},
	})
}
func GetChunk(x, z int32) *Chunk {
	Mux.Lock()
	defer Mux.Unlock()
	for _, ch := range SavedChunks {
		if ch.X == x && ch.Z == z {
			return ch
		}
	}
	return nil
}
func GetChunkAtFromPlayerCords(x, z int32) *Chunk {
	Mux.Lock()
	defer Mux.Unlock()
	for _, ch := range SavedChunks {
		if ch.X == x>>4 && ch.Z == z>>4 {
			return ch
		}
	}
	return nil
}
