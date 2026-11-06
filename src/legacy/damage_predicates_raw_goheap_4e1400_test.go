package legacy

import (
	"bytes"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

// These are four real, separate, aligned Go allocations. Every allocation
// reachable across the C entry is pinned individually; Pinner is not recursive.
type damagePredicateRawGoGraph4E1400 struct {
	source, weapon     *server.Object
	primary, alternate *server.WandUseData
	pin                runtime.Pinner
}

func newDamagePredicateRawGoGraph4E1400(t *testing.T) *damagePredicateRawGoGraph4E1400 {
	t.Helper()
	g := &damagePredicateRawGoGraph4E1400{
		source: new(server.Object), weapon: new(server.Object),
		primary: new(server.WandUseData), alternate: new(server.WandUseData),
	}
	g.source.ObjClass, g.source.ObjSubClass = object.Class(2), object.SubClass(0xf0cafe10)
	g.weapon.ObjClass, g.weapon.ObjSubClass = object.Class(0x1000), object.SubClass(0x10000)
	g.primary.Flags, g.alternate.Flags = 0xabcdef02, 0x12345600
	g.weapon.UseData.Ptr = unsafe.Pointer(g.primary)
	for _, value := range []any{g.source, g.weapon, g.primary, g.alternate} {
		g.pin.Pin(value)
	}
	t.Cleanup(func() {
		g.pin.Unpin()
		runtime.KeepAlive(g)
	})
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, value := range []unsafe.Pointer{unsafe.Pointer(g.source), unsafe.Pointer(g.weapon), unsafe.Pointer(g.primary), unsafe.Pointer(g.alternate)} {
			if uintptr(value) <= math.MaxUint32 {
				t.Fatal("pinned Go graph must exercise native pointers above 4 GiB", value)
			}
		}
	}
	return g
}

func (g *damagePredicateRawGoGraph4E1400) snapshot() [][]byte {
	var out [][]byte
	for _, allocation := range []struct {
		base unsafe.Pointer
		size uintptr
	}{
		{unsafe.Pointer(g.source), unsafe.Sizeof(*g.source)},
		{unsafe.Pointer(g.weapon), unsafe.Sizeof(*g.weapon)},
		{unsafe.Pointer(g.primary), unsafe.Sizeof(*g.primary)},
		{unsafe.Pointer(g.alternate), unsafe.Sizeof(*g.alternate)},
	} {
		out = append(out, append([]byte(nil), unsafe.Slice((*byte)(allocation.base), allocation.size)...))
	}
	runtime.KeepAlive(g)
	return out
}

func (g *damagePredicateRawGoGraph4E1400) checkSnapshot(t *testing.T, want [][]byte) {
	t.Helper()
	got := g.snapshot()
	if len(got) != 4 || len(want) != 4 {
		t.Fatal("all four pinned Go allocations must be compared")
	}
	for i := range want {
		if !bytes.Equal(got[i], want[i]) {
			t.Fatal("raw predicate changed complete pinned allocation outside declared stimulus", i)
		}
	}
}

func damagePredicateRawRequireActualGoGraphEngine4E1400(t *testing.T) {
	t.Helper()
	if damagePredicateABI4E1400(0) != 2 || damagePredicateABI4E1400(1) != 2 ||
		damagePredicateRawAttackEngine4E1400 == nil || damagePredicateRawFriendlyEngine4E1470 == nil {
		t.Fatal("both actual raw native ABIs and engines are required")
	}
	if reflect.ValueOf(damagePredicateRawAttackEngine4E1400).Pointer() != reflect.ValueOf(damagePredicateRawAttackNative4E1400).Pointer() ||
		reflect.ValueOf(damagePredicateRawFriendlyEngine4E1470).Pointer() != reflect.ValueOf(damagePredicateRawFriendlyNative4E1470).Pointer() {
		t.Fatal("actual raw engines must remain installed without substitution")
	}
	if damagePredicateRawObserve4E1400 != nil || damagePredicateRawObserveRoot4E1400 != nil {
		t.Fatal("default execution must start with nil read and root observers")
	}
}

