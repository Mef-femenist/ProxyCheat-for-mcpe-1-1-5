package item

type Item struct {
	ID    int32
	Meta  int32
	Count int32
	Name  string
	NBT   map[string]interface{}
}

var itemNames = map[int32]string{
	358: "map",
}

func NewItem(id int32, meta int32, count int32, nbt map[string]interface{}) *Item {
	id = id & 0xffff
	
	if meta != -1 {
		meta = meta & 0xffff
	} else {
		meta = -1
	}
	name := ""
	if n, ok := itemNames[id]; ok {
		name = n
	}
	return &Item{id, meta, count, name, nbt}
}

func (item *Item) GetID() int32 {
	return item.ID
}

func (item *Item) GetName() string {
	return item.Name
}

func (item *Item) GetDamage() int32 {
	return item.Meta
}

func (item *Item) GetCount() int32 {
	return item.Count
}

func (item *Item) GetNBT() map[string]interface{} {
	return item.NBT
}
