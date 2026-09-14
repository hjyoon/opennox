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

func TestSpellBlinkCreate530310Modes(t *testing.T) {
	for _, tc := range []struct {
		name      string
		quest     bool
		coop      bool
		flag20    uint32
		want      int32
		wantFrame uint32
		wantCalls []string
	}{
		{"quest glyph", true, true, 1, 1, 77, []string{"quest"}},
		{"arena", false, false, 0, 0, 101, []string{"quest", "coop", "frame"}},
		{"coop round-to-even", false, true, 0, 0, 102, []string{"quest", "coop", "delay", "frame"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			record := &DurSpell{Level: 3, Flag20: tc.flag20, Frame68: 77}
			runtime := SpellBlinkRuntime530310{
				QuestMode: func() bool { calls = append(calls, "quest"); return tc.quest },
				CoopMode:  func() bool { calls = append(calls, "coop"); return tc.coop },
				TeleportDelay: func(level uint32) float32 {
					calls = append(calls, "delay")
					if level != 2 {
						t.Fatalf("level index = %d", level)
					}
					return 2.5
				},
				Frame: func() uint32 { calls = append(calls, "frame"); return 100 },
			}
			if got := SpellBlinkCreate530310(record, runtime); got != tc.want || record.Frame68 != tc.wantFrame {
				t.Fatalf("create = %d, frame = %d; want %d, %d", got, record.Frame68, tc.want, tc.wantFrame)
			}
			if !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, tc.wantCalls)
			}
		})
	}
}

func TestSpellBlinkUpdate530380EarlyGates(t *testing.T) {
	target := &Object{ObjClass: object.ClassPlayer}
	runtime := SpellBlinkRuntime530310{
		Frame: func() uint32 { return 100 },
		Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			if id != 231 || obj != target || kind != 0 || code != 0 {
				t.Fatalf("anchor audio = %d/%p/%d/%d", id, obj, kind, code)
			}
		},
	}
	for _, tc := range []struct {
		name   string
		target *Object
		frame  uint32
		want   int32
	}{
		{"no target", nil, 101, 1},
		{"destroyed target", &Object{ObjFlags: object.FlagDestroyed}, 101, 1},
		{"wait", target, 102, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SpellBlinkUpdate530380(&DurSpell{Target48: tc.target, Frame68: tc.frame}, runtime); got != tc.want {
				t.Fatalf("update = %d, want %d", got, tc.want)
			}
		})
	}
	target.Buffs = 1 << ENCHANT_ANCHORED
	if got := SpellBlinkUpdate530380(&DurSpell{Target48: target, Frame68: 101}, runtime); got != 1 {
		t.Fatalf("anchored update = %d, want 1", got)
	}
}

