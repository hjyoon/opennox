package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestModifierDefendPredicatesPinnedObjects4E03D0(t *testing.T) {
	root, item, target, owner := new(server.Object), new(server.Object), new(server.Object), new(server.Object)
	effect, attrs := new(server.ModifierEff), new(server.ModifierInitData)
	root.InvFirstItem, item.InitData = item, unsafe.Pointer(attrs)
	item.ObjFlags, item.ObjClass = 0x100, 0x1000
	target.ObjOwner, target.ObjClass = owner, 4
	attrs.Modifiers[2] = effect
	var pin runtime.Pinner
	defer pin.Unpin()
	for _, value := range []any{root, item, target, owner, effect, attrs} {
		pin.Pin(value)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(root), unsafe.Pointer(item), unsafe.Pointer(target), unsafe.Pointer(owner), unsafe.Pointer(effect), unsafe.Pointer(attrs)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatal("pinned Go graph must exercise native addresses above 4 GiB")
			}
		}
	}
	gc := 0
	previous := modifierPredicateObserve4E03D0
	modifierPredicateObserve4E03D0 = func(write, complete bool, base unsafe.Pointer, kind, index int, value modifierPredicateValue4E03D0) {
		if !write && !complete && kind == 0 {
			if base != unsafe.Pointer(effect) {
				t.Fatal("actual C callback must retain the original pinned effect")
			}
			runtime.GC()
			gc++
		}
	}
	defer func() { modifierPredicateObserve4E03D0 = previous; modifierPredicateTracing4E03D0(false) }()
	for _, row := range []struct {
		name            string
		strength        int32
		inversion, grip int32
	}{
		{"minimum", math.MinInt32, 0, 1},
		{"negative", -1, 0, 1},
		{"zero", 0, 0, 1},
		{"one", 1, 1, 0},
		{"maximum", math.MaxInt32, 1, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			effect.DefendCollide88.Val = row.strength
			args := [4]uintptr{uintptr(unsafe.Pointer(item)), uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(target)), uintptr(unsafe.Pointer(owner))}
			for which, expected := range []int32{row.inversion, row.grip} {
				function := modifierPredicatesCallback4E03D0(which)
				effect.DefendCollide88.Fnc = function
				scratch := [3]int32{0x13579bdf, 37, 0x2468ace0}
				modifierPredicateTracing4E03D0(true)
				result := modifierPredicateInvoke4E03D0(function, unsafe.Pointer(effect), args, &scratch[1])
				trace := modifierPredicateTraceRecords4E03D0()
				if result != expected || scratch != [3]int32{0x13579bdf, expected, 0x2468ace0} || len(trace) != 2 {
					t.Fatal("actual public C predicate return and guarded scratch", which, result, scratch, trace)
				}
				wantArgs := [6]uintptr{uintptr(unsafe.Pointer(effect)), args[0], args[1], args[2], args[3], uintptr(unsafe.Pointer(&scratch[1]))}
				if trace[0].Phase != 2 || trace[1].Phase != 3 || trace[0].Function != uintptr(function) || trace[1].Function != uintptr(function) || trace[0].Args != wantArgs || trace[1].Args != wantArgs || trace[0].Before != 37 || trace[1].After != expected || trace[1].Result != expected {
					t.Fatal("all six actual native arguments survive C-to-Go and forced GC", trace, wantArgs)
				}
				modifierPredicateTracing4E03D0(false)
			}
			effect.DefendCollide88.Fnc = modifierPredicatesCallback4E03D0(0)
			if got := inversionInventoryPublic4FA4F0(unsafe.Pointer(root), unsafe.Pointer(target)); got != row.inversion {
				t.Fatal("actual raw C checker on a pinned Go inventory/owner graph", got, row.inversion)
			}
			effect.DefendCollide88.Fnc = modifierPredicatesCallback4E03D0(1)
			if got := modifierPredicateNativeScanCall4E03D0(args[3], args[1], unsafe.Pointer(item), args[2]); got != row.inversion {
				t.Fatal("actual raw C Grip scan and native opaque contexts", got, row.inversion)
			}
			if root.InvFirstItem != item || item.InitData != unsafe.Pointer(attrs) || attrs.Modifiers[2] != effect || target.ObjOwner != owner || effect.DefendCollide88.Val != row.strength {
				t.Fatal("actual calls changed pinned inventory, owner or strength")
			}
		})
	}
	if gc != 20 {
		t.Fatal("all twenty actual predicate callbacks must survive in-call GC", gc)
	}
}
