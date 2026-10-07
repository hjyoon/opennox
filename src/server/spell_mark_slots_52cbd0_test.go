package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func markSlotTestHooks52CBD0(t *testing.T, w *markCastTestWorld52CA80, id int32) markCastHooks52CA80 {
	t.Helper()
	h := w.hooks(t)
	h.castSound = func(actual int32) sound.ID {
		if actual != id {
			t.Fatalf("slot cast sound id=%d, want=%d", actual, id)
		}
		w.record("sound")
		return 321
	}
	return h
}

func TestMarkSlotCast52CBD0SelectsNamedSlotNotEmptyOrOldest(t *testing.T) {
	if spell.SPELL_MARK_1 != 46 || spell.SPELL_MARK_4 != 49 {
		t.Fatal("stock Mark slot IDs changed")
	}
	for index := 0; index < 4; index++ {
		for mask := 0; mask < 16; mask++ {
			for _, location := range []string{"nil-aim", "other-type", "glyph", "wide-cache"} {
				t.Run(fmt.Sprintf("slot-%d/mask-%02x/%s", index, mask, location), func(t *testing.T) {
					w := newMarkCastTestWorld52CA80(t)
					id := int32(spell.SPELL_MARK_1) + int32(index)
					for i := range w.markers {
						if mask&(1<<i) != 0 {
							w.data.Field29[i] = w.markers[i]
						}
						// Deliberately future/equal stamps: named slots never
						// scan oldest timestamps, even with all slots occupied.
						w.markers[i].Field34 = math.MaxUint32
					}
					point, aim := &w.caster.PosVec, w.aim
					switch location {
					case "nil-aim":
						aim = nil
					case "other-type":
						aim.TypeInd = 41
					case "glyph":
						point = &aim.PosVec
					case "wide-cache":
						w.cache = 0x1002a
					}
					beforeSlots, beforeCharges := w.data.Field29, w.data.Field39
					beforeObjects := [4]Object{*w.markers[0], *w.markers[1], *w.markers[2], *w.markers[3]}
					if got := markSlotCast52CBD0(id, w.caster, aim, markSlotTestHooks52CBD0(t, w, id)); got != 1 {
						t.Fatalf("slot result=%d", got)
					}
					wantEvents := []string{"cache", "new", "create", "sound", "audio"}
					if beforeSlots[index] == nil {
						beforeSlots[index] = w.allocated
						if w.name != fmt.Sprintf("TeleportGlyph%d", index+1) || w.created != w.allocated || w.owner != w.caster ||
							w.createPosition != *point || w.created.Field34 != w.createdStamp || w.moved != nil {
							t.Fatal("named-slot allocation/creation position or timestamp changed")
						}
					} else {
						wantEvents = []string{"cache", "move", "frame", "sound", "audio"}
						if w.moved != w.markers[index] || w.movePosition != point || w.moved.PosVec != *point || w.moved.Field34 != 100 || w.created != nil {
							t.Fatal("named-slot movement changed target or loaded an oldest-selection frame")
						}
					}
					for i := range w.markers {
						if i != index && *w.markers[i] != beforeObjects[i] {
							t.Fatalf("unrelated native marker %d changed", i)
						}
					}
					shift := uint(index * 8)
					wantCharges := beforeCharges&^(uint32(0xff)<<shift) | uint32(3)<<shift
					if w.data.Field29 != beforeSlots || w.data.Field39 != wantCharges || w.audioObject != w.caster || !reflect.DeepEqual(w.events, wantEvents) {
						t.Fatalf("slots/charges/audio/events=%v/%08x/%p/%v", w.data.Field29, w.data.Field39, w.audioObject, w.events)
					}
				})
			}
		}
	}
}

func TestMarkSlotCast52CBD0CacheAndLowByteGate(t *testing.T) {
	for _, class := range []uint32{0, 2, 4, 6, 0x400, 0x404, 0xffffffff} {
		for _, lookup := range []uint32{0, 42, 0x1002a, math.MaxUint32} {
			t.Run(fmt.Sprintf("class-%08x/lookup-%08x", class, lookup), func(t *testing.T) {
				w := newMarkCastTestWorld52CA80(t)
				w.cache, w.lookup, w.caster.ObjClass = 0, lookup, object.Class(class)
				w.caster.ObjFlags = object.FlagDead | object.FlagDestroyed
				if class&4 == 0 {
					w.caster.UpdateData = nil
				}
				if lookup == 0 {
					w.aim.TypeInd = 0
				}
				before := *w.data
				w.after = func(stage string) {
					if stage == "store" {
						w.cache = 99 // Type comparison retains the lookup return.
					}
				}
				const id = int32(47)
				if got := markSlotCast52CBD0(id, w.caster, w.aim, markSlotTestHooks52CBD0(t, w, id)); got != 1 || w.cache != 99 {
					t.Fatalf("result/cache=%d/%d", got, w.cache)
				}
				wantEvents := []string{"cache", "lookup", "store"}
				if class&4 != 0 {
					wantEvents = append(wantEvents, "new", "create", "sound", "audio")
					point := w.caster.PosVec
					if lookup == 0 || lookup == 42 {
						point = w.aim.PosVec
					}
					if w.createPosition != point {
						t.Fatal("WORD/full-DWORD comparison or saved lookup result changed")
					}
				} else if *w.data != before || w.audioObject != nil {
					t.Fatal("non-player gate touched data or audio")
				}
				if !reflect.DeepEqual(w.events, wantEvents) {
					t.Fatalf("events=%v want=%v", w.events, wantEvents)
				}
			})
		}
	}
}

