package block

import (
	"errors"
	"mefproxy/pkg/math/vec3"
)

type Block struct {
	ID       int32
	Meta     int32
	Name     string
	Position *vec3.Vector3
}

func NewBlock(id int32, meta int32, pos *vec3.Vector3) (*Block, error) {
	if id > 0xff {
		return nil, errors.New("block id larger than 255")
	}
	return &Block{
		ID:       id,
		Meta:     meta,
		Position: pos,
	}, nil
}

func (block *Block) GetID() int32 {
	return block.ID
}

func (block *Block) GetPosition() *vec3.Vector3 {
	return block.Position
}
