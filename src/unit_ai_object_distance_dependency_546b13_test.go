package opennox

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestAIObjectDistanceDependency546B13RetainedSurfaceBits(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		x, y                          uint32
		unitKind, unitR, unitW, unitH uint32
		targetKind, targetR           uint32
		want                          uint64
	}{
		{name: "above-one", x: 0x3f800000, y: 0x39000000, want: 0x3ff0000001ffffff},
		{name: "source-circle-retains-tail", x: 0x3f800000, y: 0x39000000, unitKind: 2, unitR: 0x3f000000, want: 0x3fe0000003fffffe},
		{name: "both-circles-retain-tail", x: 0x3f800000, y: 0x39000000, unitKind: 2, unitR: 0x3e800000, targetKind: 2, targetR: 0x3e800000, want: 0x3fe0000003fffffe},
		{name: "box-width-retains-tail", x: 0x3f800000, y: 0x39000000, unitKind: 3, unitW: 0x3f800000, unitH: 0x3f000000, want: 0x3fe0000003fffffe},
		{name: "exact-three-four-five", x: 0x40400000, y: 0x40800000, want: 0x4014000000000000},
		{name: "source-then-target-chop", x: 0x3f800000, unitKind: 2, unitR: 0x7f7fffff, targetKind: 2, targetR: 0xff7fffff, want: 0x44a0000000000000},
		{name: "reverse-shapes-chop-and-clamp", x: 0x3f800000, unitKind: 2, unitR: 0xff7fffff, targetKind: 2, targetR: 0x7f7fffff, want: 0x3f847ae140000000},
		{name: "unordered-width-selects-height", x: 0x41200000, unitKind: 3, unitW: 0x7fc12345, unitH: 0x40800000, want: 0x4020000000000000},
		{name: "unordered-height-clamps", x: 0x41200000, unitKind: 3, unitW: 0x40800000, unitH: 0x7fc12345, want: 0x3f847ae140000000},
		{name: "position-unordered-clamps", x: 0x7fc12345, want: 0x3f847ae140000000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, target := server.Object{}, server.Object{PosVec: types.Pointf{X: math.Float32frombits(tc.x), Y: math.Float32frombits(tc.y)}}
			unit.Shape.Kind, target.Shape.Kind = server.ShapeKind(tc.unitKind), server.ShapeKind(tc.targetKind)
			unit.Shape.Circle.R, target.Shape.Circle.R = math.Float32frombits(tc.unitR), math.Float32frombits(tc.targetR)
			unit.Shape.Box.W, unit.Shape.Box.H = math.Float32frombits(tc.unitW), math.Float32frombits(tc.unitH)
			unitWords, targetWords := aiObjectDistanceInputWords546B13(&unit), aiObjectDistanceInputWords546B13(&target)
			got := aiDependencyObjectSurfaceDistance546B13(&unit, &target)
			if math.Float64bits(got) != tc.want {
				t.Fatalf("retained surface bits=%016x want=%016x", math.Float64bits(got), tc.want)
			}
			if aiObjectDistanceInputWords546B13(&unit) != unitWords || aiObjectDistanceInputWords546B13(&target) != targetWords {
				t.Fatal("surface calculation changed raw object input")
			}
		})
	}
}

