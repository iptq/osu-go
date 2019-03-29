package osu

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/oklog/ulid"
)

type HitObject interface {
	// GetULID returns a unique identifier for this HitObject. This doesn't
	// necessarily need to persist between different instances of the editor
	// (i.e. doesn't need to be saved to disk)
	GetULID() ulid.ULID

	GetStartTime() Timestamp

	Serialize() (string, error)
}

type ObjCircle struct {
	ulid      ulid.ULID
	x, y      int
	startTime Timestamp
	newCombo  bool
	additions Hitsound
	extras    *Extras
}

func ParseHitCircle(params commonParameters, parts []string) (obj ObjCircle, err error) {
	var extras *Extras = &Extras{}

	if len(parts) > 5 {
		extras, err = ParseExtras(parts[5])
		if err != nil {
			return ObjCircle{}, err
		}
	}

	obj = ObjCircle{
		ulid:      NewULID(),
		x:         params.x,
		y:         params.y,
		startTime: TimestampAbsolute(params.startTime),
		newCombo:  params.newCombo,
		additions: Hitsound(params.hitsound),
		extras:    extras,
	}
	return
}

func (obj ObjCircle) GetULID() ulid.ULID {
	return obj.ulid
}

func (obj ObjCircle) GetStartTime() Timestamp {
	return obj.startTime
}

func (obj ObjCircle) Serialize() (string, error) {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%s",
		obj.x,
		obj.y,
		obj.startTime,
		1|(WHAT_THE_FUCK[obj.newCombo]<<2),
		obj.additions,
		obj.extras.String(),
	), nil
}

type ObjSlider struct {
	ulid      ulid.ULID
	x, y      int
	startTime Timestamp
	newCombo  bool
	additions Hitsound
	extras    *Extras

	splineKind    SplineKind
	ctlPoints     []IntPoint
	spline        []FloatPoint
	repeatCount   int
	pixelLength   float64
	edgeHitsounds []Hitsound
	edgeAdditions []Hitsound
}

func ParseSlider(params commonParameters, parts []string) (obj ObjSlider, err error) {
	var (
		pixelLength float64
		extras      *Extras = &Extras{}
	)

	// if len(parts) < 11 {
	// 	return ObjSlider{}, fmt.Errorf("len(slider) = %d < 11", len(parts))
	// }
	// extras, err := ParseExtras(parts[10])
	// if err != nil {
	// 	return ObjSlider{}, err
	// }

	if len(parts) > 7 {
		// pixelLength
		pixelLength, err = strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return
		}
	}

	if len(parts) > 10 {
		// extras
		extras, err = ParseExtras(parts[10])
		if err != nil {
			return
		}
	}

	kind, ctlPoints, err := ParseControlPoints(parts[5])
	ctlPoints = append([]IntPoint{IntPoint{params.x, params.y}}, ctlPoints...)
	spline, err := SplineFrom(kind, ctlPoints, pixelLength)

	obj = ObjSlider{
		ulid:      NewULID(),
		x:         params.x,
		y:         params.y,
		startTime: TimestampAbsolute(params.startTime),
		newCombo:  params.newCombo,
		additions: Hitsound(params.hitsound),
		extras:    extras,

		spline:      spline,
		splineKind:  kind,
		pixelLength: pixelLength,
	}
	return
}

func (obj ObjSlider) GetULID() ulid.ULID {
	return obj.ulid
}

func (obj ObjSlider) GetStartTime() Timestamp {
	return obj.startTime
}

func (obj ObjSlider) Serialize() (string, error) {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%c|%s,%d,%f,%s,%s,%s",
		obj.x,
		obj.y,
		obj.startTime,
		1|(WHAT_THE_FUCK[obj.newCombo]<<2),
		obj.additions,
		obj.splineKind,
		"TODO",
		obj.repeatCount,
		obj.pixelLength,
		"TODO",
		"TODO",
		obj.extras.String(),
	), nil
}

type ObjSpinner struct {
	ulid      ulid.ULID
	x, y      int
	startTime Timestamp
	endTime   Timestamp
	newCombo  bool
	additions Hitsound
	extras    *Extras
}

func ParseSpinner(params commonParameters, parts []string) (obj ObjSpinner, err error) {
	endTime, err := strconv.Atoi(parts[5])
	if err != nil {
		return
	}

	extras, err := ParseExtras(parts[6])
	if err != nil {
		return
	}

	obj = ObjSpinner{
		ulid:      NewULID(),
		x:         params.x,
		y:         params.y,
		startTime: TimestampAbsolute(params.startTime),
		endTime:   TimestampAbsolute(endTime),
		newCombo:  params.newCombo,
		additions: params.hitsound,
		extras:    extras,
	}
	return
}

func (obj ObjSpinner) GetULID() ulid.ULID {
	return obj.ulid
}

func (obj ObjSpinner) GetStartTime() Timestamp {
	return obj.startTime
}

func (obj ObjSpinner) Serialize() (string, error) {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%s",
		obj.x,
		obj.y,
		obj.startTime,
		8|(WHAT_THE_FUCK[obj.newCombo]<<2),
		obj.additions,
		obj.endTime,
		obj.extras.String(),
	), nil
}

type commonParameters struct {
	x, y      int
	startTime int
	newCombo  bool
	hitsound  int
}

func ParseHitObject(line string) (HitObject, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 5 {
		return nil, errors.New("len(parts) < 5")
	}

	x, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	y, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	startTime, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}

	ty, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, err
	}

	hitsound, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, err
	}

	newCombo := (ty & 4) > 0
	params := commonParameters{x, y, startTime, newCombo, hitsound}

	switch {
	case (ty & 1) > 0:
		return ParseHitCircle(params, parts)
	case (ty & 2) > 0:
		return ParseSlider(params, parts)
	case (ty & 8) > 0:
		return ParseSpinner(params, parts)
	default:
		return nil, fmt.Errorf("unknown hitobject type: %+v", ty)
	}

}
