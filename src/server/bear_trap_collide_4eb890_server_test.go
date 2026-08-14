package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestBearTrapCollide4EB890NativeLayout(t *testing.T) {
	wantPos := uintptr(56)
	wantOwner := uintptr(508)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPos = 60
		wantOwner = 552
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.PosVec.Y", unsafe.Offsetof(Object{}.PosVec) + unsafe.Offsetof(types.Pointf{}.Y), wantPos + 4},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestBearTrapCollideNative4EB890UsesNativeFieldsAndEffects(t *testing.T) {
	owner := &Object{}
	source := &Object{PosVec: types.Pointf{X: 12.5, Y: -4.25}, ObjOwner: owner}
	target := &Object{}
	closed := &Object{}
	events := make([]string, 0, 8)

	bearTrapCollideNative4EB890(source, target, nil, bearTrapCollideNativeDeps4EB890{
		allowed: func(gotSource, gotTarget *Object) int32 {
			if gotSource != source || gotTarget != target {
				t.Fatalf("gate objects = (%p,%p), want (%p,%p)", gotSource, gotTarget, source, target)
			}
			events = append(events, "allowed")
			return 1
		},
		newObject: func(name string) *Object {
			if name != bearTrapClosedType4EB890 {
				t.Fatalf("object type = %q, want %q", name, bearTrapClosedType4EB890)
			}
			events = append(events, "new")
			return closed
		},
		createAt: func(gotClosed, gotOwner *Object, pos types.Pointf, reserved uint32) {
			if gotClosed != closed || gotOwner != owner || pos != source.PosVec || reserved != 0 {
				t.Fatalf("create = (%p,%p,%+v,%d), want (%p,%p,%+v,0)", gotClosed, gotOwner, pos, reserved, closed, owner, source.PosVec)
			}
			events = append(events, "create")
		},
		delayedDelete: func(got *Object) {
			if got != source {
				t.Fatalf("deleted object = %p, want %p", got, source)
			}
			events = append(events, "delete")
		},
		applyEnchant: func(got *Object, enchant EnchantID, duration, power uint32) {
			if got != target || duration != bearTrapEnchantDuration4EB890 || power != bearTrapEnchantPower4EB890 {
				t.Fatalf("enchant = (%p,%d,%d,%d)", got, enchant, duration, power)
			}
			events = append(events, "enchant:"+enchant.String())
		},
		audio: func(id uint32, got *Object, kind int32, code uint32) {
			if id != bearTrapTriggeredSound4EB890 || got != source || kind != 0 || code != 0 {
				t.Fatalf("audio = (%d,%p,%d,%d)", id, got, kind, code)
			}
			events = append(events, "audio")
		},
	})

	want := []string{
		"allowed", "new", "create", "delete",
		"enchant:ENCHANT_HELD", "enchant:ENCHANT_ANCHORED", "audio",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestBearTrapCollide4EB890RecoveredConstants(t *testing.T) {
	if EnchantID(bearTrapHeldEnchant4EB890) != ENCHANT_HELD {
		t.Fatalf("held enchant = %d, want %d", bearTrapHeldEnchant4EB890, ENCHANT_HELD)
	}
	if EnchantID(bearTrapAnchoredEnchant4EB890) != ENCHANT_ANCHORED {
		t.Fatalf("anchored enchant = %d, want %d", bearTrapAnchoredEnchant4EB890, ENCHANT_ANCHORED)
	}
	if sound.ID(bearTrapTriggeredSound4EB890) != sound.SoundBearTrapTriggered {
		t.Fatalf("sound = %d, want %d", bearTrapTriggeredSound4EB890, sound.SoundBearTrapTriggered)
	}
}
