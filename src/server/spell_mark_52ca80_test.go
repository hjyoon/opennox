package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type markCastTestWorld52CA80 struct {
	caster, aim, allocated, replacement *Object
	markers                             [4]*Object
	data, replacementData               *PlayerUpdateData
	cache, lookup, frame                uint32
	events                              []string
	name                                string
	moved, created, owner, audioObject  *Object
	movePosition                        *types.Pointf
	createPosition                      types.Pointf
	createdStamp                        uint32
	after                               func(string)
}

func newMarkCastTestWorld52CA80(t *testing.T) *markCastTestWorld52CA80 {
	t.Helper()
	newObject := func() *Object {
		obj, free := alloc.New(Object{})
		t.Cleanup(free)
		*obj = Object{}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
			t.Fatalf("native Mark object=%p, want actual C-owned allocation above 4 GiB", obj)
		}
		return obj
	}
	newData := func() *PlayerUpdateData {
		data, free := alloc.New(PlayerUpdateData{})
		t.Cleanup(free)
		*data = PlayerUpdateData{Field39: 0xa1b2c3d4}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(data)) <= math.MaxUint32 {
			t.Fatalf("native Mark player data=%p, want actual C-owned allocation above 4 GiB", data)
		}
		return data
	}
	w := &markCastTestWorld52CA80{
		caster: newObject(), aim: newObject(), allocated: newObject(), replacement: newObject(),
		data: newData(), replacementData: newData(), cache: 42, lookup: 42, frame: 100,
		createdStamp: 0x12345678,
	}
	*w.caster = Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(300, 500), UpdateData: unsafe.Pointer(w.data)}
	*w.aim = Object{TypeInd: 42, PosVec: types.Ptf(330, 520)}
	for i := range w.markers {
		w.markers[i] = newObject()
		w.markers[i].Field34 = []uint32{91, 81, 81, math.MaxUint32}[i]
		w.markers[i].PosVec = types.Ptf(float32(i+1), float32(i+11))
	}
	return w
}

func (w *markCastTestWorld52CA80) record(stage string) {
	w.events = append(w.events, stage)
	if w.after != nil {
		w.after(stage)
	}
}

func (w *markCastTestWorld52CA80) hooks(t *testing.T) markCastHooks52CA80 {
	t.Helper()
	return markCastHooks52CA80{
		loadCache:  func() uint32 { value := w.cache; w.record("cache"); return value },
		storeCache: func(value uint32) { w.cache = value; w.record("store") },
		lookupType: func(name string) uint32 {
			if name != "Glyph" {
				t.Fatalf("lookup=%q", name)
			}
			value := w.lookup
			w.record("lookup")
			return value
		},
		frame: func() uint32 { value := w.frame; w.record("frame"); return value },
		move: func(marker *Object, point *types.Pointf) {
			w.moved, w.movePosition = marker, point
			if marker != nil && marker.ObjClass&object.ClassImmobile == 0 {
				marker.PosVec = *point
			}
			w.record("move")
		},
		newObject: func(name string) *Object {
			w.name = name
			marker := w.allocated
			w.record("new")
			return marker
		},
		createAt: func(marker, owner *Object, point types.Pointf) {
			w.created, w.owner, w.createPosition = marker, owner, point
			marker.PosVec, marker.ObjOwner, marker.Field34 = point, owner, w.createdStamp
			w.record("create")
		},
		castSound: func(id int32) sound.ID {
			if id != -17 {
				t.Fatalf("sound spell=%d, want signed -17", id)
			}
			w.record("sound")
			return sound.ID(321)
		},
		audio: func(id sound.ID, target *Object, kind int, code uint32) {
			if id != sound.ID(321) || kind != 0 || code != 0 {
				t.Fatalf("audio id/kind/code=%d/%d/%d", id, kind, code)
			}
			w.audioObject = target
			w.record("audio")
		},
	}
}

