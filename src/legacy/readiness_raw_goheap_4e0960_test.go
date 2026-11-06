package legacy

import (
	"bytes"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type readinessRawGoGraph4E0960 struct {
	item                       *server.Object
	attrs, alternate           *server.ModifierInitData
	first, second, replacement *server.ModifierEff
}

func newReadinessRawGoGraph4E0960(t *testing.T) *readinessRawGoGraph4E0960 {
	t.Helper()
	g := &readinessRawGoGraph4E0960{
		item: new(server.Object), attrs: new(server.ModifierInitData), alternate: new(server.ModifierInitData),
		first: new(server.ModifierEff), second: new(server.ModifierEff), replacement: new(server.ModifierEff),
	}
	g.first.Attack40 = server.ModifierEffFnc{Fnc: ReadinessEffectPointer4E0960(), Val: 37}
	g.second.Attack40 = server.ModifierEffFnc{Fnc: ReadinessEffectPointer4E0960(), Val: 91}
	// This replacement deliberately lacks a matching callback.
	g.replacement.Attack40.Val = math.MinInt32
	g.attrs.Modifiers = [4]*server.ModifierEff{g.second, g.second, g.first, g.second}
	g.alternate.Modifiers[2] = g.second
	g.item.ObjClass, g.item.InitData = object.ClassWand, unsafe.Pointer(g.attrs)
	var pin runtime.Pinner
	for _, value := range []any{g.item, g.attrs, g.alternate, g.first, g.second, g.replacement} {
		pin.Pin(value)
	}
	t.Cleanup(pin.Unpin)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, value := range []unsafe.Pointer{unsafe.Pointer(g.item), unsafe.Pointer(g.attrs), unsafe.Pointer(g.alternate), unsafe.Pointer(g.first), unsafe.Pointer(g.second), unsafe.Pointer(g.replacement)} {
			if uintptr(value) <= math.MaxUint32 {
				t.Fatal("actual pinned Go graph must exercise native addresses above 4 GiB", value)
			}
		}
	}
	return g
}

func (g *readinessRawGoGraph4E0960) snapshot() [][]byte {
	var out [][]byte
	for _, field := range []struct {
		base unsafe.Pointer
		size uintptr
	}{
		{unsafe.Pointer(g.item), unsafe.Sizeof(*g.item)},
		{unsafe.Pointer(g.attrs), unsafe.Sizeof(*g.attrs)},
		{unsafe.Pointer(g.alternate), unsafe.Sizeof(*g.alternate)},
		{unsafe.Pointer(g.first), unsafe.Sizeof(*g.first)},
		{unsafe.Pointer(g.second), unsafe.Sizeof(*g.second)},
		{unsafe.Pointer(g.replacement), unsafe.Sizeof(*g.replacement)},
	} {
		out = append(out, append([]byte(nil), unsafe.Slice((*byte)(field.base), field.size)...))
	}
	return out
}

func (g *readinessRawGoGraph4E0960) checkSnapshot(t *testing.T, want [][]byte) {
	t.Helper()
	got := g.snapshot()
	if len(got) != len(want) {
		t.Fatal("complete Go graph snapshot count")
	}
	for i := range want {
		if !bytes.Equal(got[i], want[i]) {
			t.Fatal("actual raw entry changed whole pinned Go graph outside declared stimuli", i)
		}
	}
}

func TestReadinessRawDefaultPinnedGoGraph4E0960(t *testing.T) {
	if readinessRawABI4E0960() != 2 || readinessRawEngine4E0960 == nil {
		t.Fatal("native raw Readiness ABI/engine required")
	}
	if readinessRawObserve4E0960 != nil || readinessRawObserveRoot4E0960 != nil {
		t.Fatal("default-service integration requires ordinary nil observers")
	}
	for _, name := range []string{"first-zero", "second-negative", "full-width-nonmatch"} {
		t.Run("C/"+name, func(t *testing.T) {
			g := newReadinessRawGoGraph4E0960(t)
			var want int32
			collision := false
			switch name {
			case "first-zero":
				g.first.Attack40.Val = 0
			case "second-negative":
				g.attrs.Modifiers[2], g.second.Attack40.Val = nil, math.MinInt32
				want = math.MinInt32
			case "full-width-nonmatch":
				g.first.Attack40.Fnc = readinessRawCollisionCallback4E0960()
				g.attrs.Modifiers[3] = nil
				if unsafe.Sizeof(uintptr(0)) == 8 {
					real, token := uintptr(ReadinessEffectPointer4E0960()), uintptr(g.first.Attack40.Fnc)
					if real <= math.MaxUint32 || token <= math.MaxUint32 || real == token || uint32(real) != uint32(token) {
						t.Fatal("same low32 and unequal native callback identities required", real, token)
					}
					collision = true
				}
			}
			before := g.snapshot()
			// Actual public C entry, with no read/root observer or substitute engine.
			if got := readinessRawPublic4E0960(unsafe.Pointer(g.item)); got != want {
				t.Fatal("actual public C raw Readiness on pinned Go graph", got, want)
			}
			g.checkSnapshot(t, before)
			t.Logf("READINESS_RAW_PINNED_GO_GRAPH_DEFAULT_C case=%s pointer-bytes=%d low32-collision=%t observers=nil whole-graph-memory=true", name, unsafe.Sizeof(uintptr(0)), collision)
			runtime.KeepAlive(g)
		})
	}
}

