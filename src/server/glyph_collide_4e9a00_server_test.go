package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestGlyphCollide4E9A00NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantFlags = 20
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestGlyphCollideGateNative4E9A30PointersAndOrder(t *testing.T) {
	sourceParent := &Object{ObjClass: object.ClassPlayer}
	targetParent := &Object{ObjClass: object.ClassPlayer}
	source := &Object{}
	target := &Object{}
	parents := map[*Object]*Object{source: sourceParent, target: targetParent}
	var events []string
	got := glyphCollideGateNative4E9A30(source, target, glyphCollideGateNativeDeps4E9A30{
		gameFlag: func(flag uint32) int32 {
			events = append(events, "game")
			if flag == glyphCollideCoopTeamFlag4E9A30 {
				return 1
			}
			return 0
		},
		firstPlayerUnit: func() *Object {
			t.Fatal("non-Coop path read first player")
			return nil
		},
		unitsOnSameTeam: func(first, second *Object) int32 {
			events = append(events, "same")
			if first != source || second != target {
				t.Fatalf("same-team args = (%p, %p), want (%p, %p)", first, second, source, target)
			}
			return 0
		},
		findParent: func(obj *Object) *Object {
			events = append(events, "parent")
			return parents[obj]
		},
		abilityActive: func(obj *Object, ability int32) int32 {
			events = append(events, "ability")
			if obj != target || ability != glyphCollideTreadLightly4E9A30 {
				t.Fatalf("ability args = (%p, %d), want (%p, 4)", obj, ability, target)
			}
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"game", "same", "game", "parent", "parent", "ability"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestGlyphCollideGateNative4E9A30ReadsNativeFirstPlayerFlags(t *testing.T) {
	first := &Object{ObjFlags: object.Flags(0xa5000002)}
	var events []string
	got := glyphCollideGateNative4E9A30(nil, nil, glyphCollideGateNativeDeps4E9A30{
		gameFlag: func(flag uint32) int32 {
			events = append(events, "game")
			return 1
		},
		firstPlayerUnit: func() *Object {
			events = append(events, "first")
			return first
		},
	})
	if got != 0 || !reflect.DeepEqual(events, []string{"game", "first"}) {
		t.Fatalf("result/events = (%d, %#v), want (0, [game first])", got, events)
	}
}
