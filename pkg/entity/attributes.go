package entity

type Attribute struct {
	ID           int32
	MinValue     float32
	MaxValue     float32
	Value        float32
	DefaultValue float32
	Name         string
}

func NewAttribute(id int32, name string, min, max, def float32) Attribute {
	return Attribute{
		MinValue:     min,
		MaxValue:     max,
		Value:        def,
		DefaultValue: def,
		Name:         name,
		ID:           id,
	}
}