func TestSpellBlinkUpdate530380VisibleAndNativePointers(t *testing.T) {
	origin, destination := types.Ptf(123.5, -456.25), types.Ptf(789.5, 234.25)
	target := &Object{ObjClass: object.ClassPlayer, PosVec: origin, NetCode: 0xdeadbeef}
	owner := &Object{ObjClass: object.ClassPlayer}
	record := &DurSpell{Spell: uint32(spell.SPELL_BLINK), Obj12: owner, Target48: target, Frame68: 101}
	wakeData := &TeleportWakeCollideData{}
	wake := &Object{CollideData: unsafe.Pointer(wakeData)}
	var events []string
	var pointFX []types.Pointf
	runtime := SpellBlinkRuntime530310{
		Frame: func() uint32 { return 100 }, TickRate: func() uint32 { return 30 },
		QuestMode: func() bool { return false },
		Waypoint: func(out *types.Pointf, obj *Object) int {
			if obj != target {
				t.Fatalf("waypoint target = %p", obj)
			}
			*out = destination
			events = append(events, "waypoint")
			return 1
		},
		NewObject: func(name string) *Object {
			if name != "TeleportWake" {
				t.Fatalf("new object = %q", name)
			}
			events = append(events, "new wake")
			return wake
		},
		CreateAt: func(obj, source *Object, at types.Pointf) {
			if obj != wake || source != target || at != origin || wakeData.Destination != destination {
				t.Fatalf("wake create = %p/%p/%v, destination %v", obj, source, at, wakeData.Destination)
			}
			events = append(events, "create wake")
		},
		SendPointFX: func(effect netmsg.Op, at types.Pointf) {
			if effect != netmsg.MSG_FX_TELEPORT {
				t.Fatalf("point FX = %d", effect)
			}
			pointFX = append(pointFX, at)
			events = append(events, "fx")
		},
		CastSound: func(id spell.ID) sound.ID {
			if id != spell.SPELL_BLINK {
				t.Fatalf("sound spell = %d", id)
			}
			return sound.SoundBlinkCast
		},
		Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			if id != sound.SoundBlinkCast || obj != target || kind != 0 || code != 0 {
				t.Fatalf("audio = %d/%p/%d/%d", id, obj, kind, code)
			}
			events = append(events, "audio")
		},
		Teleport: func(obj *Object, at types.Pointf) {
			if obj != target || at != destination {
				t.Fatalf("teleport = %p/%v", obj, at)
			}
			obj.PosVec = at
			events = append(events, "teleport")
		},
		Attribution: func(source, obj *Object) {
			if source != owner || obj != target {
				t.Fatalf("attribution = %p/%p, want %p/%p", source, obj, owner, target)
			}
			events = append(events, "attribution")
		},
	}
	if got := SpellBlinkUpdate530380(record, runtime); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if wake.Field34 != 130 || wakeData.Destination != destination || target.PosVec != destination {
		t.Fatalf("post-update state: wake frame=%d dest=%v player=%v", wake.Field34, wakeData.Destination, target.PosVec)
	}
	if want := []types.Pointf{origin, origin, destination, destination}; !reflect.DeepEqual(pointFX, want) {
		t.Fatalf("point FX = %v, want %v", pointFX, want)
	}
	if want := []string{"waypoint", "new wake", "create wake", "fx", "audio", "fx", "fx", "teleport", "fx", "audio", "attribution"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellBlinkUpdate530380QuestSoulGateInvisible(t *testing.T) {
	gate := &Object{PosVec: types.Ptf(22, 33)}
	update := &PlayerUpdateData{SoulGate: gate}
	target := &Object{ObjClass: object.ClassPlayer, Buffs: 1 << ENCHANT_INVISIBLE, UpdateData: unsafe.Pointer(update), PosVec: types.Ptf(1, 2), NetCode: 0x12345678}
	wakeData := &TeleportWakeCollideData{}
	wake := &Object{CollideData: unsafe.Pointer(wakeData)}
	var fx, audio int
	runtime := SpellBlinkRuntime530310{
		Frame: func() uint32 { return 10 }, TickRate: func() uint32 { return 30 },
		QuestMode: func() bool { return true },
		RandomReachable: func(radius float32, center types.Pointf) types.Pointf {
			if radius != 60 || center != gate.PosVec {
				t.Fatalf("random point = %f/%v", radius, center)
			}
			return types.Ptf(44, 55)
		},
		NewObject: func(string) *Object { return wake },
		CreateAt: func(obj, source *Object, at types.Pointf) {
			if obj != wake || source != target || at != (types.Ptf(1, 2)) {
				t.Fatalf("wake create = %p/%p/%v", obj, source, at)
			}
		},
		SendPointFX: func(op netmsg.Op, point types.Pointf) { fx++ },
		CastSound:   func(spell.ID) sound.ID { return sound.SoundBlinkCast },
		Audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			audio++
			if audio == 2 && (kind != 2 || code != target.NetCode) {
				t.Fatalf("invisible directed audio = %d/%d", kind, code)
			}
		},
		Teleport:    func(obj *Object, point types.Pointf) { obj.PosVec = point },
		Attribution: func(*Object, *Object) {},
	}
	if got := SpellBlinkUpdate530380(&DurSpell{Target48: target, Frame68: 11}, runtime); got != 1 {
		t.Fatalf("update = %d", got)
	}
	if fx != 1 || audio != 2 || wakeData.Destination != (types.Ptf(44, 55)) || target.PosVec != (types.Ptf(44, 55)) {
		t.Fatalf("quest invisible state: fx=%d audio=%d wake=%v target=%v", fx, audio, wakeData.Destination, target.PosVec)
	}
}
