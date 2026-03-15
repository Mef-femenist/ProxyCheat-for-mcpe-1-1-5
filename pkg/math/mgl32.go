package mgl32

import "math"

type Vec2 [2]float32
type Vec3 [3]float32

func (v Vec3) X() float32 { return v[0] }
func (v Vec3) Y() float32 { return v[1] }
func (v Vec3) Z() float32 { return v[2] }

func (v Vec2) X() float32 { return v[0] }
func (v Vec2) Y() float32 { return v[1] }

func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v[0] + other[0], v[1] + other[1], v[2] + other[2]}
}

func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v[0] - other[0], v[1] - other[1], v[2] - other[2]}
}

func (v Vec3) Mul(s float32) Vec3 {
	return Vec3{v[0] * s, v[1] * s, v[2] * s}
}

func (v Vec3) Len() float32 {
	return float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
}

func (v Vec3) Normalize() Vec3 {
	l := v.Len()
	if l == 0 {
		return Vec3{}
	}
	return Vec3{v[0] / l, v[1] / l, v[2] / l}
}
