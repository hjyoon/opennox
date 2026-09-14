package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestTeleportMarkIndex530650(t *testing.T) {
	for _, first := range []spell.ID{
		spell.SPELL_TELEPORT_OTHER_TO_MARK_1,
		spell.SPELL_TELEPORT_TO_MARK_1,
	} {
		for index := 0; index < 4; index++ {
			if got, ok := teleportMarkIndex530650(first + spell.ID(index)); !ok || got != index {
				t.Fatalf("spell %d: index = %d, ok = %t", first+spell.ID(index), got, ok)
			}
		}
	}
	for _, id := range []spell.ID{spell.SPELL_TELEPORT_POP, spell.SPELL_TELEPORT_TO_TARGET} {
		if _, ok := teleportMarkIndex530650(id); ok {
			t.Fatalf("spell %d unexpectedly has a mark index", id)
		}
	}
}

func TestSpellTeleportToMarkCreate5305D0(t *testing.T) {
	data := &PlayerUpdateData{}
	data.Field29[1] = &Object{}
	source := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(data)}
	for _, tc := range []struct {
		name string
		coop bool
		want uint32
	}{
		{"normal", false, 101},
		{"coop", true, 102},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := &DurSpell{Spell: uint32(spell.SPELL_TELEPORT_TO_MARK_2), Level: 3, Obj12: source}
			rt := SpellTeleportToMarkRuntime5305D0{
				CoopMode: func() bool { return tc.coop },
				Frame:    func() uint32 { return 100 },
				TeleportDelay: func(level uint32) float32 {
					if level != 2 {
						t.Fatalf("delay level = %d", level)
					}
					return 2.5
				},
			}
			if got := SpellTeleportToMarkCreate5305D0(record, rt); got != 0 || record.Frame68 != tc.want {
				t.Fatalf("create/frame = %d/%d, want 0/%d", got, record.Frame68, tc.want)
			}
		})
	}
	if got := SpellTeleportToMarkCreate5305D0(&DurSpell{Spell: uint32(spell.SPELL_TELEPORT_TO_MARK_1), Obj12: source}, SpellTeleportToMarkRuntime5305D0{}); got != 1 {
		t.Fatalf("missing mark create = %d", got)
	}
	if got := SpellTeleportToMarkCreate5305D0(&DurSpell{}, SpellTeleportToMarkRuntime5305D0{}); got != 1 {
		t.Fatalf("nil source create = %d", got)
	}
}

