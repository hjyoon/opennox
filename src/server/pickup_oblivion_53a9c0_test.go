package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

type pickupOblivionTestObject53A9C0 struct {
	name     string
	class    uint32
	subclass uint32
}

type pickupOblivionTestWorld53A9C0 struct {
	result int32
	state  int32
	trace  []string
}

func (w *pickupOblivionTestWorld53A9C0) event(format string, args ...any) {
	w.trace = append(w.trace, fmt.Sprintf(format, args...))
}

func (w *pickupOblivionTestWorld53A9C0) hooks() pickupOblivionHooks53A9C0[*pickupOblivionTestObject53A9C0] {
	return pickupOblivionHooks53A9C0[*pickupOblivionTestObject53A9C0]{
		weaponPickup: func(owner, item *pickupOblivionTestObject53A9C0, arg3, arg4 int32) int32 {
			w.event("weapon:%s:%s:%d:%d", owner.name, item.name, arg3, arg4)
			return w.result
		},
		loadOwnerClass: func(owner *pickupOblivionTestObject53A9C0) uint32 {
			w.event("class:%08x", owner.class)
			return owner.class
		},
		playerState: func(owner *pickupOblivionTestObject53A9C0) int32 {
			w.event("state:%s=%d", owner.name, w.state)
			return w.state
		},
		loadItemSubclass: func(item *pickupOblivionTestObject53A9C0) uint32 {
			w.event("subclass:%08x", item.subclass)
			return item.subclass
		},
		priorityMessage: func(owner *pickupOblivionTestObject53A9C0, message string, value uint8) {
			w.event("message:%s:%s:%d", owner.name, message, value)
		},
		audio: func(id uint32, owner *pickupOblivionTestObject53A9C0, kind int32, code uint32) {
			w.event("audio:%d:%s:%d:%d", id, owner.name, kind, code)
		},
		pauseFX: func(owner *pickupOblivionTestObject53A9C0, mode int32) {
			w.event("pause:%s:%d", owner.name, mode)
		},
		tryEquip: func(owner, item *pickupOblivionTestObject53A9C0) {
			w.event("equip:%s:%s", owner.name, item.name)
		},
	}
}

func TestPickupOblivion53A9C0RulesAndPriority(t *testing.T) {
	for index, rule := range pickupOblivionRules53A9C0 {
		t.Run(rule.message, func(t *testing.T) {
			owner := &pickupOblivionTestObject53A9C0{name: "owner", class: pickupOblivionPlayerClass53A9C0}
			item := &pickupOblivionTestObject53A9C0{name: "item", subclass: rule.mask}
			w := &pickupOblivionTestWorld53A9C0{result: 1}
			if got := pickupOblivion53A9C0(owner, item, math.MinInt32, math.MaxInt32, w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			want := []string{
				"weapon:owner:item:-2147483648:2147483647",
				"class:00000004",
				"state:owner=0",
			}
			for range index + 1 {
				want = append(want, fmt.Sprintf("subclass:%08x", rule.mask))
			}
			want = append(want,
				"message:owner:"+rule.message+":0",
				fmt.Sprintf("audio:%d:owner:0:0", rule.sound),
				"pause:owner:1",
				"equip:owner:item",
			)
			if !reflect.DeepEqual(w.trace, want) {
				t.Fatalf("trace =\n%v\nwant\n%v", w.trace, want)
			}
		})
	}

	t.Run("first matching bit wins", func(t *testing.T) {
		owner := &pickupOblivionTestObject53A9C0{name: "owner", class: pickupOblivionPlayerClass53A9C0}
		item := &pickupOblivionTestObject53A9C0{
			name: "item",
			subclass: pickupOblivionRules53A9C0[0].mask |
				pickupOblivionRules53A9C0[len(pickupOblivionRules53A9C0)-1].mask,
		}
		w := &pickupOblivionTestWorld53A9C0{result: 1}
		pickupOblivion53A9C0(owner, item, 0, 0, w.hooks())
		if got := w.trace[4]; got != "message:owner:weapon.c:PickupHalberdOblivion:0" {
			t.Fatalf("message = %q", got)
		}
	})
}

func TestPickupOblivion53A9C0ShortCircuits(t *testing.T) {
	owner := &pickupOblivionTestObject53A9C0{name: "owner", class: pickupOblivionPlayerClass53A9C0}
	item := &pickupOblivionTestObject53A9C0{name: "item", subclass: pickupOblivionRules53A9C0[0].mask}

	t.Run("weapon result", func(t *testing.T) {
		w := &pickupOblivionTestWorld53A9C0{result: math.MinInt32}
		if got := pickupOblivion53A9C0(owner, item, -17, -23, w.hooks()); got != math.MinInt32 {
			t.Fatalf("result = %d", got)
		}
		want := []string{"weapon:owner:item:-17:-23"}
		if !reflect.DeepEqual(w.trace, want) {
			t.Fatalf("trace = %v, want %v", w.trace, want)
		}
	})

	t.Run("non player", func(t *testing.T) {
		nonPlayer := &pickupOblivionTestObject53A9C0{name: "owner", class: uint32(object.ClassMonster)}
		w := &pickupOblivionTestWorld53A9C0{result: 1}
		if got := pickupOblivion53A9C0(nonPlayer, item, 0, 0, w.hooks()); got != 1 {
			t.Fatalf("result = %d", got)
		}
		want := []string{"weapon:owner:item:0:0", fmt.Sprintf("class:%08x", nonPlayer.class)}
		if !reflect.DeepEqual(w.trace, want) {
			t.Fatalf("trace = %v, want %v", w.trace, want)
		}
	})

	t.Run("inactive player state required", func(t *testing.T) {
		w := &pickupOblivionTestWorld53A9C0{result: 1, state: 7}
		if got := pickupOblivion53A9C0(owner, item, 0, 0, w.hooks()); got != 1 {
			t.Fatalf("result = %d", got)
		}
		want := []string{"weapon:owner:item:0:0", "class:00000004", "state:owner=7"}
		if !reflect.DeepEqual(w.trace, want) {
			t.Fatalf("trace = %v, want %v", w.trace, want)
		}
	})
}

func TestPickupOblivion53A9C0UnknownSubclassStillPausesAndEquips(t *testing.T) {
	owner := &pickupOblivionTestObject53A9C0{name: "owner", class: pickupOblivionPlayerClass53A9C0}
	item := &pickupOblivionTestObject53A9C0{name: "item"}
	w := &pickupOblivionTestWorld53A9C0{result: 1}
	if got := pickupOblivion53A9C0(owner, item, 0, 0, w.hooks()); got != 1 {
		t.Fatalf("result = %d", got)
	}
	wantTail := []string{"pause:owner:1", "equip:owner:item"}
	if got := w.trace[len(w.trace)-2:]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("trace tail = %v, want %v", got, wantTail)
	}
}

