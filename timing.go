package osu

import (
	"fmt"
	"math"
	"sort"
)

var (
	// allow objects to be up to 2 milliseconds off
	ESTIMATE_THRESHOLD = 2.0

	// list of snappings that the editor uses
	SNAPPINGS = []int{1, 2, 3, 4, 6, 8, 12, 16}
)

type Timestamp interface {
	Milliseconds() int
}

type TimestampAbsolute int

func (t TimestampAbsolute) Milliseconds() int {
	return int(t)
}

type snapping struct {
	num   int
	denom int
	delta float64
}

type snappings []snapping

func (s snappings) Len() int {
	return len(s)
}

func (s snappings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s snappings) Less(i, j int) bool {
	return s[i].delta < s[j].delta
}

// IntoRelative attempts to convert an absolute timestamp into a relative one
func (t TimestampAbsolute) IntoRelative(to Timestamp, bpm float64, meter int) (*TimestampRelative, error) {
	// return nil, fmt.Errorf("to = %+v", to)

	msPerBeat := 60000.0 / bpm
	msPerMeasure := msPerBeat * float64(meter)

	base := to.Milliseconds()
	cur := t.Milliseconds()

	measures := int(float64(cur-base) / msPerMeasure)
	measureStart := float64(base) + float64(measures)*msPerMeasure
	offset := float64(cur) - measureStart

	snapTimes := make([]snapping, len(SNAPPINGS)*16)
	for _, denom := range SNAPPINGS {
		for i := 0; i < denom; i++ {
			var snapAt float64

			snapAt = msPerMeasure * float64(i) / float64(denom)
			snapTimes = append(snapTimes, snapping{
				num:   i,
				denom: denom,
				delta: math.Abs(offset - snapAt),
			})

			snapAt = msPerMeasure * float64(i+denom) / float64(denom)
			snapTimes = append(snapTimes, snapping{
				num:   i + denom,
				denom: denom,
				delta: math.Abs(offset - snapAt),
			})
		}
	}
	sort.Sort(snappings(snapTimes))

	first := snapTimes[0]
	if first.delta > ESTIMATE_THRESHOLD {
		return nil, fmt.Errorf("Could not find accurate snapping.")
	}

	t2 := &TimestampRelative{
		previous: to,
		bpm:      bpm,
		meter:    meter,
		measures: measures,
		num:      first.num,
		denom:    first.denom,
	}
	return t2, nil
}

type TimestampRelative struct {
	previous Timestamp
	bpm      float64
	meter    int

	measures int
	num      int
	denom    int
}

func (t TimestampRelative) Milliseconds() int {
	// fmt.Println("previous:", t.previous, t.previous.Milliseconds())
	base := t.previous.Milliseconds()
	msPerBeat := 60000.0 / t.bpm
	msPerMeasure := msPerBeat * float64(t.meter)

	measureOffset := msPerMeasure * float64(t.measures)
	remainingOffset := msPerMeasure * float64(t.num) / float64(t.denom)
	return int(float64(base) + measureOffset + remainingOffset)
}

type TimingPoint interface {
	// Get the timestamp
	GetTimestamp() Timestamp

	// Get the BPM of the nearest uninherited timing section to which this belongs
	GetBPM() float64

	// Get the meter of the nearest uninherited timing section to which this belongs
	GetMeter() int
}

type UninheritedTimingPoint struct {
	BPM   float64
	Meter int
	Time  Timestamp
}

func (tp UninheritedTimingPoint) GetTimestamp() Timestamp {
	return tp.Time
}

func (tp UninheritedTimingPoint) GetBPM() float64 {
	return tp.BPM
}

func (tp UninheritedTimingPoint) GetMeter() int {
	return tp.Meter
}

type InheritedTimingPoint struct {
	Parent       TimingPoint
	Time         Timestamp
	SvMultiplier float64
}

func (tp InheritedTimingPoint) GetTimestamp() Timestamp {
	return tp.Time
}

func (tp InheritedTimingPoint) GetBPM() float64 {
	return tp.Parent.GetBPM()
}

func (tp InheritedTimingPoint) GetMeter() int {
	return tp.Parent.GetMeter()
}
