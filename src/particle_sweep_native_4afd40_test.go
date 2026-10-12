package opennox

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"slices"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func particleSweepOwner4AFD40(t *testing.T) *partFXes {
	t.Helper()
	handles.Init()
	pfx := new(partFXes)
	pfx.Init(nil)
	t.Cleanup(pfx.Free)
	return pfx
}

// Labels are fixture inputs, not allocator addresses or map iteration order.
// Callbacks observe the real Draw contract; they do not choose its traversal.
func particleSweepTrace4AFD40(pfx *partFXes, label int, trace *[]string, callback func(string, int)) *particleFx {
	p := pfx.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, 64, color.White)
	p.field_68 = label
	p.drawFunc = func(p *particleFx) {
		*trace = append(*trace, fmt.Sprintf("D%d", p.field_68))
		if callback != nil {
			callback("D", p.field_68)
		}
	}
	p.updateFunc = func(p *particleFx) bool {
		*trace = append(*trace, fmt.Sprintf("U%d", p.field_68))
		if callback != nil {
			callback("U", p.field_68)
		}
		return true
	}
	return p
}

func particleSweepLiteralTrace4AFD40(labels ...int) []string {
	var result []string
	for _, label := range labels {
		result = append(result, fmt.Sprintf("D%d", label), fmt.Sprintf("U%d", label))
	}
	return result
}

func TestParticleSweepOriginal4AFD40Order(t *testing.T) {
	for _, count := range []int{6, 17, 64} {
		t.Run(fmt.Sprintf("slots-%d", count), func(t *testing.T) {
			pfx := particleSweepOwner4AFD40(t)
			other := new(partFXes)
			other.Init(nil)
			t.Cleanup(other.Free)
			var trace, otherTrace []string
			labels := make([]int, count)
			for i := range labels {
				labels[i] = i
				particleSweepTrace4AFD40(pfx, i, &trace, nil)
				particleSweepTrace4AFD40(other, 1000+i, &otherTrace, nil)
				for j := 0; j < 16; j++ {
					_ = handles.New()
				}
			}
			want := particleSweepLiteralTrace4AFD40(labels...)
			for frame := 0; frame < 16; frame++ {
				trace = nil
				pfx.Draw()
				if !slices.Equal(trace, want) || len(otherTrace) != 0 {
					t.Fatalf("frame=%d original low-slot draw/update order: got=%v want=%v other=%v", frame, trace, want, otherTrace)
				}
			}
		})
	}
}

func TestParticleSweepOriginal4AFD40Dynamic(t *testing.T) {
	cases := []struct {
		name   string
		first  []string
		second []int
	}{
		{"append-from-draw", particleSweepLiteralTrace4AFD40(0, 1, 2, 3, 4), []int{0, 1, 2, 3, 4}},
		{"append-from-update", particleSweepLiteralTrace4AFD40(0, 1, 2, 3, 4), []int{0, 1, 2, 3, 4}},
		{"reuse-before-current", particleSweepLiteralTrace4AFD40(0, 1, 2, 3), []int{4, 1, 2, 3}},
		{"activate-future-hole", particleSweepLiteralTrace4AFD40(0, 1, 4, 3), []int{0, 1, 4, 3}},
		{"activate-future-hole-gc", particleSweepLiteralTrace4AFD40(0, 1, 4, 3), []int{0, 1, 4, 3}},
		{"replace-future", particleSweepLiteralTrace4AFD40(0, 1, 2, 4), []int{0, 1, 2, 4}},
		{"remove-future", particleSweepLiteralTrace4AFD40(0, 1, 2), []int{0, 1, 2}},
		{"replace-current", []string{"D0", "U0", "D1", "U4", "D2", "U2", "D3", "U3"}, []int{0, 4, 2, 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pfx := particleSweepOwner4AFD40(t)
			var trace []string
			var slots [4]*particleFx
			mutated := false
			change := func(phase string, label int) {
				when := "D"
				if tc.name == "append-from-update" {
					when = "U"
				}
				if phase != when || label != 1 || mutated {
					return
				}
				mutated = true
				switch tc.name {
				case "reuse-before-current":
					slots[0].free()
				case "replace-future", "remove-future":
					slots[3].free()
				case "replace-current":
					slots[1].free()
				case "activate-future-hole-gc":
					runtime.GC()
				}
				if tc.name != "remove-future" {
					particleSweepTrace4AFD40(pfx, 4, &trace, nil)
				}
			}
			for i := range slots {
				slots[i] = particleSweepTrace4AFD40(pfx, i, &trace, change)
			}
			if tc.name == "activate-future-hole" || tc.name == "activate-future-hole-gc" {
				slots[2].free()
			}
			pfx.Draw()
			if !mutated || !slices.Equal(trace, tc.first) {
				t.Fatalf("dynamic original forward cursor: got=%v want=%v mutation=%t", trace, tc.first, mutated)
			}
			trace = nil
			pfx.Draw()
			if want := particleSweepLiteralTrace4AFD40(tc.second...); !slices.Equal(trace, want) {
				t.Fatalf("next-frame original bitmap: got=%v want=%v", trace, want)
			}
		})
	}
}

