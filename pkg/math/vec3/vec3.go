package vec3

import (
	"mefproxy/pkg/math"
	"math"
)

type Vector3 struct {
	X, Y, Z, Pitch, Yaw, HeadYaw float32
}

func Add(a, b Vector3) Vector3 {
	return Vector3{a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.Pitch, a.Yaw, a.HeadYaw}
}

func Mult(a Vector3, b float32) Vector3 {
	return Vector3{a.X * b, a.Y * b, a.Z * b, a.Pitch, a.Yaw, a.HeadYaw}
}

func Subtract(a Vector3, b Vector3) Vector3 {
	return Add(a, Vector3{X: -b.X, Y: -b.Y, Z: -b.Z})
}

func (a Vector3) Length() float32 {
	return float32(math.Sqrt(float64(a.X*a.X + a.Y*a.Y + a.Z*a.Z)))
}

func Distance(a, b Vector3) float32 {
	xDiff := a.X - b.X
	yDiff := b.Y - b.Y
	zDiff := a.Z - b.Z
	return float32(math.Sqrt(float64(xDiff*xDiff + yDiff*yDiff + zDiff*zDiff)))
}

func DistanceSquared(a, b Vector3) float32 {
	xDiff := a.X - b.X
	yDiff := b.Y - b.Y
	zDiff := a.Z - b.Z
	return xDiff*xDiff + yDiff*yDiff + zDiff*zDiff
}

func Normalize(a Vector3) Vector3 {
	lend := a.Length()
	return Vector3{a.X / lend, a.Y / lend, a.Z / lend, a.Pitch, a.Yaw, a.HeadYaw}
}

func ToMgl32(a Vector3) mgl32.Vec3 {
	return mgl32.Vec3{a.X, a.Y, a.Z}
}