func TestSpellTeleportToMarkUpdate530650Visible(t *testing.T) {
	marker := &Object{PosVec: types.Ptf(30, 40)}
	data := &PlayerUpdateData{}
	data.Field29[2] = marker
	data.Field39 = uint32(1) << 16
	player := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(data), PosVec: types.Ptf(10, 20)}
	record := &DurSpell{Spell: uint32(spell.SPELL_TELEPORT_TO_MARK_3), Obj12: player,
		Caster16: player, Target48: player, Frame68: 101,
		Pos: types.Ptf(math.Float32frombits(0x3fdccccc), 0)}
	wakeData := &TeleportWakeCollideData{}
	wake := &Object{CollideData: unsafe.Pointer(wakeData)}
	var events []string
	rt := SpellTeleportToMarkRuntime5305D0{
		Frame:    func() uint32 { return 100 },
		TickRate: func() uint32 { return 30 },
		NewObject: func(name string) *Object {
			events = append(events, "new wake")
			if name != "TeleportWake" {
				t.Fatalf("new object = %q", name)
			}
			return wake
		},
		CreateAt: func(obj, owner *Object, pos types.Pointf) {
			events = append(events, "create wake")
			if obj != wake || owner != player || pos != (types.Ptf(10, 20)) || wakeData.Destination != marker.PosVec {
				t.Fatalf("wake = %p/%p/%v, destination %v", obj, owner, pos, wakeData.Destination)
			}
		},
		OnSound: func(id spell.ID) sound.ID {
			events = append(events, "on sound")
			if id != spell.SPELL_TELEPORT_TO_MARK_3 {
				t.Fatalf("sound spell = %d", id)
			}
			return 97
		},
		Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			events = append(events, "audio")
			if id != 97 || obj != player || kind != 0 || code != 0 {
				t.Fatalf("audio = %d/%p/%d/%d", id, obj, kind, code)
			}
		},
		Teleport: func(obj *Object, pos types.Pointf) {
			events = append(events, "teleport")
			if obj != player || pos != marker.PosVec {
				t.Fatalf("teleport = %p/%v", obj, pos)
			}
			obj.PosVec = pos
		},
		SendPointFX: func(op netmsg.Op, pos types.Pointf) {
			switch op {
			case netmsg.MSG_FX_TELEPORT:
				events = append(events, "fx teleport")
			case netmsg.MSG_FX_BLUE_SPARKS:
				events = append(events, "fx sparks")
				if pos != marker.PosVec {
					t.Fatalf("sparks at %v", pos)
				}
			default:
				t.Fatalf("effect = %d", op)
			}
		},
		DelayedDelete: func(obj *Object) {
			events = append(events, "delete")
			if obj != marker {
				t.Fatal("deleted wrong marker")
			}
		},
		Attribution: func(from, to *Object) {
			events = append(events, "attribution")
			if from != player || to != player {
				t.Fatal("wrong attribution")
			}
		},
	}
	if got := SpellTeleportToMarkUpdate530650(record, rt); got != 1 || player.PosVec != marker.PosVec ||
		data.Field29[2] != nil || byte(data.Field39>>16) != 0 || marker.Field34 != 100 ||
		wakeData.Destination != marker.PosVec || wake.Field34 != 130 || math.Float32bits(record.Pos.X) != 0x3fdccccc {
		t.Fatalf("result = %d, pos = %v, charge = %#x, mark = %p", got, player.PosVec, data.Field39, data.Field29[2])
	}
	want := []string{"new wake", "create wake", "on sound", "audio", "teleport", "fx teleport", "fx teleport", "on sound", "audio", "fx sparks", "delete", "attribution"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellTeleportToMarkUpdate530650OtherInvisible(t *testing.T) {
	marker := &Object{PosVec: types.Ptf(30, 40)}
	data := &PlayerUpdateData{}
	data.Field29[1] = marker
	data.Field39 = 2 << 8
	source := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(data)}
	target := &Object{PosVec: types.Ptf(10, 20), Buffs: 1 << ENCHANT_INVISIBLE}
	record := &DurSpell{Spell: uint32(spell.SPELL_TELEPORT_OTHER_TO_MARK_2), Obj12: source,
		Caster16: source, Target48: target, Flag20: 1, Frame68: 101}
	var audioCount, attributionCount int
	rt := SpellTeleportToMarkRuntime5305D0{
		Frame:     func() uint32 { return 100 },
		TickRate:  func() uint32 { return 30 },
		NewObject: func(string) *Object { return nil },
		OnSound: func(id spell.ID) sound.ID {
			if id != spell.SPELL_TELEPORT_OTHER_TO_MARK_2 {
				t.Fatalf("sound spell = %d", id)
			}
			return 97
		},
		Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			audioCount++
			if id != 97 || obj != target || kind != 0 || code != 0 {
				t.Fatalf("audio = %d/%p/%d/%d", id, obj, kind, code)
			}
		},
		Teleport: func(obj *Object, pos types.Pointf) { obj.PosVec = pos },
		Attribution: func(from, to *Object) {
			attributionCount++
			if from != source || to != target {
				t.Fatal("wrong attribution")
			}
		},
	}
	if got := SpellTeleportToMarkUpdate530650(record, rt); got != 1 || target.PosVec != marker.PosVec ||
		byte(data.Field39>>8) != 1 || data.Field29[1] != marker || marker.Field34 != 100 ||
		audioCount != 1 || attributionCount != 1 {
		t.Fatalf("result = %d, pos = %v, charge = %#x, audio = %d, attribution = %d",
			got, target.PosVec, data.Field39, audioCount, attributionCount)
	}
}

func TestSpellTeleportToMarkUpdate530650Gates(t *testing.T) {
	source, target := &Object{}, &Object{}
	for _, tc := range []struct {
		name   string
		record DurSpell
		want   int32
	}{
		{"nil source", DurSpell{Target48: target}, 1},
		{"dead source", DurSpell{Obj12: &Object{ObjFlags: object.FlagDead}, Target48: target}, 1},
		{"nil target", DurSpell{Obj12: source}, 1},
		{"destroyed target", DurSpell{Obj12: source, Target48: &Object{ObjFlags: object.FlagDestroyed}}, 1},
		{"other without mode", DurSpell{Obj12: source, Target48: target}, 1},
		{"wrong frame", DurSpell{Obj12: source, Target48: target, Flag20: 1, Frame68: 103}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SpellTeleportToMarkUpdate530650(&tc.record,
				SpellTeleportToMarkRuntime5305D0{Frame: func() uint32 { return 100 }}); got != tc.want {
				t.Fatalf("update = %d, want %d", got, tc.want)
			}
		})
	}
	target.Buffs = 1 << ENCHANT_ANCHORED
	var heard bool
	if got := SpellTeleportToMarkUpdate530650(&DurSpell{Obj12: source, Target48: target, Flag20: 1, Frame68: 101},
		SpellTeleportToMarkRuntime5305D0{
			Frame: func() uint32 { return 100 },
			Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
				heard = id == 231 && obj == target && kind == 0 && code == 0
			},
		}); got != 1 || !heard {
		t.Fatalf("anchored update = %d, heard = %t", got, heard)
	}
}