func TestMarkSlotCast52CBD0AllocationFailureAndCallbackReloads(t *testing.T) {
	for index := 0; index < 4; index++ {
		for _, mode := range []string{"allocation-fails", "new-caster", "new-glyph", "move", "immobile"} {
			t.Run(fmt.Sprintf("slot-%d/%s", index, mode), func(t *testing.T) {
				w := newMarkCastTestWorld52CA80(t)
				id := int32(46 + index)
				point := w.caster
				if mode == "new-glyph" {
					point = w.aim
				} else {
					w.aim.TypeInd = 41
				}
				if mode == "allocation-fails" {
					w.allocated = nil
				}
				if mode == "move" || mode == "immobile" {
					w.data.Field29 = w.markers
					if mode == "immobile" {
						w.markers[index].ObjClass = object.ClassImmobile
					}
				}
				beforePosition, beforeStamp := w.markers[index].PosVec, w.markers[index].Field34
				w.after = func(stage string) {
					switch stage {
					case "new":
						w.caster.UpdateData = unsafe.Pointer(w.replacementData)
						w.data.Field29[index] = w.replacement
						point.PosVec = types.Ptf(901, 902)
						w.aim.TypeInd = 41
					case "create":
						if w.data.Field29[index] != w.allocated {
							t.Fatal("allocation pointer must be stored before creation")
						}
						w.data.Field29[index] = w.replacement
						w.data.Field39 = 0x10203040
					case "move":
						w.caster.UpdateData = unsafe.Pointer(w.replacementData)
						w.data.Field29[index] = w.replacement
						w.data.Field39, w.frame = 0x10203040, 123
					case "frame":
						w.data.Field29[index] = w.allocated // after pointer reload
					}
				}
				got := markSlotCast52CBD0(id, w.caster, w.aim, markSlotTestHooks52CBD0(t, w, id))
				if got != 1 || w.audioObject != w.caster || w.replacementData.Field29 != [4]*Object{} || w.replacementData.Field39 != 0xa1b2c3d4 {
					t.Fatal("cached record or caster sound changed")
				}
				if mode == "allocation-fails" {
					if w.data.Field29[index] != nil || w.data.Field39 != 0xa1b2c3d4 || w.created != nil ||
						!reflect.DeepEqual(w.events, []string{"cache", "new", "sound", "audio"}) {
						t.Fatal("failed allocation recharged or failed to overwrite the callback slot")
					}
					return
				}
				shift := uint(index * 8)
				wantCharges := uint32(0x10203040)&^(uint32(0xff)<<shift) | uint32(3)<<shift
				if w.data.Field39 != wantCharges {
					t.Fatalf("packed charge=%08x want=%08x", w.data.Field39, wantCharges)
				}
				if mode == "move" || mode == "immobile" {
					wantPosition := point.PosVec
					if mode == "immobile" {
						wantPosition = beforePosition
					}
					if w.moved != w.markers[index] || w.moved.PosVec != wantPosition || w.moved.Field34 != beforeStamp ||
						w.replacement.Field34 != 123 || w.allocated.Field34 != 0 || w.data.Field29[index] != w.allocated ||
						!reflect.DeepEqual(w.events, []string{"cache", "move", "frame", "sound", "audio"}) {
						t.Fatal("move did not reload the slot before frame or preserve the immobile gate")
					}
				} else if w.createPosition != types.Ptf(901, 902) || w.created != w.allocated || w.data.Field29[index] != w.replacement ||
					w.created.Field34 != w.createdStamp || !reflect.DeepEqual(w.events, []string{"cache", "new", "create", "sound", "audio"}) {
					t.Fatal("new marker lost entry-selected live point, cached data, or creation-owned timestamp")
				}
			})
		}
	}
}

func TestMarkSlotCast52CBD0RequiredFaultPrefixes(t *testing.T) {
	for _, fault := range []string{"nil-caster", "nil-player-data", "slot-cleared-by-move"} {
		t.Run(fault, func(t *testing.T) {
			w := newMarkCastTestWorld52CA80(t)
			caster := w.caster
			want := []string{"cache"}
			switch fault {
			case "nil-caster":
				caster = nil
			case "nil-player-data":
				caster.UpdateData = nil
			case "slot-cleared-by-move":
				w.data.Field29[2] = w.markers[2]
				w.after = func(stage string) {
					if stage == "move" {
						w.data.Field29[2] = nil
					}
				}
				want = append(want, "move", "frame")
			}
			defer func() {
				if recover() == nil || w.data.Field39 != 0xa1b2c3d4 || !reflect.DeepEqual(w.events, want) {
					t.Fatalf("required fault prefix=%v want=%v", w.events, want)
				}
			}()
			markSlotCast52CBD0(48, caster, w.aim, markSlotTestHooks52CBD0(t, w, 48))
		})
	}
}
