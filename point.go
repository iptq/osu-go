package osu

import "math"

type IntPoint struct {
	x, y int
}

func (p IntPoint) ToFloat() FloatPoint {
	return FloatPoint{
		x: float64(p.x),
		y: float64(p.y),
	}
}

type FloatPoint struct {
	x, y float64
}

func (p FloatPoint) ToInt() IntPoint {
	return IntPoint{
		x: int(p.x),
		y: int(p.y),
	}
}

func (p1 FloatPoint) Add(p2 FloatPoint) FloatPoint {
	return FloatPoint{
		x: p1.x + p2.x,
		y: p1.y + p2.y,
	}
}

func (p1 FloatPoint) Sub(p2 FloatPoint) FloatPoint {
	return p1.Add(p2.ScalarMul(-1))
}

func (p FloatPoint) ScalarMul(c float64) FloatPoint {
	return FloatPoint{
		x: p.x * c,
		y: p.y * c,
	}
}

func (p FloatPoint) Magnitude() float64 {
	return math.Pow(p.x*p.x+p.y*p.y, 0.5)
}

func (p FloatPoint) Norm() FloatPoint {
	magnitude := p.Magnitude()
	return FloatPoint{
		x: p.x / magnitude,
		y: p.y / magnitude,
	}
}
