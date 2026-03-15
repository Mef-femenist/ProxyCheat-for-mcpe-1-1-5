package voxelraytrace

import (
	"errors"
	"math"
)

type Vec3 [3]float64

func (v Vec3) X() float64 { return v[0] }
func (v Vec3) Y() float64 { return v[1] }
func (v Vec3) Z() float64 { return v[2] }

func (v Vec3) Add(o Vec3) Vec3 { return Vec3{v[0] + o[0], v[1] + o[1], v[2] + o[2]} }
func (v Vec3) Sub(o Vec3) Vec3 { return Vec3{v[0] - o[0], v[1] - o[1], v[2] - o[2]} }
func (v Vec3) Mul(s float64) Vec3 { return Vec3{v[0] * s, v[1] * s, v[2] * s} }
func (v Vec3) LenSqr() float64   { return v[0]*v[0] + v[1]*v[1] + v[2]*v[2] }
func (v Vec3) Normalize() Vec3 {
	l := math.Sqrt(v.LenSqr())
	if l == 0 {
		return Vec3{}
	}
	return Vec3{v[0] / l, v[1] / l, v[2] / l}
}

func InDirection(start, directionVector Vec3, maxDistance float64) (vectors []Vec3, err error) {
	return BetweenPoints(start, start.Add(directionVector.Mul(maxDistance)))
}

func BetweenPoints(start, end Vec3) (vectors []Vec3, err error) {
	currentPoint := Vec3{math.Floor(start.X()), math.Floor(start.Y()), math.Floor(start.Z())}

	directionVector := end.Sub(start).Normalize()
	if directionVector.LenSqr() <= 0 {
		return nil, errors.New("start and end points are the same, giving a zero direction vector")
	}

	radius := distance(start, end)

	stepX := compareTo(directionVector.X(), 0)
	stepY := compareTo(directionVector.Y(), 0)
	stepZ := compareTo(directionVector.Z(), 0)

	tMaxX := rayTraceDistanceToBoundary(start.X(), directionVector.X())
	tMaxY := rayTraceDistanceToBoundary(start.Y(), directionVector.Y())
	tMaxZ := rayTraceDistanceToBoundary(start.Z(), directionVector.Z())

	tDeltaX := findDelta(directionVector.X(), stepX)
	tDeltaY := findDelta(directionVector.Y(), stepY)
	tDeltaZ := findDelta(directionVector.Z(), stepZ)

	for {
		vectors = append(vectors, currentPoint)

		if tMaxX < tMaxY && tMaxX < tMaxZ {
			if tMaxX > radius {
				break
			}
			currentPoint = currentPoint.Add(Vec3{stepX})
			tMaxX += tDeltaX
		} else if tMaxY < tMaxZ {
			if tMaxY > radius {
				break
			}
			currentPoint = currentPoint.Add(Vec3{0, stepY})
			tMaxY += tDeltaY
		} else {
			if tMaxZ > radius {
				break
			}
			currentPoint = currentPoint.Add(Vec3{0, 0, stepZ})
			tMaxZ += tDeltaZ
		}
	}

	return
}

func findDelta(first, second float64) float64 {
	if first == 0 {
		return 0
	}
	return second / first
}

func rayTraceDistanceToBoundary(first, second float64) float64 {
	if second == 0 {
		return math.Inf(0)
	}
	if second < 0 {
		first = -first
		second = -second
		if math.Floor(first) == first {
			return 0
		}
	}
	return (1 - (first - math.Floor(first))) / second
}

func compareTo(first, second float64) float64 {
	if first == second {
		return 0
	} else if first < second {
		return -1
	} else {
		return 1
	}
}

func distance(a, b Vec3) float64 {
	xDiff, yDiff, zDiff := b[0]-a[0], b[1]-a[1], b[2]-a[2]
	return math.Sqrt(xDiff*xDiff + yDiff*yDiff + zDiff*zDiff)
}
