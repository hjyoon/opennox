package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	positionDeltaObject4FEA70 = uint64(0x100000101)
	positionDeltaPoint4FEA70  = uint64(0x200000202)
)

type positionDeltaTestWorld4FEA70 struct {
	events  []string
	faultAt int
	pointX  float32
	objectX float32
	pointY  float32
	objectY float32
}

func (w *positionDeltaTestWorld4FEA70) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *positionDeltaTestWorld4FEA70) hooks() positionDeltaHooks4FEA70[uint64, uint64] {
	return positionDeltaHooks4FEA70[uint64, uint64]{
		loadPointX: func(point uint64) float32 {
			w.observe(fmt.Sprintf("point-x:%x", point))
			return w.pointX
		},
		loadObjectX: func(object uint64) float32 {
			w.observe(fmt.Sprintf("object-x:%x", object))
			return w.objectX
		},
		loadPointY: func(point uint64) float32 {
			w.observe(fmt.Sprintf("point-y:%x", point))
			return w.pointY
		},
		loadObjectY: func(object uint64) float32 {
			w.observe(fmt.Sprintf("object-y:%x", object))
			return w.objectY
		},
	}
}

func TestPositionDelta4FEA70ExactLoadOrderAndEagerY(t *testing.T) {
	w := &positionDeltaTestWorld4FEA70{
		pointX:  12,
		objectX: 1,
		pointY:  4,
		objectY: 3,
	}

	if got := positionDelta4FEA70(positionDeltaObject4FEA70, positionDeltaPoint4FEA70, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want canonical one", got)
	}
	want := []string{
		"point-x:200000202",
		"object-x:100000101",
		"point-y:200000202",
		"object-y:100000101",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact eager-load order %q", w.events, want)
	}
}

func TestPositionDelta4FEA70ThresholdsAndSpecialValues(t *testing.T) {
	negativeZero := math.Float32frombits(0x80000000)
	nan := math.Float32frombits(0x7fc12345)
	below := math.Nextafter32(positionDeltaLimit4FEA70, 0)
	tests := []struct {
		name            string
		pointX, objectX float32
		pointY, objectY float32
		want            int32
	}{
		{name: "equal", pointX: 7, objectX: 7, pointY: -3, objectY: -3},
		{name: "X just below", pointX: below, pointY: 1, want: 0},
		{name: "X exact positive", pointX: 5, want: 1},
		{name: "X exact negative", pointX: -5, want: 1},
		{name: "Y just below", pointY: below, want: 0},
		{name: "Y exact positive", pointY: 5, want: 1},
		{name: "Y exact negative", pointY: -5, want: 1},
		{name: "positive infinity", pointX: float32(math.Inf(1)), want: 1},
		{name: "negative infinity", pointY: float32(math.Inf(-1)), want: 1},
		{name: "infinity minus infinity unordered", pointX: float32(math.Inf(1)), objectX: float32(math.Inf(1))},
		{name: "X NaN unordered", pointX: nan, pointY: 4},
		{name: "Y NaN unordered", pointX: 4, pointY: nan},
		{name: "signed zero", pointX: negativeZero, objectY: negativeZero},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &positionDeltaTestWorld4FEA70{
				pointX: tc.pointX, objectX: tc.objectX,
				pointY: tc.pointY, objectY: tc.objectY,
			}
			if got := positionDelta4FEA70(positionDeltaObject4FEA70, positionDeltaPoint4FEA70, w.hooks()); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if len(w.events) != 4 {
				t.Fatalf("coordinate loads = %q, want all four", w.events)
			}
		})
	}
}

func TestPositionDeltaNormalizeX874FEA70StatusBitSemantics(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	if got := positionDeltaNormalizeX87_4FEA70(negativeZero); math.Float64bits(got) != math.Float64bits(negativeZero) {
		t.Fatalf("negative zero bits = %#016x, want %#016x", math.Float64bits(got), math.Float64bits(negativeZero))
	}

	positiveNaN := math.Float64frombits(0x7ff8000000012345)
	negativeNaN := math.Float64frombits(0xfff8000000054321)
	if got := positionDeltaNormalizeX87_4FEA70(positiveNaN); math.Float64bits(got) != math.Float64bits(positiveNaN)^uint64(1<<63) {
		t.Fatalf("positive NaN normalization bits = %#016x", math.Float64bits(got))
	}
	if got := positionDeltaNormalizeX87_4FEA70(negativeNaN); math.Float64bits(got) != math.Float64bits(negativeNaN)^uint64(1<<63) {
		t.Fatalf("negative NaN normalization bits = %#016x", math.Float64bits(got))
	}
}

func TestPositionDelta4FEA70AllFaultPrefixes(t *testing.T) {
	baseline := &positionDeltaTestWorld4FEA70{pointX: 9, objectX: 1, pointY: 8, objectY: 2}
	positionDelta4FEA70(positionDeltaObject4FEA70, positionDeltaPoint4FEA70, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &positionDeltaTestWorld4FEA70{
				faultAt: faultAt,
				pointX:  9, objectX: 1,
				pointY: 8, objectY: 2,
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				positionDelta4FEA70(positionDeltaObject4FEA70, positionDeltaPoint4FEA70, w.hooks())
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
