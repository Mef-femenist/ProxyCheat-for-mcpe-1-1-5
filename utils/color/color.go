package color

type Color struct {
	Red   int32
	Green int32
	Blue  int32
	Alpha int32

	WasCode int32
}

func FromARGB(code int32) Color {
	return Color{(code >> 16) & 0xff, (code >> 8) & 0xff, code & 0xff, (code >> 24) & 0xff, code}
}

func (cl *Color) ToARGB() (int32, int32, int32, int32, int32) {
	return cl.Red << 16, cl.Green << 8, cl.Blue, cl.Alpha << 24, cl.WasCode
}