func TestPickupOblivionNative53A9C0BindsNativeFieldsAndPointers(t *testing.T) {
	owner := &Object{ObjClass: object.ClassPlayer | object.Class(0xa5000000)}
	item := &Object{ObjSubClass: object.SubClass(object.WeaponStaffOblivionHeart)}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(item)) <= math.MaxUint32) {
		t.Fatalf("native pointers = %p/%p, want above 32 bits", owner, item)
	}

	var trace []string
	deps := pickupOblivionNativeDeps53A9C0{
		weaponPickup: func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
			if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
				t.Fatalf("weapon args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
			}
			trace = append(trace, "weapon")
			return 1
		},
		playerState: func(gotOwner *Object) int32 {
			if gotOwner != owner {
				t.Fatalf("state owner = %p", gotOwner)
			}
			trace = append(trace, "state")
			return 0
		},
		priorityMessage: func(gotOwner *Object, message string, value uint8) {
			if gotOwner != owner || message != "weapon.c:PickupHeartOblivion" || value != 0 {
				t.Fatalf("message args = %p/%q/%d", gotOwner, message, value)
			}
			trace = append(trace, "message")
		},
		audio: func(id uint32, gotOwner *Object, kind int32, code uint32) {
			if id != 915 || gotOwner != owner || kind != 0 || code != 0 {
				t.Fatalf("audio args = %d/%p/%d/%d", id, gotOwner, kind, code)
			}
			trace = append(trace, "audio")
		},
		pauseFX: func(gotOwner *Object, mode int32) {
			if gotOwner != owner || mode != 1 {
				t.Fatalf("pause args = %p/%d", gotOwner, mode)
			}
			trace = append(trace, "pause")
		},
		tryEquip: func(gotOwner, gotItem *Object) {
			if gotOwner != owner || gotItem != item {
				t.Fatalf("equip args = %p/%p", gotOwner, gotItem)
			}
			trace = append(trace, "equip")
		},
	}
	if got := pickupOblivionNative53A9C0(owner, item, math.MinInt32, math.MaxInt32, deps); got != 1 {
		t.Fatalf("result = %d", got)
	}
	want := []string{"weapon", "state", "message", "audio", "pause", "equip"}
	if !reflect.DeepEqual(trace, want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}
