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

func TestSpellTeleportPopCreate530820(t *testing.T) {
	for _, tc := range []struct {
		coop bool
		want uint32
	}{
		{false, 101}, {true, 102},
	} {
		record := &DurSpell{Level: 3}
		rt := SpellTeleportPopRuntime530820{
			CoopMode: func() bool { return tc.coop },
			Frame:    func() uint32 { return 100 },
			TeleportDelay: func(level uint32) float32 {
				if level != 2 {
					t.Fatalf("delay level = %d", level)
				}
				return 2.5
			},
		}
		if got := SpellTeleportPopCreate530820(record, rt); got != 0 || record.Frame68 != tc.want {
			t.Fatalf("create/frame = %d/%d, want 0/%d", got, record.Frame68, tc.want)
		}
	}
}

func TestSpellTeleportPopUpdate530880Gates(t *testing.T) {
	caster, target := &Object{}, &Object{}
	for _, tc := range []struct {
		name   string
		record DurSpell
		want   int32
	}{
		{"nil caster", DurSpell{Target48: target}, 1},
		{"nil target", DurSpell{Caster16: caster}, 1},
		{"dead caster", DurSpell{Caster16: &Object{ObjFlags: object.FlagDead}, Target48: target}, 1},
		{"destroyed target", DurSpell{Caster16: caster, Target48: &Object{ObjFlags: object.FlagDestroyed}}, 1},
		{"wrong frame", DurSpell{Caster16: caster, Target48: target, Frame68: 103}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SpellTeleportPopUpdate530880(&tc.record,
				SpellTeleportPopRuntime530820{Frame: func() uint32 { return 100 }}); got != tc.want {
				t.Fatalf("update = %d, want %d", got, tc.want)
			}
		})
	}
	target.Buffs = 1 << ENCHANT_ANCHORED
	var heard bool
	if got := SpellTeleportPopUpdate530880(&DurSpell{Caster16: caster, Target48: target, Frame68: 101},
		SpellTeleportPopRuntime530820{
			Frame: func() uint32 { return 100 },
			Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
				heard = id == 231 && obj == target && kind == 0 && code == 0
			},
		}); got != 1 || !heard {
		t.Fatalf("anchored update = %d, heard = %t", got, heard)
	}
}

func TestSpellTeleportPopUpdate530880MarkerSelectionAndCharge(t *testing.T) {
	for _, tc := range []struct {
		name       string
		invisible  bool
		charges    byte
		wantEvents []string
	}{
		{"visible last charge", false, 1, []string{"random", "random", "new wake", "create wake", "teleport", "fx teleport", "fx teleport", "cast sound", "audio", "fx sparks", "delete", "attribution"}},
		{"invisible remaining charge", true, 2, []string{"random", "random", "new wake", "create wake", "teleport", "cast sound", "audio", "attribution"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marker := &Object{PosVec: types.Ptf(30, 40)}
			data := &PlayerUpdateData{}
			data.Field29[2] = marker
			data.Field39 = uint32(tc.charges) << 16
			caster := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(data)}
			target := &Object{PosVec: types.Ptf(10, 20)}
			if tc.invisible {
				target.Buffs = 1 << ENCHANT_INVISIBLE
			}
			record := &DurSpell{Spell: uint32(spell.SPELL_TELEPORT_POP), Caster16: caster,
				Target48: target, Frame68: 101, Pos: types.Ptf(math.Float32frombits(0x3fdccccc), 0)}
			wakeData := &TeleportWakeCollideData{}
			wake := &Object{CollideData: unsafe.Pointer(wakeData)}
			var events []string
			randomCalls := 0
			rt := SpellTeleportPopRuntime530820{
				Frame:    func() uint32 { return 100 },
				TickRate: func() uint32 { return 30 },
				RandomIndex: func() int {
					events = append(events, "random")
					randomCalls++
					if randomCalls == 1 {
						return 0
					}
					return 2
				},
				NewObject: func(name string) *Object {
					events = append(events, "new wake")
					if name != "TeleportWake" {
						t.Fatalf("new object = %q", name)
					}
					return wake
				},
				CreateAt: func(obj, owner *Object, pos types.Pointf) {
					events = append(events, "create wake")
					if obj != wake || owner != target || pos != (types.Ptf(10, 20)) || wakeData.Destination != marker.PosVec {
						t.Fatalf("wake = %p %p %v, destination %v", obj, owner, pos, wakeData.Destination)
					}
				},
				Teleport: func(obj *Object, pos types.Pointf) {
					events = append(events, "teleport")
					if obj != target || pos != marker.PosVec {
						t.Fatalf("teleport = %p %v", obj, pos)
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
							t.Fatalf("spark point = %v", pos)
						}
					default:
						t.Fatalf("effect = %d", op)
					}
				},
				CastSound: func(id spell.ID) sound.ID {
					events = append(events, "cast sound")
					if id != spell.SPELL_TELEPORT_POP {
						t.Fatalf("cast spell = %d", id)
					}
					return 97
				},
				Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
					events = append(events, "audio")
					if id != 97 || obj != target || kind != 0 || code != 0 {
						t.Fatalf("audio = %d %p %d %d", id, obj, kind, code)
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
					if from != caster || to != target {
						t.Fatal("wrong attribution")
					}
				},
			}
			if got := SpellTeleportPopUpdate530880(record, rt); got != 1 || target.PosVec != marker.PosVec ||
				wakeData.Destination != marker.PosVec || wake.Field34 != 130 ||
				byte(data.Field39>>16) != tc.charges-1 || math.Float32bits(record.Pos.X) != 0x3fdccccc {
				t.Fatalf("result = %d, target = %v, charge = %#x", got, target.PosVec, data.Field39)
			}
			if tc.charges == 1 && data.Field29[2] != nil || tc.charges > 1 && data.Field29[2] != marker {
				t.Fatalf("marker after use = %p", data.Field29[2])
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}
