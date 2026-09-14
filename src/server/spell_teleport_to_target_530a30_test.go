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

func TestSpellTeleportToTargetCreate530A30(t *testing.T) {
	caster := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(10, 20)}
	target := &Object{PosVec: types.Ptf(30, 40)}
	for _, tc := range []struct {
		name      string
		blocked   bool
		clear     bool
		coop      bool
		want      int32
		wantFrame uint32
		wantCalls []string
	}{
		{"blocked tile", true, true, false, 1, 0, []string{"tile", "unseen", "audio"}},
		{"blocked ray", false, false, false, 1, 0, []string{"tile", "ray", "inform"}},
		{"arena", false, true, false, 0, 101, []string{"tile", "ray", "coop", "frame"}},
		{"coop round to even", false, true, true, 0, 102, []string{"tile", "ray", "coop", "delay", "frame"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			record := &DurSpell{Caster16: caster, Target48: target, Pos2: types.Ptf(50, 60), Level: 2}
			rt := SpellTeleportToTargetRuntime530A30{
				TileBlocksTeleport: func(pos *types.Pointf) bool {
					calls = append(calls, "tile")
					if *pos != record.Pos2 {
						t.Fatalf("tile point = %v", *pos)
					}
					return tc.blocked
				},
				TraceRay9: func(from, to types.Pointf) bool {
					calls = append(calls, "ray")
					if from != target.PosVec || to != record.Pos2 {
						t.Fatalf("ray = %v to %v", from, to)
					}
					return tc.clear
				},
				SendUnseenTarget: func(obj *Object) {
					calls = append(calls, "unseen")
					if obj != target {
						t.Fatal("unseen recipient is not target")
					}
				},
				Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
					calls = append(calls, "audio")
					if id != 231 || obj != target || kind != 0 || code != 0 {
						t.Fatalf("audio = %d %p %d %d", id, obj, kind, code)
					}
				},
				InformNoLOS: func(obj *Object) {
					calls = append(calls, "inform")
					if obj != caster {
						t.Fatal("LOS recipient is not caster")
					}
				},
				CoopMode: func() bool { calls = append(calls, "coop"); return tc.coop },
				TeleportDelay: func(level uint32) float32 {
					calls = append(calls, "delay")
					if level != 1 {
						t.Fatalf("delay level = %d", level)
					}
					return 2.5
				},
				Frame: func() uint32 { calls = append(calls, "frame"); return 100 },
			}
			if got := SpellTeleportToTargetCreate530A30(record, rt); got != tc.want || record.Frame68 != tc.wantFrame {
				t.Fatalf("create/frame = %d/%d, want %d/%d", got, record.Frame68, tc.want, tc.wantFrame)
			}
			if !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, tc.wantCalls)
			}
		})
	}
}

func TestSpellTeleportToTargetCreateDefaultsTargetAndKeepsNativePointer(t *testing.T) {
	caster := &Object{PosVec: types.Ptf(10, 20)}
	record := &DurSpell{Caster16: caster, Pos2: types.Ptf(11, 21), Pos: types.Ptf(math.Float32frombits(0x3fdccccc), 0)}
	rt := SpellTeleportToTargetRuntime530A30{
		TileBlocksTeleport: func(*types.Pointf) bool { return false },
		TraceRay9:          func(from, to types.Pointf) bool { return from == caster.PosVec && to == record.Pos2 },
		CoopMode:           func() bool { return false },
		Frame:              func() uint32 { return 10 },
	}
	if got := SpellTeleportToTargetCreate530A30(record, rt); got != 0 || record.Target48 != caster ||
		record.Frame68 != 11 || math.Float32bits(record.Pos.X) != 0x3fdccccc {
		t.Fatalf("create = %d, target = %p, frame = %d, sentinel = %#x", got, record.Target48, record.Frame68, math.Float32bits(record.Pos.X))
	}
}