func TestDamagePredicatesRawDefaultPinnedGoGraph4E1400(t *testing.T) {
	damagePredicateRawRequireActualGoGraphEngine4E1400(t)
	for _, name := range []string{"source-player", "wand-flags-byte", "wand-fallback-bit", "friendly-weapon", "friendly-nil"} {
		t.Run("C/"+name, func(t *testing.T) {
			g := newDamagePredicateRawGoGraph4E1400(t)
			entry, want := 0, int32(1)
			source, weapon := unsafe.Pointer(g.source), unsafe.Pointer(g.weapon)
			switch name {
			case "source-player":
				g.source.ObjClass, weapon = object.Class(4), nil
			case "wand-flags-byte":
				// Upper Flags bytes are real unchanged payload, not a wider read.
				g.primary.Flags = 0xabcdef02
			case "wand-fallback-bit":
				g.weapon.ObjClass, g.primary.Flags = object.Class(0x1002), 0xabcdef00
			case "friendly-weapon":
				entry = 1
				g.weapon.ObjClass, g.weapon.ObjSubClass = object.Class(0x1000000), object.SubClass(0x80004000)
			case "friendly-nil":
				entry, weapon, want = 1, nil, 0
			}
			before := g.snapshot()
			damagePredicateRawTracing4E1400(true)
			defer damagePredicateRawTracing4E1400(false)
			// Both observers remain nil; the public C symbol reaches the engine.
			got := damagePredicateRawInvoke4E1400(true, entry, source, weapon)
			if got != want || damagePredicateRawObserve4E1400 != nil || damagePredicateRawObserveRoot4E1400 != nil {
				t.Fatal("actual public C predicate result and nil default observers", got, want)
			}
			damagePredicateRawCheckCTrace4E1400(t, damagePredicateRawTraces4E1400(), entry, damagePredicateRawRootPointer4E1400(entry), source, weapon, false, want)
			g.checkSnapshot(t, before)
			t.Logf("DAMAGE_PREDICATES_RAW_PINNED_GO_GRAPH_DEFAULT_C case=%s pointer-bytes=%d observers=nil whole-graph-memory=true", name, unsafe.Sizeof(uintptr(0)))
			runtime.KeepAlive(g)
		})
	}
}

type damagePredicateRawGoRead4E1400 struct {
	base  unsafe.Pointer
	kind  int
	value damagePredicateRawValue4E1400
}

