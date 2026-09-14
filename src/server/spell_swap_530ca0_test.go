package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestSpellSwapCreate530CA0Gates(t *testing.T) {
	caster := &Object{ObjClass: object.ClassPlayer}
	target := &Object{ObjClass: object.ClassMonster}
	for _, tc := range []struct {
		name   string
		record DurSpell
	}{
		{"nil caster", DurSpell{Target48: target}},
		{"pending flag low byte", DurSpell{Caster16: caster, Target48: target, Flags88: 0x20}},
		{"nil target", DurSpell{Caster16: caster}},
		{"dead target", DurSpell{Caster16: caster, Target48: &Object{ObjFlags: object.FlagDead}}},
		{"destroyed target", DurSpell{Caster16: caster, Target48: &Object{ObjFlags: object.FlagDestroyed}}},
		{"same object", DurSpell{Caster16: caster, Target48: caster}},
		{"untargetable monster", DurSpell{Caster16: caster, Target48: &Object{ObjClass: object.ClassMonster, ObjSubClass: 0x4000}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.record.Frame68 = 77
			if got := SpellSwapCreate530CA0(&tc.record, SpellSwapRuntime530CA0{}); got != 1 || tc.record.Frame68 != 77 {
				t.Fatalf("create = %d, frame = %d; want 1, 77", got, tc.record.Frame68)
			}
		})
	}
}

func TestSpellSwapCreate530CA0Delay(t *testing.T) {
	caster, target := &Object{}, &Object{ObjClass: object.ClassMonster}
	for _, tc := range []struct {
		name      string
		coop      bool
		delay     float32
		wantFrame uint32
		wantCalls []string
	}{
		{"arena", false, 2.5, 101, []string{"coop", "frame"}},
		{"coop ties to even", true, 2.5, 102, []string{"coop", "delay", "frame"}},
		{"coop rounds up", true, 3.5, 104, []string{"coop", "delay", "frame"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			record := &DurSpell{Caster16: caster, Target48: target, Level: 3, Flags88: 0x2000}
			rt := SpellSwapRuntime530CA0{
				CoopMode: func() bool { calls = append(calls, "coop"); return tc.coop },
				Frame:    func() uint32 { calls = append(calls, "frame"); return 100 },
				TeleportDelay: func(level uint32) float32 {
					calls = append(calls, "delay")
					if level != 2 {
						t.Fatalf("level index = %d, want 2", level)
					}
					return tc.delay
				},
			}
			if got := SpellSwapCreate530CA0(record, rt); got != 0 || record.Frame68 != tc.wantFrame {
				t.Fatalf("create = %d, frame = %d; want 0, %d", got, record.Frame68, tc.wantFrame)
			}
			if !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, tc.wantCalls)
			}
		})
	}
}

