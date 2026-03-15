package resource

type ResourcePackInfoEntry struct {
	PackID   string
	Version  string
	PackSize uint64
}

func NewResourcePackInfoEntry(packid string, version string, size uint64) *ResourcePackInfoEntry {
	return &ResourcePackInfoEntry{packid, version, size}
}

func (rpie *ResourcePackInfoEntry) GetPackID() string {
	return rpie.PackID
}

func (rpie *ResourcePackInfoEntry) GetVersion() string {
	return rpie.Version
}

func (rpie *ResourcePackInfoEntry) GetSize() uint64 {
	return rpie.PackSize
}