func TestSpellTeleportToTargetUpdate530B70Gates(t *testing.T) {
	caster, target := &Object{}, &Object{}
	for _, tc := range []struct {
		name   string
		record DurSpell
		want   int32
	}{
		{"nil target", DurSpell{Caster16: caster}, 1},
		{"nil caster", DurSpell{Target48: target}, 1},
		{"dead target", DurSpell{Caster16: caster, Target48: &Object{ObjFlags: object.FlagDead}}, 1},
		{"destroyed caster", DurSpell{Caster16: &Object{ObjFlags: object.FlagDestroyed}, Target48: target}, 1},
		{"wrong frame", DurSpell{Caster16: caster, Target48: target, Frame68: 103}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SpellTeleportToTargetUpdate530B70(&tc.record,
				SpellTeleportToTargetRuntime530A30{Frame: func() uint32 { return 100 }}); got != tc.want {
				t.Fatalf("update = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSpellTeleportToTargetUpdate530B70Anchored(t *testing.T) {
	caster := &Object{}
	target := &Object{Buffs: 1 << ENCHANT_ANCHORED}
	var heard bool
	rt := SpellTeleportToTargetRuntime530A30{
		Frame: func() uint32 { return 100 },
		Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			if id != 231 || obj != target || kind != 0 || code != 0 {
				t.Fatalf("audio = %d %p %d %d", id, obj, kind, code)
			}
			heard = true
		},
	}
	if got := SpellTeleportToTargetUpdate530B70(&DurSpell{Caster16: caster, Target48: target, Frame68: 101}, rt); got != 1 || !heard {
		t.Fatalf("anchored update = %d, heard = %t", got, heard)
	}
}

func TestSpellTeleportToTargetUpdate530B70Effects(t *testing.T) {
	for _, tc := range []struct {
		name      string
		invisible bool
		player    bool
		wantCalls []string
	}{
		{"visible", false, false, []string{"fx", "cast sound", "audio 0", "new wake", "create wake", "teleport", "fx", "audio 0", "attribution"}},
		{"invisible monster", true, false, []string{"fx", "cast sound", "audio 0", "new wake", "create wake", "teleport", "attribution"}},
		{"invisible player", true, true, []string{"fx", "cast sound", "audio 0", "new wake", "create wake", "teleport", "audio 2", "attribution"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caster := &Object{}
			target := &Object{PosVec: types.Ptf(10, 20), NetCode: 0x1234}
			if tc.invisible {
				target.Buffs = 1 << ENCHANT_INVISIBLE
			}
			if tc.player {
				target.ObjClass = object.ClassPlayer
			}
			wakeData := &TeleportWakeCollideData{}
			wake := &Object{CollideData: unsafe.Pointer(wakeData)}
			record := &DurSpell{Caster16: caster, Target48: target, Spell: uint32(spell.SPELL_TELEPORT_TO_TARGET),
				Pos2: types.Ptf(30, 40), Pos: types.Ptf(math.Float32frombits(0x3fdccccc), 0), Frame68: 101}
			var calls []string
			rt := SpellTeleportToTargetRuntime530A30{
				Frame:    func() uint32 { return 100 },
				TickRate: func() uint32 { return 30 },
				NewObject: func(name string) *Object {
					calls = append(calls, "new wake")
					if name != "TeleportWake" {
						t.Fatalf("new object = %q", name)
					}
					return wake
				},
				CreateAt: func(obj, owner *Object, pos types.Pointf) {
					calls = append(calls, "create wake")
					if obj != wake || owner != target || pos != (types.Ptf(10, 20)) || wakeData.Destination != record.Pos2 {
						t.Fatalf("wake = %p %p %v %v", obj, owner, pos, wakeData.Destination)
					}
				},
				SendPointFX: func(op netmsg.Op, pos types.Pointf) {
					calls = append(calls, "fx")
					if op != netmsg.MSG_FX_TELEPORT || pos != target.PosVec {
						t.Fatalf("FX = %d %v", op, pos)
					}
				},
				CastSound: func(id spell.ID) sound.ID {
					calls = append(calls, "cast sound")
					if id != spell.SPELL_TELEPORT_TO_TARGET {
						t.Fatalf("spell = %d", id)
					}
					return 97
				},
				Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
					if id != 97 || obj != target || code != map[bool]uint32{true: target.NetCode, false: 0}[kind == 2] {
						t.Fatalf("audio = %d %p %d %d", id, obj, kind, code)
					}
					if kind == 2 {
						calls = append(calls, "audio 2")
					} else {
						calls = append(calls, "audio 0")
					}
				},
				Teleport: func(obj *Object, pos types.Pointf) {
					calls = append(calls, "teleport")
					if obj != target || pos != record.Pos2 {
						t.Fatalf("teleport = %p %v", obj, pos)
					}
					obj.PosVec = pos
				},
				Attribution: func(from, to *Object) {
					calls = append(calls, "attribution")
					if from != caster || to != target {
						t.Fatal("wrong attribution")
					}
				},
			}
			if got := SpellTeleportToTargetUpdate530B70(record, rt); got != 1 || target.PosVec != record.Pos2 ||
				wakeData.Destination != record.Pos2 || wake.Field34 != 130 || math.Float32bits(record.Pos.X) != 0x3fdccccc {
				t.Fatalf("result = %d, target = %v, wake = %v/%d", got, target.PosVec, wakeData.Destination, wake.Field34)
			}
			if !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, tc.wantCalls)
			}
		})
	}
}
