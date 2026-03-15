package world

type Chunk struct {
	X         int32
	Z         int32
	SubChunks []*SubChunk
	
	Height [256]int16
}

func (chunk *Chunk) GetSubChunkAtFromPlayerCords(y int) *SubChunk {
	for ny, sch := range chunk.SubChunks {
		if ny == y>>4 {
			return sch
		}
	}
	return nil
}

func (chunk *Chunk) GetBlock(x, y, z int) (g byte) {
	defer func() {
		if r := recover(); r != nil {
			g = 0
		}
	}()
	return chunk.SubChunks[y>>4].GetBlock(x, y-16*(y>>4), z)
}

func (chunk *Chunk) GetBlockFromPlayerCords(x, y, z int) (g byte) {
	defer func() {
		if r := recover(); r != nil {
			g = 0
		}
	}()
	return chunk.SubChunks[(y&0xff)>>4].GetBlock(x&0x0f, (y&0xff)-16*(y>>4), z&0x0f)
}

func (chunk *Chunk) SetBlockFromPlayerCords(x, y, z int, bid byte) {
	chunk.SubChunks[(y&0xff)>>4].SetBlock(x&0x0f, (y&0xff)-16*(y>>4), z&0x0f, bid)
}

func (chunk *Chunk) SetBlock(x, y, z int, bid byte) {
	chunk.SubChunks[y>>4].SetBlock(x, y-16*(y>>4), z, bid)
}

func (chunk *Chunk) SetHeight(x, z int, h int16) {
	chunk.Height[((z << 4) + (x))] = h
}

func (chunk *Chunk) GetHeight(x, z int) int16 {
	return chunk.Height[((z << 4) + (x))]
}

func (chunk *Chunk) GetMetadata(x, y, z int) byte {
	return chunk.SubChunks[y>>4].GetMetadata(x, y-16*(y>>4), z)
}

func (chunk *Chunk) SetMetadata(x, y, z int, d byte) {
	chunk.SubChunks[y>>4].SetMetadata(x, y-16*(y>>4), z, d)
}
