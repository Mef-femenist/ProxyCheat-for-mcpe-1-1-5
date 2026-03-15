package world

type SubChunk struct {
	Blocks   []byte 
	Metadata []byte 
	
}

func NewSubChunk(blocks, meta []byte) *SubChunk {
	ch := &SubChunk{Blocks: blocks, Metadata: meta}
	
	return ch
}

func (SubChunk *SubChunk) GetIndex(x, y, z int) int {
	return (x * 256) + (z * 16) + y
}

func (SubChunk *SubChunk) GetBlock(x, y, z int) byte {
	return SubChunk.Blocks[SubChunk.GetIndex(x, y, z)]
}

func (SubChunk *SubChunk) SetBlock(x, y, z int, bid byte) {
	SubChunk.Blocks[SubChunk.GetIndex(x, y, z)] = bid
}

func (SubChunk *SubChunk) GetMetadata(x, y, z int) byte {
	return SubChunk.Metadata[SubChunk.GetIndex(x, y, z)]
}

func (SubChunk *SubChunk) SetMetadata(x, y, z int, data byte) {
	SubChunk.Metadata[SubChunk.GetIndex(x, y, z)] = data
}