func TestParticleSweepOriginal4AFD40GrowthDuringCallback(t *testing.T) {
	for _, count := range []int{512, 513, 1025} {
		t.Run(fmt.Sprintf("slots-%d", count), func(t *testing.T) {
			pfx := particleSweepOwner4AFD40(t)
			var trace []string
			created := false
			for i := 0; i < count; i++ {
				particleSweepTrace4AFD40(pfx, i, &trace, func(phase string, label int) {
					if phase == "U" && label == count-1 && !created {
						created = true
						for j := count; j < count+3; j++ {
							particleSweepTrace4AFD40(pfx, j, &trace, nil)
						}
					}
				})
			}
			labels := make([]int, count+3)
			for i := range labels {
				labels[i] = i
			}
			pfx.Draw()
			if want := particleSweepLiteralTrace4AFD40(labels...); !created || !slices.Equal(trace, want) || len(pfx.byHandle) != count+3 {
				t.Fatalf("current capacity was not reread after callback: records=%d trace=%d original=%d", len(pfx.byHandle), len(trace), len(want))
			}
			t.Logf("PARTICLE SWEEP GROWTH: %d initial + 3 callback-created records traversed in one real Draw", count)
		})
	}
}

func TestParticleSweepOriginal4AFD40HiddenAndNil(t *testing.T) {
	for bits := 0; bits < 8; bits++ {
		t.Run(fmt.Sprintf("mask-%d", bits), func(t *testing.T) {
			pfx := particleSweepOwner4AFD40(t)
			var trace, want []string
			p := particleSweepTrace4AFD40(pfx, 0, &trace, nil)
			if bits&1 != 0 {
				p.flags = 8
			}
			if bits&2 != 0 {
				p.drawFunc = nil
			}
			if bits&4 != 0 {
				p.updateFunc = nil
			}
			if bits&3 == 0 {
				want = append(want, "D0")
			}
			if bits&4 == 0 {
				want = append(want, "U0")
			}
			pfx.Draw()
			if !slices.Equal(trace, want) || pfx.byHandle[p.chnd] != p {
				t.Fatalf("original hidden/nil dispatch: got=%v want=%v", trace, want)
			}
		})
	}
}

func TestParticleSweepOriginal4AFD40TypesAfterParticles(t *testing.T) {
	for _, interval := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("interval-%d", interval), func(t *testing.T) {
			pfx := particleSweepOwner4AFD40(t)
			var trace []string
			p := particleSweepTrace4AFD40(pfx, 0, &trace, nil)
			emitted := 0
			if pfx.registerParticleFx(func(it *particlefxType) bool {
				if it.field1 != 1 {
					t.Fatalf("original callback observes field1=1, not decremented: %d", it.field1)
				}
				trace = append(trace, "T")
				emitted++
				particleSweepTrace4AFD40(pfx, emitted, &trace, nil)
				return true
			}, p, interval, 3) != 1 {
				t.Fatal("actual type registration failed")
			}
			for frame := 0; frame < 2; frame++ {
				trace = nil
				pfx.Draw()
				want := []string{"D0", "U0"}
				if interval == 1 {
					if frame == 1 {
						want = append(want, "D1", "U1")
					}
					want = append(want, "T")
				}
				if !slices.Equal(trace, want) {
					t.Fatalf("type table must run after the dynamic particle sweep, frame=%d got=%v want=%v", frame, trace, want)
				}
			}
		})
	}
}
