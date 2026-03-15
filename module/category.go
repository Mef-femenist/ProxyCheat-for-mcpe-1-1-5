package module

type Category int

const (
	Combat   Category = iota
	Movement
	Misc
	Exploit
	Visual
	Network
)

func (c Category) String() string {
	switch c {
	case Combat:
		return "Combat"
	case Movement:
		return "Movement"
	case Misc:
		return "Misc"
	case Exploit:
		return "Exploit"
	case Visual:
		return "Visual"
	case Network:
		return "Network"
	default:
		return "Unknown"
	}
}

func (c Category) Color() string {
	switch c {
	case Combat:
		return "§c"
	case Movement:
		return "§a"
	case Misc:
		return "§7"
	case Exploit:
		return "§6"
	case Visual:
		return "§b"
	case Network:
		return "§d"
	default:
		return "§f"
	}
}