func TestMarkCast52CA80FirstEmptyAndUnsignedOldestSlots(t *testing.T) {
	// Index 15 is the first of two equally old markers, not the final slot
	// whose timestamp is 0xffffffff (unsigned, not negative).
	wantIndex := [16]int{0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 1}
	for mask := 0; mask < 16; mask++ {
		for _, position := range []string{"caster-nil-aim", "caster-other-type", "aim-glyph", "caster-wide-cache"} {
			t.Run(fmt.Sprintf("mask-%02x/%s", mask, position), func(t *testing.T) {
				w := newMarkCastTestWorld52CA80(t)
				for i := range w.data.Field29 {
					if mask&(1<<i) != 0 {
						w.data.Field29[i] = w.markers[i]
					}
				}
				aim := w.aim
				point := &w.caster.PosVec
				switch position {
				case "caster-nil-aim":
					aim = nil
				case "caster-other-type":
					aim.TypeInd = 41
				case "aim-glyph":
					point = &w.aim.PosVec
				case "caster-wide-cache":
					w.cache = 0x1002a // compare zero-extended WORD to the full DWORD
				}
				beforeSlots, beforeCharges := w.data.Field29, w.data.Field39
				if got := markCast52CA80(-17, w.caster, aim, w.hooks(t)); got != 1 {
					t.Fatalf("result=%d", got)
				}
				index := wantIndex[mask]
				wantEvents := []string{"cache", "new", "create", "sound", "audio"}
				if mask == 15 {
					wantEvents = []string{"cache", "frame", "move", "frame", "sound", "audio"}
					if w.moved != w.markers[index] || w.movePosition != point || w.moved.PosVec != *point || w.moved.Field34 != 100 || w.created != nil {
						t.Fatal("oldest marker relocation/frame/position changed")
					}
				} else {
					beforeSlots[index] = w.allocated
					if w.name != fmt.Sprintf("TeleportGlyph%d", index+1) || w.created != w.allocated || w.owner != w.caster ||
						w.createPosition != *point || w.created.Field34 != w.createdStamp || w.moved != nil {
						t.Fatal("first-empty marker allocation/placement or creation-owned timestamp changed")
					}
				}
				shift := uint(index * 8)
				wantCharges := beforeCharges&^(uint32(0xff)<<shift) | uint32(3)<<shift
				if w.data.Field29 != beforeSlots || w.data.Field39 != wantCharges || w.audioObject != w.caster || !reflect.DeepEqual(w.events, wantEvents) {
					t.Fatalf("slots/charges/audio/events=%v/%08x/%p/%v, want %v/%08x/%p/%v", w.data.Field29, w.data.Field39, w.audioObject, w.events, beforeSlots, wantCharges, w.caster, wantEvents)
				}
			})
		}
	}
}

func TestMarkCast52CA80CacheAndLowBytePlayerGate(t *testing.T) {
	for _, class := range []uint32{0, 2, 4, 6, 0x400, 0x404, 0xffffffff} {
		for _, cache := range []uint32{0, 42, 0x1002a} {
			t.Run(fmt.Sprintf("class-%08x/cache-%08x", class, cache), func(t *testing.T) {
				w := newMarkCastTestWorld52CA80(t)
				w.cache, w.caster.ObjClass = cache, object.Class(class)
				w.caster.ObjFlags = object.FlagDead | object.FlagDestroyed
				if class&4 == 0 {
					w.caster.UpdateData = nil // non-player gate must precede the data load
				}
				beforeCaster, beforeData := *w.caster, *w.data
				got := markCast52CA80(-17, w.caster, nil, w.hooks(t))
				want := []string{"cache"}
				if cache == 0 {
					want = append(want, "lookup", "store")
					cache = 42
				}
				if class&4 != 0 {
					want = append(want, "new", "create", "sound", "audio")
				} else if *w.caster != beforeCaster || *w.data != beforeData || w.audioObject != nil {
					t.Fatal("non-player inspected update data or changed records")
				}
				if got != 1 || w.cache != cache || !reflect.DeepEqual(w.events, want) {
					t.Fatalf("result/cache/events=%d/%08x/%v, want 1/%08x/%v", got, w.cache, w.events, cache, want)
				}
			})
		}
	}
}

