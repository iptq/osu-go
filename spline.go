package osu

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type SplineKind = rune

const (
	SPLINE_LINEAR  = 'L'
	SPLINE_PERFECT = 'P'
	SPLINE_BEZIER  = 'B'
	SPLINE_CATMULL = 'C'

	CURVE_THRESHOLD = 1.0
)

func ParseControlPoints(line string) (kind SplineKind, points []IntPoint, err error) {
	pointsStr := strings.Split(line, "|")

	for i, s := range pointsStr {
		if i == 0 {
			kind = []rune(s)[0]
			continue
		}

		var x, y int
		pair := strings.Split(s, ":")

		x, err = strconv.Atoi(pair[0])
		if err != nil {
			return
		}

		y, err = strconv.Atoi(pair[1])
		if err != nil {
			return
		}

		points = append(points, IntPoint{x, y})
	}

	return
}

func SplineFrom(kind SplineKind, points []IntPoint, length float64) (spline []FloatPoint, err error) {
	switch kind {
	case SPLINE_LINEAR:
		if len(points) < 2 {
			err = errors.New("not enough points to create a linear spline")
			return
		} else if len(points) > 2 {
			err = errors.New("trying to create linear spline with more than 2 points")
			return
		}

		A := points[0].ToFloat()
		B := points[1].ToFloat()

		// since this is linear, and we can draw lines via the graphics library anyway,
		// we don't need to calculate a million points
		spline = append(spline, A)

		// for the end point, we have to shorten the slider based on the pixel length
		// given to us in the osu file
		unit := B.Sub(A).Norm()
		spline = append(spline, unit.ScalarMul(length))
	case SPLINE_PERFECT:
	case SPLINE_BEZIER:
	case SPLINE_CATMULL:
		// deprecated, but it still appears in older maps
		// so we'll just error it out for now and implement it later
		err = fmt.Errorf("catmull hasn't been implemented yet")
	default:
		err = fmt.Errorf("unknown spline kind: %v", kind)
	}
	return
}
