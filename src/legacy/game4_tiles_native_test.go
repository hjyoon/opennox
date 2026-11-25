package legacy

import (
	"bytes"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestGame4TilesNativeAllocation5040A0(t *testing.T) {
	b := game4TilesTestBegin()
	if b == nil {
		t.Fatal("C fixture allocation failed")
	}
	defer b.close()
	cases := []struct {
		x, y int32
		bits uint32
	}{
		{3, 4, 1},
		{-7, 9, 0x3f800000},
		{math.MaxInt32, math.MinInt32, 0x7fc00001},
		{16777217, -16777217, 0x800000ff},
	}
	var want []cfixture.Game4TilesAllocation
	for _, reference := range []bool{true, false} {
		cfixture.Game4TilesAllocReset()
		var live []cfixture.Game4TilesAllocation
		for i, tc := range cases {
			got := b.Alloc(tc.x, tc.y, tc.bits, reference)
			live = append(live, got)
			if got.Entry == 0 || got.Layer == 0 || got.Entry == got.Layer || !got.Links || !got.ZeroLayer {
				t.Fatalf("allocation ownership/link/zero state: %+v", got)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (got.Entry <= math.MaxUint32 || got.Layer <= math.MaxUint32) {
				t.Fatalf("native owners not high-address: %+v", got)
			}
			if got.Kind != tc.bits&255 || got.Count != uint32(i) {
				t.Fatalf("raw kind/count wrap lost: %+v", got)
			}
			if reference {
				want = append(want, got)
				continue
			}
			expected := want[i]
			expected.Entry = got.Entry
			expected.Layer = got.Layer
			if got != expected {
				t.Fatalf("public C entry differs from retained C: %+v / %+v", got, expected)
			}
		}
		runtime.GC()
		for i, expected := range live {
			expected.Count = 3
			if got := b.Existing(uint32(i)); got != expected {
				t.Fatalf("native C-owned list changed across GC: %+v / %+v", got, expected)
			}
		}
	}
}

func TestGame4TilesNativeInit504150ReferenceOrder(t *testing.T) {
	b := game4TilesTestBegin()
	if b == nil {
		t.Fatal("C fixture allocation failed")
	}
	defer b.close()
	old := game4TilesServiceFactory
	defer func() { game4TilesServiceFactory = old }()
	game4TilesServiceFactory = game4TilesTestServices
	for _, tc := range []struct {
		name               string
		shape, definitions int
		cached             int32
	}{
		{"find exact transparent", 0, 0, -1},
		{"missing case-sensitive transparent continues", 0, 1, -1},
		{"empty definition table continues", 0, 2, -1},
		{"preloaded cache is not rescanned", 0, 0, 7},
		{"live layer links and cache, cached coordinates", 1, 0, -1},
		{"empty list still restores snapshot", 2, 0, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b.Seed(tc.shape, tc.definitions, tc.cached)
			wantResult, wantCached := cfixture.Game4TilesInit(16777217, math.MinInt32, true)
			want := b.Data()
			b.Seed(tc.shape, tc.definitions, tc.cached)
			runtime.GC()
			gotResult, gotCached := cfixture.Game4TilesInit(16777217, math.MinInt32, false)
			if gotResult != wantResult || gotResult != 1 || gotCached != wantCached || !bytes.Equal(b.Data(), want) || !b.Canaries() {
				t.Fatalf("ordered C callback/state comparison failed: return %d/%d cache %d/%d", gotResult, wantResult, gotCached, wantCached)
			}
			if tc.definitions == 1 && gotCached != -1 {
				t.Fatal("missing TransparentFloor policy was silently changed")
			}
		})
	}
}

func TestFloorEdgeSelectionNativeCEntry(t *testing.T) {
	selected, edge, variation := cfixture.Game4TilesEdgeWords()
	count := thingEdgeCountCell411850()
	oldSelected, oldEdge, oldVariation, oldCount := *selected, *edge, *variation, *count
	var saved [3][60]byte
	for i, name := range []string{"First", "Bend", "Bend"} {
		record := thingEdgeBytes411850(uint32(i), 28644, 60)
		copy(saved[i][:], record)
		clear(record)
		copy(record, name)
	}
	t.Cleanup(func() {
		*selected, *edge, *variation, *count = oldSelected, oldEdge, oldVariation, oldCount
		for i := range saved {
			copy(thingEdgeBytes411850(uint32(i), 28644, 60), saved[i][:])
		}
	})
	*selected, *edge, *variation, *count = 1, 0, 42, 3
	thingEdgeStore16411850(1, 28688, 1)
	thingEdgeStore16411850(2, 28688, 3)

	if cfixture.Game4TilesSelectEdge("Bend") != 1 || *selected != 1 || *edge != 1 {
		t.Fatalf("first exact match selected = %d, edge = %d", *selected, *edge)
	}
	if nox_xxx_tileCheckByte3_544070(3) != 0 || *selected != 1 || *edge != 1 {
		t.Fatalf("out-of-range edge changed selection to %d, edge = %d", *selected, *edge)
	}
	if nox_xxx_tileCheckByte4_5440A0(1) != 0 || *variation != 42 {
		t.Fatalf("variation equal to its limit changed the state to %d", *variation)
	}
	if nox_xxx_tileCheckByte4_5440A0(2) != 1 || *variation != 2 {
		t.Fatalf("variation using its own record was rejected, state = %d", *variation)
	}
	if cfixture.Game4TilesSelectEdge("bend") != 0 || *selected != 0 || *edge != 1 {
		t.Fatalf("case-sensitive rejection selected = %d, edge = %d", *selected, *edge)
	}
	if cfixture.Game4TilesSelectEdge("nOnE") != 1 || *selected != 0 || *edge != 255 ||
		nox_xxx_tileCheckByte4_5440A0(-1) != 1 || *variation != 2 {
		t.Fatalf("NONE selected = %d, edge = %d, variation = %d", *selected, *edge, *variation)
	}
	*selected, *edge, *count = 1, math.MaxUint32, math.MaxUint32
	if sub_543FB0(nil) != -1 || nox_xxx_tileCheckByte3_544070(0) != 0 ||
		nox_xxx_tileCheckByte4_5440A0(2) != 0 || *variation != 2 {
		t.Fatal("null name or signed edge sentinel was accepted")
	}
}