func TestMarkCast52CA80AllocationFailureLeavesChargesAndSounds(t *testing.T) {
	for index := 0; index < 4; index++ {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			w := newMarkCastTestWorld52CA80(t)
			for i := 0; i < index; i++ {
				w.data.Field29[i] = w.markers[i]
			}
			before := *w.data
			w.allocated = nil
			w.after = func(stage string) {
				if stage == "new" {
					w.data.Field29[index] = w.replacement // overwritten by original return value
				}
			}
			if got := markCast52CA80(-17, w.caster, w.aim, w.hooks(t)); got != 1 || *w.data != before ||
				w.name != fmt.Sprintf("TeleportGlyph%d", index+1) || w.created != nil || w.audioObject != w.caster ||
				!reflect.DeepEqual(w.events, []string{"cache", "new", "sound", "audio"}) {
				t.Fatalf("failed allocation result/data/events=%d/%+v/%v", got, *w.data, w.events)
			}
		})
	}
}

func TestMarkCast52CA80NewMarkerKeepsCachedDataAndLivePoint(t *testing.T) {
	for _, position := range []string{"caster", "glyph"} {
		for index := 0; index < 4; index++ {
			t.Run(fmt.Sprintf("%s/slot-%d", position, index), func(t *testing.T) {
				w := newMarkCastTestWorld52CA80(t)
				for i := 0; i < index; i++ {
					w.data.Field29[i] = w.markers[i]
				}
				point := w.caster
				if position == "glyph" {
					point = w.aim
				} else {
					w.aim.TypeInd = 41
				}
				w.after = func(stage string) {
					switch stage {
					case "new":
						w.caster.UpdateData = unsafe.Pointer(w.replacementData)
						w.data.Field29[index] = w.replacement
						point.PosVec = types.Ptf(901, 902)
						w.aim.TypeInd = 41 // position selection must not be recomputed
					case "create":
						if w.data.Field29[index] != w.allocated {
							t.Fatal("allocation result was not stored before creation")
						}
						w.data.Field29[index] = w.replacement
						w.data.Field39 = 0x10203040
					}
				}
				if got := markCast52CA80(-17, w.caster, w.aim, w.hooks(t)); got != 1 || w.created != w.allocated ||
					w.createPosition != types.Ptf(901, 902) || w.data.Field29[index] != w.replacement ||
					w.replacementData.Field39 != 0xa1b2c3d4 || w.replacementData.Field29 != [4]*Object{} ||
					w.data.Field39 != (uint32(0x10203040)&^(uint32(0xff)<<uint(index*8))|uint32(3)<<uint(index*8)) {
					t.Fatalf("cached data/live point result=%d, position=%v, data=%+v", got, w.createPosition, *w.data)
				}
			})
		}
	}
}

func TestMarkCast52CA80MoveReloadsSlotBeforeFrame(t *testing.T) {
	for _, immobile := range []bool{false, true} {
		t.Run(fmt.Sprintf("immobile-%t", immobile), func(t *testing.T) {
			w := newMarkCastTestWorld52CA80(t)
			w.data.Field29 = w.markers
			if immobile {
				w.markers[1].ObjClass = object.ClassImmobile
			}
			beforePosition := w.markers[1].PosVec
			frames := 0
			w.after = func(stage string) {
				switch stage {
				case "frame":
					frames++
					if frames == 2 {
						w.data.Field29[1] = w.allocated // later than selected pointer reload
					}
				case "move":
					w.caster.UpdateData = unsafe.Pointer(w.replacementData)
					w.data.Field29[1] = w.replacement
					w.data.Field39, w.frame = 0x10203040, 123
				}
			}
			if got := markCast52CA80(-17, w.caster, w.aim, w.hooks(t)); got != 1 || w.moved != w.markers[1] ||
				w.replacement.Field34 != 123 || w.allocated.Field34 != 0 || w.markers[1].Field34 != 81 ||
				w.data.Field29[1] != w.allocated || w.data.Field39 != 0x10200340 || w.replacementData.Field39 != 0xa1b2c3d4 || frames != 2 {
				t.Fatalf("move reload/frame/charge result=%d frames=%d data=%+v", got, frames, *w.data)
			}
			wantPosition := w.aim.PosVec
			if immobile {
				wantPosition = beforePosition
			}
			if w.markers[1].PosVec != wantPosition || !reflect.DeepEqual(w.events, []string{"cache", "frame", "move", "frame", "sound", "audio"}) {
				t.Fatal("move early return or service order changed")
			}
		})
	}
}