func TestSpellSwapUpdate530D30EarlyGates(t *testing.T) {
	caster, target := &Object{}, &Object{}
	for _, tc := range []struct {
		name   string
		record DurSpell
		want   int32
	}{
		{"no target", DurSpell{Caster16: caster}, 1},
		{"no caster", DurSpell{Target48: target}, 1},
		{"dead target", DurSpell{Caster16: caster, Target48: &Object{ObjFlags: object.FlagDead}}, 1},
		{"destroyed caster", DurSpell{Caster16: &Object{ObjFlags: object.FlagDestroyed}, Target48: target}, 1},
		{"wrong frame", DurSpell{Caster16: caster, Target48: target, Frame68: 103}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt := SpellSwapRuntime530CA0{Frame: func() uint32 { return 100 }}
			if got := SpellSwapUpdate530D30(&tc.record, rt); got != tc.want {
				t.Fatalf("update = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSpellSwapUpdate530D30Anchored(t *testing.T) {
	for _, anchorCaster := range []bool{false, true} {
		caster, target := &Object{}, &Object{}
		if anchorCaster {
			caster.Buffs = 1 << ENCHANT_ANCHORED
		} else {
			target.Buffs = 1 << ENCHANT_ANCHORED
		}
		var heard []*Object
		rt := SpellSwapRuntime530CA0{
			Frame: func() uint32 { return 100 },
			Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
				if id != 231 || kind != 0 || code != 0 {
					t.Fatalf("audio = %d/%d/%d", id, kind, code)
				}
				heard = append(heard, obj)
			},
		}
		if got := SpellSwapUpdate530D30(&DurSpell{Caster16: caster, Target48: target, Frame68: 101}, rt); got != 1 {
			t.Fatalf("update = %d, want 1", got)
		}
		if want := []*Object{target, caster}; !reflect.DeepEqual(heard, want) {
			t.Fatalf("audio targets = %v, want %v", heard, want)
		}
	}
}

func TestSpellSwapUpdate530D30InteractionAndPlayerBounds(t *testing.T) {
	caster := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(100.25, -100.5)}
	player := &Player{Field10: 10, Field12: 20}
	data := &PlayerUpdateData{Player: player}
	caster.UpdateData = unsafe.Pointer(data)
	target := &Object{PosVec: types.Ptf(110.25, -80.5)} // inclusive upper bounds
	for _, tc := range []struct {
		name      string
		flag20    uint32
		interacts bool
		position  types.Pointf
		wantInfo  bool
		wantSwap  bool
	}{
		{"blocked interaction", 0, false, target.PosVec, true, false},
		{"outside player bounds", 0, true, types.Ptf(110.5, -80.5), true, false},
		{"inside player bounds", 0, true, target.PosVec, false, true},
		{"glyph skips interaction and bounds", 1, false, types.Ptf(1000, 1000), false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caster.PosVec = types.Ptf(100.25, -100.5)
			target.PosVec = tc.position
			var informed, teleported bool
			rt := SpellSwapRuntime530CA0{
				Frame:       func() uint32 { return 100 },
				CanInteract: func(from, to *Object) bool { return tc.interacts },
				InformNoLOS: func(who *Object) {
					if who != caster {
						t.Fatalf("informed %p, want %p", who, caster)
					}
					informed = true
				},
				Teleport:    func(obj *Object, point types.Pointf) { teleported = true; obj.PosVec = point },
				SendPointFX: func(netmsg.Op, types.Pointf) {},
				CastSound:   func(spell.ID) sound.ID { return sound.SoundBlinkCast },
				Audio:       func(sound.ID, *Object, int, uint32) {},
				Attribution: func(*Object, *Object) {},
			}
			if got := SpellSwapUpdate530D30(&DurSpell{Caster16: caster, Target48: target, Frame68: 101, Flag20: tc.flag20}, rt); got != 1 {
				t.Fatalf("update = %d, want 1", got)
			}
			if informed != tc.wantInfo || teleported != tc.wantSwap {
				t.Fatalf("informed/teleported = %t/%t, want %t/%t", informed, teleported, tc.wantInfo, tc.wantSwap)
			}
		})
	}
}

func TestSpellSwapUpdate530D30EffectsAndNativePointers(t *testing.T) {
	for _, invisible := range []bool{false, true} {
		caster := &Object{PosVec: types.Ptf(10, 20)}
		target := &Object{PosVec: types.Ptf(30, 40)}
		if invisible {
			target.Buffs = 1 << ENCHANT_INVISIBLE
		}
		record := &DurSpell{Caster16: caster, Target48: target, Spell: uint32(spell.SPELL_SWAP), Flag20: 1, Frame68: 101}
		var events []string
		var pointFX []types.Pointf
		rt := SpellSwapRuntime530CA0{
			Frame: func() uint32 { return 100 },
			Teleport: func(obj *Object, pos types.Pointf) {
				if obj == target {
					events = append(events, "teleport target")
				} else if obj == caster {
					events = append(events, "teleport caster")
				} else {
					t.Fatal("unexpected teleport object")
				}
				obj.PosVec = pos
			},
			SendPointFX: func(op netmsg.Op, point types.Pointf) {
				if op != netmsg.MSG_FX_TELEPORT {
					t.Fatalf("FX opcode = %d", op)
				}
				pointFX = append(pointFX, point)
				events = append(events, "fx")
			},
			CastSound: func(id spell.ID) sound.ID {
				if id != spell.SPELL_SWAP {
					t.Fatalf("spell sound ID = %d", id)
				}
				events = append(events, "cast sound")
				return sound.SoundBlinkCast
			},
			Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
				if id != sound.SoundBlinkCast || kind != 0 || code != 0 {
					t.Fatalf("audio = %d/%d/%d", id, kind, code)
				}
				if obj == target {
					events = append(events, "audio target")
				} else if obj == caster {
					events = append(events, "audio caster")
				} else {
					t.Fatal("unexpected audio object")
				}
			},
			Attribution: func(from, to *Object) {
				if from != caster || to != target {
					t.Fatalf("attribution = %p/%p", from, to)
				}
				events = append(events, "attribution")
			},
		}
		if got := SpellSwapUpdate530D30(record, rt); got != 1 {
			t.Fatalf("update = %d, want 1", got)
		}
		if caster.PosVec != (types.Ptf(30, 40)) || target.PosVec != (types.Ptf(10, 20)) {
			t.Fatalf("positions = %v/%v, want (30,40)/(10,20)", caster.PosVec, target.PosVec)
		}
		wantFX := []types.Pointf(nil)
		wantEvents := []string{"teleport target", "teleport caster", "cast sound", "audio target", "cast sound", "audio caster", "attribution"}
		if !invisible {
			wantFX = []types.Pointf{types.Ptf(30, 40), types.Ptf(10, 20)}
			wantEvents = []string{"teleport target", "teleport caster", "fx", "fx", "cast sound", "audio target", "cast sound", "audio caster", "attribution"}
		}
		if !reflect.DeepEqual(pointFX, wantFX) || !reflect.DeepEqual(events, wantEvents) {
			t.Fatalf("point FX/events = %v/%v, want %v/%v", pointFX, events, wantFX, wantEvents)
		}
	}
}