type readinessRawGoGraphRead4E0960 struct {
	base        unsafe.Pointer
	kind, index int
	value       readinessRawValue4E0960
}

func TestReadinessRawPinnedGoGraphGC4E0960(t *testing.T) {
	if readinessRawABI4E0960() != 2 || readinessRawEngine4E0960 == nil {
		t.Fatal("native raw Readiness ABI/engine required")
	}
	if readinessRawObserve4E0960 != nil || readinessRawObserveRoot4E0960 != nil {
		t.Fatal("fresh native read/root observers required")
	}
	for _, public := range []bool{false, true} {
		backend := "Go"
		if public {
			backend = "C"
		}
		for _, name := range []string{"cached-table-reload", "cached-first-slot", "cached-strength"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				g := newReadinessRawGoGraph4E0960(t)
				selected, want := g.replacement, int32(math.MinInt32)
				if name == "cached-strength" {
					g.first.Attack40.Val = 0
					selected, want = g.first, 0
				}
				// Build expected whole-memory bytes from the declared stimulus
				// writes before invoking the engine; then restore initial inputs.
				if name == "cached-table-reload" {
					g.item.InitData = unsafe.Pointer(g.alternate)
				}
				if name == "cached-strength" {
					g.first.Attack40.Val = 37
				} else {
					g.attrs.Modifiers[2] = g.replacement
				}
				expectedMemory := g.snapshot()
				g.item.InitData, g.attrs.Modifiers[2] = unsafe.Pointer(g.attrs), g.first
				if name == "cached-strength" {
					g.first.Attack40.Val = 0
				}
				expectedReads := []readinessRawGoGraphRead4E0960{
					{unsafe.Pointer(g.item), 0, 0, readinessRawValue4E0960{Pointer: unsafe.Pointer(g.attrs)}},
					{unsafe.Pointer(g.item), 1, 0, readinessRawValue4E0960{Word: uint32(object.ClassWand)}},
					{unsafe.Pointer(g.attrs), 2, 2, readinessRawValue4E0960{Pointer: unsafe.Pointer(g.first)}},
					{unsafe.Pointer(g.first), 3, 0, readinessRawValue4E0960{Pointer: ReadinessEffectPointer4E0960()}},
					{unsafe.Pointer(g.attrs), 2, 2, readinessRawValue4E0960{Pointer: unsafe.Pointer(selected)}},
					{unsafe.Pointer(selected), 4, 0, readinessRawValue4E0960{Word: uint32(want)}},
				}
				var reads []readinessRawGoGraphRead4E0960
				var pending *readinessRawGoGraphRead4E0960
				gc, stimuli := 0, 0
				old := readinessRawObserve4E0960
				defer func() { readinessRawObserve4E0960 = old; readinessRawTracing4E0960(false) }()
				readinessRawObserve4E0960 = func(complete bool, base unsafe.Pointer, kind, index int, value readinessRawValue4E0960) {
					if !complete {
						if pending != nil {
							t.Fatal("nested actual raw typed read")
						}
						pending = &readinessRawGoGraphRead4E0960{base: base, kind: kind, index: index}
						runtime.GC()
						gc++
						return
					}
					if pending == nil || pending.base != base || pending.kind != kind || pending.index != index {
						t.Fatal("actual raw read before/after pair changed")
					}
					reads = append(reads, readinessRawGoGraphRead4E0960{base, kind, index, value})
					pending = nil
					if name == "cached-table-reload" && kind == 0 {
						g.item.InitData = unsafe.Pointer(g.alternate)
						stimuli++
					}
					if (name == "cached-table-reload" && kind == 3) || (name == "cached-first-slot" && kind == 2 && len(reads) == 3) {
						g.attrs.Modifiers[2] = g.replacement
						stimuli++
					}
					if name == "cached-strength" && kind == 4 {
						g.first.Attack40.Val = 37
						stimuli++
					}
				}
				if public {
					readinessRawTracing4E0960(true)
				}
				got := readinessRawInvoke4E0960(public, unsafe.Pointer(g.item))
				if got != want || pending != nil || gc != 6 || len(reads) != len(expectedReads) {
					t.Fatal("actual native result/read completion and forced in-call GC", got, want, pending, gc, reads)
				}
				for i := range reads {
					if reads[i] != expectedReads[i] {
						t.Fatal("actual raw Go-heap read order, cached values or reloads", i, reads[i], expectedReads[i])
					}
				}
				wantStimuli := 1
				if name == "cached-table-reload" {
					wantStimuli = 2
				}
				if stimuli != wantStimuli {
					t.Fatal("all declared Go-heap after-read stimuli must execute", stimuli, wantStimuli)
				}
				if public {
					trace := readinessRawTraces4E0960()
					root, item := readinessRawRootPointer4E0960(), uintptr(unsafe.Pointer(g.item))
					if len(trace) != 2 || trace[0].Function != root || trace[1].Function != root || trace[0].Item != item || trace[1].Item != item || trace[0].Phase != 0 || trace[1].Phase != 1 || trace[1].Result != want {
						t.Fatal("actual public C full-width argument, root and signed return transport", trace)
					}
				}
				g.checkSnapshot(t, expectedMemory)
				t.Logf("READINESS_RAW_PINNED_GO_GRAPH_GC backend=%s case=%s reads=%d forced-GC=%d stimuli=%d actual-public-C=%t whole-graph-memory=true", backend, name, len(reads), gc, stimuli, public)
				runtime.KeepAlive(g)
			})
		}
	}
}