func TestMarkCast52CA80OriginalRequiredFaultPrefixes(t *testing.T) {
	for _, fault := range []string{"nil-caster", "nil-player-data", "all-marks-not-older", "slot-cleared-by-move"} {
		t.Run(fault, func(t *testing.T) {
			w := newMarkCastTestWorld52CA80(t)
			caster := w.caster
			want := []string{"cache"}
			switch fault {
			case "nil-caster":
				caster = nil
			case "nil-player-data":
				caster.UpdateData = nil
			case "all-marks-not-older":
				w.data.Field29 = w.markers
				for _, marker := range w.markers {
					marker.Field34 = 100
				}
				want = append(want, "frame")
			case "slot-cleared-by-move":
				w.data.Field29 = w.markers
				w.after = func(stage string) {
					if stage == "move" {
						w.data.Field29[1] = nil
					}
				}
				want = append(want, "frame", "move", "frame")
			}
			defer func() {
				if recover() == nil || !reflect.DeepEqual(w.events, want) || w.data.Field39 != 0xa1b2c3d4 {
					t.Fatalf("required fault/prefix/charges=%v/%08x, want %v/a1b2c3d4", w.events, w.data.Field39, want)
				}
			}()
			markCast52CA80(-17, caster, w.aim, w.hooks(t))
		})
	}
}

func TestMarkCast52CA80LookupReturnSnapshotAndLiveClass(t *testing.T) {
	for _, lookup := range []uint32{0, 42, 0x1002a, math.MaxUint32} {
		for _, turnPlayer := range []bool{false, true} {
			t.Run(fmt.Sprintf("lookup-%08x/player-%t", lookup, turnPlayer), func(t *testing.T) {
				w := newMarkCastTestWorld52CA80(t)
				w.cache, w.lookup, w.caster.ObjClass = 0, lookup, object.ClassMonster
				if lookup == 0 {
					w.aim.TypeInd = 0 // the original does not reject a zero lookup result
				}
				w.after = func(stage string) {
					if stage == "store" {
						w.cache = 99 // comparison uses lookup's saved return, not a reload
						if turnPlayer {
							w.caster.ObjClass = object.ClassPlayer
						}
					}
				}
				if got := markCast52CA80(-17, w.caster, w.aim, w.hooks(t)); got != 1 || w.cache != 99 {
					t.Fatalf("lookup snapshot result/cache=%d/%d", got, w.cache)
				}
				want := []string{"cache", "lookup", "store"}
				if turnPlayer {
					want = append(want, "new", "create", "sound", "audio")
					point := w.caster.PosVec
					if lookup == 0 || lookup == 42 {
						point = w.aim.PosVec
					}
					if w.createPosition != point {
						t.Fatal("local Glyph cache comparison or zero-result behavior changed")
					}
				}
				if !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events=%v, want %v", w.events, want)
				}
			})
		}
	}
}

func TestMarkCast52CA80UnsignedTimestampBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		frame  uint32
		stamps [4]uint32
		index  int
	}{
		{"first-tie", 100, [4]uint32{10, 10, 20, 30}, 0},
		{"last-oldest", 100, [4]uint32{40, 30, 20, 10}, 3},
		{"equality-not-older", 100, [4]uint32{100, 99, 100, 101}, 1},
		{"zero-oldest", 1, [4]uint32{math.MaxUint32, 0x80000000, 1, 0}, 3},
		{"unsigned-high-half", math.MaxUint32, [4]uint32{math.MaxUint32, 0x80000001, 0x80000000, 0x80000002}, 2},
		{"unsigned-low-half", 0x80000001, [4]uint32{0x80000000, 0x7fffffff, 0x80000001, math.MaxUint32}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newMarkCastTestWorld52CA80(t)
			w.frame, w.data.Field29 = tc.frame, w.markers
			for i, stamp := range tc.stamps {
				w.markers[i].Field34 = stamp
			}
			if got := markCast52CA80(-17, w.caster, nil, w.hooks(t)); got != 1 || w.moved != w.markers[tc.index] || w.moved.Field34 != tc.frame ||
				!reflect.DeepEqual(w.events, []string{"cache", "frame", "move", "frame", "sound", "audio"}) {
				t.Fatalf("unsigned oldest index/result/events=%p/%d/%v, want marker %d", w.moved, got, w.events, tc.index)
			}
		})
	}
}