type damagePredicateRawGoStimulus4E1400 struct {
	name                                                string
	entry                                               int
	weaponNil                                           bool
	class, subclass, sourceClass, sourceSubclass, flags uint32
	triggerKind, triggerOrdinal                         int
	// The stimulus writes a Go typed field. It never writes Go pointers as bits.
	mutate   func(*damagePredicateRawGoGraph4E1400)
	expected func(*damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400
}

// Input-derived metamorphic schedules are external test data, not a branch
// schedule passed into the engine. All eight outcomes intentionally remain1.
func damagePredicateRawGoStimuli4E1400() []damagePredicateRawGoStimulus4E1400 {
	word := func(base unsafe.Pointer, kind int, value uint32) damagePredicateRawGoRead4E1400 {
		return damagePredicateRawGoRead4E1400{base, kind, damagePredicateRawValue4E1400{Word: value}}
	}
	wand := func(g *damagePredicateRawGoGraph4E1400, class, flags uint32) []damagePredicateRawGoRead4E1400 {
		return []damagePredicateRawGoRead4E1400{
			word(unsafe.Pointer(g.weapon), 2, class),
			word(unsafe.Pointer(g.weapon), 3, 0x10000),
			{unsafe.Pointer(g.weapon), 4, damagePredicateRawValue4E1400{Pointer: unsafe.Pointer(g.primary)}},
			word(unsafe.Pointer(g.primary), 5, flags),
		}
	}
	return []damagePredicateRawGoStimulus4E1400{
		{
			name: "cached-weapon-class", class: 0x1002, subclass: 0x10000, flags: 0xabcdef00, triggerKind: 2, triggerOrdinal: 1,
			mutate:   func(g *damagePredicateRawGoGraph4E1400) { g.weapon.ObjClass = object.Class(0x1000) },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 { return wand(g, 0x1002, 0) },
		},
		{
			name: "cached-subclass-test", class: 0x1000, subclass: 0, flags: 0xabcdef00, triggerKind: 3, triggerOrdinal: 2,
			mutate: func(g *damagePredicateRawGoGraph4E1400) { g.weapon.ObjSubClass = object.SubClass(0x10000) },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 {
				return []damagePredicateRawGoRead4E1400{word(unsafe.Pointer(g.weapon), 2, 0x1000), word(unsafe.Pointer(g.weapon), 3, 0)}
			},
		},
		{
			name: "cached-use-pointer", class: 0x1000, subclass: 0x10000, flags: 0xabcdef02, triggerKind: 4, triggerOrdinal: 3,
			mutate:   func(g *damagePredicateRawGoGraph4E1400) { g.weapon.UseData.Ptr = unsafe.Pointer(g.alternate) },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 { return wand(g, 0x1000, 2) },
		},
		{
			name: "cached-flags-byte", class: 0x1000, subclass: 0x10000, flags: 0xabcdef02, triggerKind: 5, triggerOrdinal: 4,
			mutate:   func(g *damagePredicateRawGoGraph4E1400) { g.primary.Flags &^= 2 },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 { return wand(g, 0x1000, 2) },
		},
		{
			name: "cached-source-class", weaponNil: true, sourceClass: 2, sourceSubclass: 0xf0cafe10, triggerKind: 0, triggerOrdinal: 1,
			mutate: func(g *damagePredicateRawGoGraph4E1400) { g.source.ObjClass = 0 },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 {
				return []damagePredicateRawGoRead4E1400{word(unsafe.Pointer(g.source), 0, 2), word(unsafe.Pointer(g.source), 1, 0x10)}
			},
		},
		{
			name: "cached-source-subclass-byte", weaponNil: true, sourceClass: 2, sourceSubclass: 0xf0cafe10, triggerKind: 1, triggerOrdinal: 2,
			mutate: func(g *damagePredicateRawGoGraph4E1400) { g.source.ObjSubClass &^= 0x10 },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 {
				return []damagePredicateRawGoRead4E1400{word(unsafe.Pointer(g.source), 0, 2), word(unsafe.Pointer(g.source), 1, 0x10)}
			},
		},
		{
			name: "cached-friendly-class", entry: 1, class: 0x1000000, subclass: 0x80004000, triggerKind: 2, triggerOrdinal: 1,
			mutate: func(g *damagePredicateRawGoGraph4E1400) { g.weapon.ObjClass = 0 },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 {
				return []damagePredicateRawGoRead4E1400{word(unsafe.Pointer(g.weapon), 2, 0x1000000), word(unsafe.Pointer(g.weapon), 3, 0x80004000)}
			},
		},
		{
			name: "cached-friendly-subclass-DWORD", entry: 1, class: 0x1000000, subclass: 0x80004000, triggerKind: 3, triggerOrdinal: 2,
			mutate: func(g *damagePredicateRawGoGraph4E1400) { g.weapon.ObjSubClass = 0 },
			expected: func(g *damagePredicateRawGoGraph4E1400) []damagePredicateRawGoRead4E1400 {
				return []damagePredicateRawGoRead4E1400{word(unsafe.Pointer(g.weapon), 2, 0x1000000), word(unsafe.Pointer(g.weapon), 3, 0x80004000)}
			},
		},
	}
}

func TestDamagePredicatesRawPinnedGoGraphGC4E1400(t *testing.T) {
	damagePredicateRawRequireActualGoGraphEngine4E1400(t)
	for _, public := range []bool{false, true} {
		backend := "Go"
		if public {
			backend = "C"
		}
		backendReads, backendGC, backendStimuli := 0, 0, 0
		for _, row := range damagePredicateRawGoStimuli4E1400() {
			t.Run(backend+"/"+row.name, func(t *testing.T) {
				g := newDamagePredicateRawGoGraph4E1400(t)
				g.source.ObjClass, g.source.ObjSubClass = object.Class(row.sourceClass), object.SubClass(row.sourceSubclass)
				g.weapon.ObjClass, g.weapon.ObjSubClass = object.Class(row.class), object.SubClass(row.subclass)
				g.primary.Flags = row.flags
				source, weapon := unsafe.Pointer(g.source), unsafe.Pointer(g.weapon)
				if row.weaponNil {
					weapon = nil
				}
				before, expectedReads := g.snapshot(), row.expected(g)
				// Compose expected whole bytes from the sole declared input write,
				// then restore all typed fields without copying pointer bits.
				row.mutate(g)
				expectedMemory := g.snapshot()
				g.source.ObjClass, g.source.ObjSubClass = object.Class(row.sourceClass), object.SubClass(row.sourceSubclass)
				g.weapon.ObjClass, g.weapon.ObjSubClass = object.Class(row.class), object.SubClass(row.subclass)
				g.weapon.UseData.Ptr, g.primary.Flags = unsafe.Pointer(g.primary), row.flags
				g.checkSnapshot(t, before)
				var reads []damagePredicateRawGoRead4E1400
				var pending *damagePredicateRawGoRead4E1400
				var roots []damagePredicateRawRoot4E1400
				gc, stimuli := 0, 0
				oldRead, oldRoot := damagePredicateRawObserve4E1400, damagePredicateRawObserveRoot4E1400
				defer func() {
					damagePredicateRawObserve4E1400, damagePredicateRawObserveRoot4E1400 = oldRead, oldRoot
					damagePredicateRawTracing4E1400(false)
				}()
				damagePredicateRawObserve4E1400 = func(complete bool, base unsafe.Pointer, kind int, value damagePredicateRawValue4E1400) {
					if !complete {
						if pending != nil {
							t.Fatal("nested actual raw field read")
						}
						pending = &damagePredicateRawGoRead4E1400{base: base, kind: kind}
						runtime.GC()
						gc++
						return
					}
					if pending == nil || pending.base != base || pending.kind != kind {
						t.Fatal("actual field read completion mismatch")
					}
					reads = append(reads, damagePredicateRawGoRead4E1400{base, kind, value})
					pending = nil
					if kind == row.triggerKind && len(reads) == row.triggerOrdinal {
						row.mutate(g)
						stimuli++
					}
				}
				damagePredicateRawObserveRoot4E1400 = func(complete bool, entry int, args [2]unsafe.Pointer, result int32) {
					roots = append(roots, damagePredicateRawRoot4E1400{complete, entry, args, result})
				}
				if public {
					damagePredicateRawTracing4E1400(true)
				}
				got := damagePredicateRawInvoke4E1400(public, row.entry, source, weapon)
				if got != 1 || pending != nil || len(reads) != len(expectedReads) || gc != len(expectedReads) || stimuli != 1 {
					t.Fatal("actual result, complete reads, in-read GC and single declared stimulus", got, pending, reads, gc, stimuli)
				}
				for i := range expectedReads {
					if reads[i] != expectedReads[i] {
						t.Fatal("actual read order, cached full pointer or BYTE/DWORD value", i, reads[i], expectedReads[i])
					}
				}
				args := damagePredicateRawArgs4E1400(row.entry, source, weapon)
				if len(roots) != 2 || roots[0].Complete || roots[0].Entry != row.entry || roots[0].Args != args ||
					!roots[1].Complete || roots[1].Entry != row.entry || roots[1].Args != args || roots[1].Result != 1 {
					t.Fatal("actual root entry/return and full native arguments", roots)
				}
				if public {
					damagePredicateRawCheckCTrace4E1400(t, damagePredicateRawTraces4E1400(), row.entry, damagePredicateRawRootPointer4E1400(row.entry), source, weapon, false, 1)
				}
				g.checkSnapshot(t, expectedMemory)
				backendReads, backendGC, backendStimuli = backendReads+len(reads), backendGC+gc, backendStimuli+stimuli
				t.Logf("DAMAGE_PREDICATES_RAW_PINNED_GO_GRAPH_GC backend=%s case=%s reads=%d forced-GC=%d stimuli=%d actual-public-C=%t whole-graph-memory=true", backend, row.name, len(reads), gc, stimuli, public)
				runtime.KeepAlive(g)
			})
		}
		if backendReads != 22 || backendGC != 22 || backendStimuli != 8 {
			t.Fatal("all eight Go graph cases per actual backend", backend, backendReads, backendGC, backendStimuli)
		}
	}
}