func TestAIObjectDistanceDependency546B13OriginalHeightStore(t *testing.T) {
	// Literal m32real round-toward-zero outcomes. In particular 3/2 minimum
	// subnormal units store one, not nearest-even's two; sign is retained.
	for _, tc := range []struct {
		height, want uint32
	}{
		{0, 0}, {0x80000000, 0x80000000},
		{1, 0}, {3, 1}, {5, 2}, {7, 3},
		{0x80000001, 0x80000000}, {0x80000003, 0x80000001}, {0x80000005, 0x80000002},
		{0x007fffff, 0x003fffff}, {0x00800001, 0x00400000},
		{0x807fffff, 0x803fffff}, {0x80800001, 0x80400000},
		{0x3f800000, 0x3f000000}, {0xbf800000, 0xbf000000},
		{0x7f7fffff, 0x7effffff}, {0x7f800000, 0x7f800000}, {0xff800000, 0xff800000},
	} {
		t.Run(fmt.Sprintf("%08x", tc.height), func(t *testing.T) {
			got := aiDependencyBoxHeightHalf546B13(math.Float32frombits(tc.height))
			want := float64(math.Float32frombits(tc.want))
			if math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("height half=%016x want=%016x", math.Float64bits(got), math.Float64bits(want))
			}
			if tc.height&0x7f800000 != 0x7f800000 {
				shape := server.Shape{Kind: server.ShapeKindBox}
				shape.Box.W, shape.Box.H = math.Float32frombits(0xff7fffff), math.Float32frombits(tc.height)
				if math.Float64bits(aiObjectDistanceReferenceExtent546B13(&shape)) != math.Float64bits(want) {
					t.Fatal("independent finite height-store reference differs from literal")
				}
			}
		})
	}
}

// aiObjectDistanceReferenceExtent546B13 uses integer exponent/significand
// truncation for finite height stores, independently of production Nextafter32.
// The deterministic model supplies finite words only; native/literal cases
// separately cover infinity and unordered inputs.
func aiObjectDistanceReferenceExtent546B13(shape *server.Shape) float64 {
	if shape.Kind == server.ShapeKindCircle {
		return float64(shape.Circle.R)
	}
	if shape.Kind != server.ShapeKindBox {
		return 0
	}
	width := float64(shape.Box.W) / 2
	bits := math.Float32bits(shape.Box.H)
	magnitude, sign := bits&0x7fffffff, bits&0x80000000
	var half uint32
	if magnitude < 0x01000000 {
		if magnitude < 0x00800000 {
			half = magnitude >> 1
		} else {
			half = (0x00800000 | (magnitude & 0x007fffff)) >> 1
		}
	} else {
		half = magnitude - 0x00800000
	}
	height := float64(math.Float32frombits(sign | half))
	if width > height {
		return width
	}
	return height
}

func TestAIObjectDistanceDependency546B13IndependentSurfaceModel(t *testing.T) {
	for _, shapes := range [][2]server.ShapeKind{{0, 0}, {2, 2}, {3, 3}, {2, 3}, {3, 2}, {0x60000002, 3}} {
		t.Run(fmt.Sprintf("%08x-%08x", shapes[0], shapes[1]), func(t *testing.T) {
			word := uint32(0x546b1301)
			next := func() float32 {
				word = 1664525*word + 1013904223
				bits := word
				if bits&0x7f800000 == 0x7f800000 {
					bits ^= 0x00800000
				}
				return math.Float32frombits(bits)
			}
			for i := 0; i < 256; i++ {
				unit := server.Object{PosVec: types.Pointf{X: next(), Y: next()}}
				target := server.Object{PosVec: types.Pointf{X: next(), Y: next()}}
				for index, obj := range []*server.Object{&unit, &target} {
					obj.Shape.Kind = shapes[index]
					obj.Shape.Circle.R, obj.Shape.Box.W, obj.Shape.Box.H = next(), next(), next()
				}
				dx := aiLocationReferenceChop53_546B63(0).Sub(aiLocationReferenceChop53_546B63(float64(unit.PosVec.X)), aiLocationReferenceChop53_546B63(float64(target.PosVec.X)))
				dy := aiLocationReferenceChop53_546B63(0).Sub(aiLocationReferenceChop53_546B63(float64(unit.PosVec.Y)), aiLocationReferenceChop53_546B63(float64(target.PosVec.Y)))
				ySquare := aiLocationReferenceChop53_546B63(0).Mul(dy, dy)
				xSquare := aiLocationReferenceChop53_546B63(0).Mul(dx, dx)
				sum := aiLocationReferenceChop53_546B63(0).Add(ySquare, xSquare)
				result := aiLocationReferenceSqrtChop53_546B63(sum)
				result.Sub(result, aiLocationReferenceChop53_546B63(aiObjectDistanceReferenceExtent546B13(&unit.Shape)))
				result.Sub(result, aiLocationReferenceChop53_546B63(aiObjectDistanceReferenceExtent546B13(&target.Shape)))
				minimum := aiLocationReferenceChop53_546B63(float64(math.Float32frombits(0x3c23d70a)))
				if result.Cmp(minimum) < 0 {
					result = minimum
				}
				want, accuracy := result.Float64()
				got := aiDependencyObjectSurfaceDistance546B13(&unit, &target)
				if accuracy != big.Exact || math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("surface model set %d differs: got=%016x want=%016x accuracy=%v", i, math.Float64bits(got), math.Float64bits(want), accuracy)
				}
			}
		})
	}
}

