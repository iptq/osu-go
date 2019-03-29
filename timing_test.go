package osu

import (
	"fmt"
	"testing"
)

var uTP = UninheritedTimingPoint{
	BPM:   200,
	Meter: 4,
	Time:  TimestampAbsolute(12345),
}

var iTP = InheritedTimingPoint{
	Parent: TimingPoint(uTP),
	Time: TimestampRelative{
		previous: uTP.GetTimestamp(),
		bpm:      uTP.GetBPM(),
		meter:    uTP.GetMeter(),
		measures: 1,
		num:      0,
		denom:    1,
	},
}

type testCase struct {
	t1 TimestampRelative
	t2 TimestampAbsolute
}

var testCases = []testCase{
	// no change from the measure at all
	{
		t1: TimestampRelative{
			previous: uTP.GetTimestamp(),
			bpm:      uTP.GetBPM(),
			meter:    uTP.GetMeter(),
			measures: 0,
			num:      0,
			denom:    1,
		},
		t2: TimestampAbsolute(12345),
	},

	// +1 measure (measure is 300ms, times 4 beats)
	{
		t1: TimestampRelative{
			previous: uTP.GetTimestamp(),
			bpm:      uTP.GetBPM(),
			meter:    uTP.GetMeter(),
			measures: 1,
			num:      0,
			denom:    1,
		},
		t2: TimestampAbsolute(13545),
	},

	// a single beat
	{
		t1: TimestampRelative{
			previous: uTP.GetTimestamp(),
			bpm:      uTP.GetBPM(),
			meter:    uTP.GetMeter(),
			measures: 0,
			num:      1,
			denom:    4,
		},
		t2: TimestampAbsolute(12645),
	},

	// half of a measure
	{
		t1: TimestampRelative{
			previous: uTP.GetTimestamp(),
			bpm:      uTP.GetBPM(),
			meter:    uTP.GetMeter(),
			measures: 0,
			num:      1,
			denom:    2,
		},
		t2: TimestampAbsolute(12945),
	},

	// 3 quarter notes
	{
		t1: TimestampRelative{
			previous: uTP.GetTimestamp(),
			bpm:      uTP.GetBPM(),
			meter:    uTP.GetMeter(),
			measures: 0,
			num:      3,
			denom:    4,
		},
		t2: TimestampAbsolute(13245),
	},

	// ok same thing again except with the inherited timing point
	// no change from the measure at all
	{
		t1: TimestampRelative{
			previous: iTP.GetTimestamp(),
			bpm:      iTP.GetBPM(),
			meter:    iTP.GetMeter(),
			measures: 0,
			num:      0,
			denom:    1,
		},
		t2: TimestampAbsolute(13545),
	},

	// +1 measure, same as above
	{
		t1: TimestampRelative{
			previous: iTP.GetTimestamp(),
			bpm:      iTP.GetBPM(),
			meter:    iTP.GetMeter(),
			measures: 1,
			num:      0,
			denom:    1,
		},
		t2: TimestampAbsolute(14745),
	},

	// a single beat
	{
		t1: TimestampRelative{
			previous: iTP.GetTimestamp(),
			bpm:      iTP.GetBPM(),
			meter:    iTP.GetMeter(),
			measures: 0,
			num:      1,
			denom:    4,
		},
		t2: TimestampAbsolute(13845),
	},

	// half of a measure
	{
		t1: TimestampRelative{
			previous: iTP.GetTimestamp(),
			bpm:      iTP.GetBPM(),
			meter:    iTP.GetMeter(),
			measures: 0,
			num:      1,
			denom:    2,
		},
		t2: TimestampAbsolute(14145),
	},

	// 3 quarter notes
	{
		t1: TimestampRelative{
			previous: iTP.GetTimestamp(),
			bpm:      iTP.GetBPM(),
			meter:    iTP.GetMeter(),
			measures: 0,
			num:      3,
			denom:    4,
		},
		t2: TimestampAbsolute(14445),
	},
}

func timingSubtest(c int, tcase testCase) func(t *testing.T) {
	return func(t *testing.T) {
		t1t := tcase.t1.Milliseconds()
		t2t := tcase.t2.Milliseconds()

		// check that they're equal first
		if t1t != t2t {
			t.Errorf("case %dA: expected %d, got %d (%v)", c, t2t, t1t, tcase.t1)
			return
		}

		// now check in reverse
		t2r, err := tcase.t2.IntoRelative(tcase.t1.previous, tcase.t1.bpm, tcase.t1.meter)
		if err != nil {
			t.Errorf("case %dB: error in IntoRelative: %v", c, err)
			return
		}
		// t.Logf("t2r = %+v\n", t2r)

		t2t = t2r.Milliseconds()
		if t1t != t2t {
			t.Errorf("case %dB: expected %d, got %d", c, t1t, t2t)
			return
		}
	}
}

func TestTiming(t *testing.T) {
	for c, tcase := range testCases {
		t.Run(fmt.Sprintf("test%d", c), timingSubtest(c, tcase))
	}
}
