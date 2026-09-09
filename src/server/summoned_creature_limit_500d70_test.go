package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const summonedCreatureLimitHighOwner500D70 = uint64(0x7f031b61e540)

type summonedCreatureLimitTestWorld500D70 struct {
	events      []string
	faultAt     int
	guideResult int32
	countResult int32
}

func (w *summonedCreatureLimitTestWorld500D70) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *summonedCreatureLimitTestWorld500D70) hooks() summonedCreatureLimitHooks500D70[uint64] {
	return summonedCreatureLimitHooks500D70[uint64]{
		loadGuideSize: func(index int32) int32 {
			w.observe(fmt.Sprintf("guide:%d", index))
			return w.guideResult
		},
		countControlled: func(owner uint64) int32 {
			w.observe(fmt.Sprintf("count:%x", owner))
			return w.countResult
		},
	}
}

func TestSummonedCreatureLimitCheck500D70LowByteAndBoundary(t *testing.T) {
	for _, tc := range []struct {
		name        string
		guide       int32
		count       int32
		wantAllowed bool
	}{
		{name: "exact limit", guide: 0x12345604, count: 0, wantAllowed: true},
		{name: "over limit", guide: 0x12345604, count: 1, wantAllowed: false},
		{name: "upper guide bits ignored", guide: 0x7fffff01, count: 3, wantAllowed: true},
		{name: "low byte 255", guide: -1, count: 0, wantAllowed: false},
		{name: "negative count", guide: 4, count: -1, wantAllowed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &summonedCreatureLimitTestWorld500D70{
				guideResult: tc.guide,
				countResult: tc.count,
			}
			if got := summonedCreatureLimitCheck500D70(uint64(1), 75, w.hooks()); got != tc.wantAllowed {
				t.Fatalf("allowed = %t, want %t", got, tc.wantAllowed)
			}
		})
	}
}

func TestSummonedCreatureLimitCheck500D70OrderAndNativeWidthOwner(t *testing.T) {
	w := &summonedCreatureLimitTestWorld500D70{guideResult: 2, countResult: 2}
	if !summonedCreatureLimitCheck500D70(summonedCreatureLimitHighOwner500D70, -74, w.hooks()) {
		t.Fatal("exactly four controlled-creature slots were rejected")
	}
	want := []string{"guide:-74", "count:7f031b61e540"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact oracle order %q", w.events, want)
	}
}

func TestSummonedCreatureLimitCheck500D70FaultPrefixes(t *testing.T) {
	baseline := &summonedCreatureLimitTestWorld500D70{guideResult: 1}
	summonedCreatureLimitCheck500D70(uint64(0), 5, baseline.hooks())
	want := []string{"guide:5", "count:0"}
	if !reflect.DeepEqual(baseline.events, want) {
		t.Fatalf("zero-owner events = %q, want %q", baseline.events, want)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &summonedCreatureLimitTestWorld500D70{faultAt: faultAt, guideResult: 1}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				summonedCreatureLimitCheck500D70(uint64(0), 5, w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}

func TestSummonedCreatureLimitCheck500D70WrapsBeforeSignedCompare(t *testing.T) {
	w := &summonedCreatureLimitTestWorld500D70{
		guideResult: 0xff,
		countResult: math.MaxInt32,
	}
	if !summonedCreatureLimitCheck500D70(summonedCreatureLimitHighOwner500D70, 5, w.hooks()) {
		t.Fatal("wrapped signed-negative PE32 total was rejected")
	}
	if hostTotal := int64(math.MaxInt32) + 0xff; hostTotal <= int64(summonedCreatureLimit500D70) {
		t.Fatalf("test setup did not distinguish host-width addition: %d", hostTotal)
	}
}