func TestAIObjectDistanceDependency546B13MissingInputFaultPrefix(t *testing.T) {
	s, unit, target, health := aiAliveDependencyNative546A70(t)
	unit.PosVec, target.PosVec = types.Pointf{X: 1, Y: 2}, types.Pointf{X: 3, Y: 4}
	for _, closer := range []bool{false, true} {
		label := "far"
		if closer {
			label = "close"
		}
		for _, tc := range []struct {
			name                               string
			missingUnit, missingSlot, noTarget bool
			wantFault                          bool
		}{
			{"nil-slot", false, true, false, true},
			{"nil-unit-with-target", true, false, false, true},
			{"nil-unit-and-target", true, false, true, false},
			{"nil-target", false, false, true, false},
			{"nil-both-inputs", true, true, false, true},
		} {
			t.Run(label+"/"+tc.name, func(t *testing.T) {
				high := uintptr(0x12345678)
				high <<= 32
				slot := &server.AIStackItem{Action: 49, Args: [4]uintptr{high | 0x7fc12345, high | 7, uintptr(unsafe.Pointer(target)), high | 9}, Field5: 17}
				if closer {
					slot.Action = 50
				}
				if tc.noTarget {
					slot.Args[2] = 0
				}
				selectedUnit, selectedSlot := unit, slot
				if tc.missingUnit {
					selectedUnit = nil
				}
				if tc.missingSlot {
					selectedSlot = nil
				}
				beforeSlot, beforeUpdate, beforeHealth := *slot, *unit.UpdateDataMonster(), *health
				unitWords, targetWords := aiObjectDistanceInputWords546B13(unit), aiObjectDistanceInputWords546B13(target)
				healthPointer := target.HealthData
				s.AI.StackChanged = false
				var fault any
				var got bool
				func() {
					defer func() { fault = recover() }()
					got = aiDependencyObjectDistance546B13(selectedUnit, selectedSlot, closer)
				}()
				if (fault != nil) != tc.wantFault || got {
					t.Fatalf("input prefix fault=%v wantFault=%t result=%t", fault, tc.wantFault, got)
				}
				if *slot != beforeSlot || *unit.UpdateDataMonster() != beforeUpdate || *health != beforeHealth || target.HealthData != healthPointer || s.AI.StackChanged || aiObjectDistanceInputWords546B13(unit) != unitWords || aiObjectDistanceInputWords546B13(target) != targetWords {
					t.Fatal("input prefix mutated native object, health, update or cached slot")
				}
			})
		}
	}
}

func TestAIObjectDistanceDependency546B13SurfaceMissingInput(t *testing.T) {
	_, unit, target, _ := aiAliveDependencyNative546A70(t)
	for _, missing := range []string{"unit", "target"} {
		t.Run(missing, func(t *testing.T) {
			selectedUnit, selectedTarget := unit, target
			if missing == "unit" {
				selectedUnit = nil
			} else {
				selectedTarget = nil
			}
			unitWords, targetWords := aiObjectDistanceInputWords546B13(unit), aiObjectDistanceInputWords546B13(target)
			var fault any
			func() {
				defer func() { fault = recover() }()
				aiDependencyObjectSurfaceDistance546B13(selectedUnit, selectedTarget)
			}()
			if fault == nil || aiObjectDistanceInputWords546B13(unit) != unitWords || aiObjectDistanceInputWords546B13(target) != targetWords {
				t.Fatalf("surface prefix fault=%v or changed object inputs", fault)
			}
		})
	}
}
