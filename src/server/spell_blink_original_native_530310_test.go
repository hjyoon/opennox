package server

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

type blinkOriginalInput530310 struct {
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Quest       bool      `json:"quest"`
	Coop        bool      `json:"coop"`
	Flag        uint32    `json:"flag"`
	Level       uint32    `json:"level"`
	Timer       uint32    `json:"timer"`
	Frame       uint32    `json:"frame"`
	Tick        uint32    `json:"tick"`
	Source      int       `json:"source"`
	Caster      int       `json:"caster"`
	Target      int       `json:"target"`
	Classes     [6]uint32 `json:"classes"`
	Flags       [6]uint32 `json:"flags"`
	Buffs       [6]uint32 `json:"buffs"`
	Gates       [2]int    `json:"gates"`
	Cache       uint32    `json:"cache"`
	Lookup      uint32    `json:"lookup"`
	Wake        bool      `json:"wake"`
	Alias       bool      `json:"alias"`
	Waypoint    int       `json:"waypoint"`
	Destination [2]uint32 `json:"destination"`
	Delay       string    `json:"delay_bits64"`
	Sound       uint32    `json:"sound"`
	Mutation    struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}
type blinkOriginalFixture530310 struct {
	OriginalSHA string `json:"original_sha256"`
	Control     uint32 `json:"control_word"`
	Rows        []struct {
		Input  blinkOriginalInput530310 `json:"input"`
		Result json.RawMessage          `json:"result"`
	} `json:"rows"`
}

func blinkOriginalFixtureRead530310(t *testing.T) blinkOriginalFixture530310 {
	t.Helper()
	raw, err := os.ReadFile("testdata/spell_blink_native_original_530310.json")
	if err != nil {
		t.Fatal(err)
	}
	var f blinkOriginalFixture530310
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	if f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Control != 0xe3f || len(f.Rows) != 69 {
		t.Fatal("original Blink identity/profile/count")
	}
	return f
}
func TestSpellBlinkOriginalNative530310(t *testing.T) {
	f := blinkOriginalFixtureRead530310(t)
	count := 0
	for _, row := range f.Rows {
		if row.Input.Kind == "wake" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			blinkOriginalCompare530310(t, row.Result, blinkOriginalRun530310(t, row.Input, nil))
		})
	}
	if count != 58 {
		t.Fatal("incomplete original Blink root inputs", count)
	}
}
func blinkOriginalCompare530310(t *testing.T, wantRaw, actual json.RawMessage) {
	t.Helper()
	var got, want any
	if json.Unmarshal(actual, &got) != nil || json.Unmarshal(wantRaw, &want) != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("original native Blink mismatch\ngot %s\nwant %s", actual, wantRaw)
	}
}
func blinkOriginalRun530310(t *testing.T, input blinkOriginalInput530310, wakeHelper func(*Object, *types.Pointf, *types.Pointf, SpellBlinkRuntime530310)) json.RawMessage {
	t.Helper()
	var blocks []*tagOriginalBlock530160
	record, rb := tagOriginalAllocate530160[DurSpell](t, &blocks)
	var objects [6]*Object
	var ob [6]*tagOriginalBlock530160
	positions := [6]types.Pointf{types.Ptf(7, 11), types.Ptf(13, 17), types.Ptf(123.5, -456.25), types.Ptf(88.25, 99.5), {}, types.Ptf(22, 33)}
	for i := range objects {
		objects[i], ob[i] = tagOriginalAllocate530160[Object](t, &blocks)
		*objects[i] = Object{ObjClass: object.Class(input.Classes[i]), ObjFlags: object.Flags(input.Flags[i]), Buffs: input.Buffs[i], NetCode: 0x12340000 + uint32(i), PosVec: positions[i]}
	}
	obj := func(i int) *Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	var updates [2]*PlayerUpdateData
	var ub [2]*tagOriginalBlock530160
	for i := range updates {
		updates[i], ub[i] = tagOriginalAllocate530160[PlayerUpdateData](t, &blocks)
		updates[i].SoulGate = obj(input.Gates[i])
		objects[i+2].UpdateData = unsafe.Pointer(updates[i])
	}
	var data [2]*TeleportWakeCollideData
	var db [2]*tagOriginalBlock530160
	for i := range data {
		data[i], db[i] = tagOriginalAllocate530160[TeleportWakeCollideData](t, &blocks)
		data[i].Destination = types.Ptf(-1, -2)
		db[i].allow(unsafe.Pointer(&data[i].Destination), unsafe.Sizeof(data[i].Destination))
	}
	objects[4].CollideData = unsafe.Pointer(data[0])
	objects[4].Field34 = 0x13579bdf
	dest, _ := tagOriginalAllocate530160[types.Pointf](t, &blocks)
	*dest = types.Ptf(math.Float32frombits(input.Destination[0]), math.Float32frombits(input.Destination[1]))
	cache, _ := tagOriginalAllocate530160[uint32](t, &blocks)
	*cache = input.Cache
	*record = DurSpell{ID: 0xabcd, Spell: 4, Level: input.Level, Obj12: obj(input.Source), Caster16: obj(input.Caster), Flag20: input.Flag, Target48: obj(input.Target), Frame68: input.Timer, Field36: 0x13579bdf, Flags88: 0x77556601}
	// Every allowed mutable field is also copied into every observed state;
	// all other native bytes and all surrounding canaries remain immutable.
	for _, field := range []struct {
		p unsafe.Pointer
		n uintptr
	}{{unsafe.Pointer(&record.Obj12), unsafe.Sizeof(record.Obj12)}, {unsafe.Pointer(&record.Caster16), unsafe.Sizeof(record.Caster16)}, {unsafe.Pointer(&record.Target48), unsafe.Sizeof(record.Target48)}, {unsafe.Pointer(&record.Spell), unsafe.Sizeof(record.Spell)}, {unsafe.Pointer(&record.Frame68), unsafe.Sizeof(record.Frame68)}} {
		rb.allow(field.p, field.n)
	}
	for i, o := range objects {
		ob[i].allow(unsafe.Pointer(&o.PosVec), unsafe.Sizeof(o.PosVec))
		ob[i].allow(unsafe.Pointer(&o.Buffs), unsafe.Sizeof(o.Buffs))
	}
	ob[4].allow(unsafe.Pointer(&objects[4].Field34), unsafe.Sizeof(objects[4].Field34))
	ob[4].allow(unsafe.Pointer(&objects[4].CollideData), unsafe.Sizeof(objects[4].CollideData))
	ub[0].allow(unsafe.Pointer(&updates[0].SoulGate), unsafe.Sizeof(updates[0].SoulGate))
	// Cache and caller destination have explicit original writes or mutations.
	for _, b := range blocks {
		if uintptr(unsafe.Pointer(cache)) >= uintptr(b.pointer) && uintptr(unsafe.Pointer(cache)) < uintptr(b.pointer)+uintptr(b.size) {
			b.allow(unsafe.Pointer(cache), unsafe.Sizeof(*cache))
		}
		if uintptr(unsafe.Pointer(dest)) >= uintptr(b.pointer) && uintptr(unsafe.Pointer(dest)) < uintptr(b.pointer)+uintptr(b.size) {
			b.allow(unsafe.Pointer(dest), unsafe.Sizeof(*dest))
		}
	}
	frame, tick := input.Frame, input.Tick
	index := func(p *Object) int {
		if p == nil {
			return -1
		}
		for i, o := range objects {
			if p == o {
				return i
			}
		}
		t.Fatal("unknown native object")
		return -2
	}
	words := func(p types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	state := func() map[string]any {
		var pos [6][2]uint32
		var cls, flags, buffs [6]uint32
		for i, o := range objects {
			pos[i] = words(o.PosVec)
			cls[i], flags[i], buffs[i] = uint32(o.ObjClass), uint32(o.ObjFlags), o.Buffs
		}
		dataIndex := -1
		for i, d := range data {
			if unsafe.Pointer(d) == objects[4].CollideData {
				dataIndex = i
			}
		}
		if dataIndex < 0 {
			t.Fatal("unknown collision data")
		}
		return map[string]any{"source": index(record.Obj12), "caster": index(record.Caster16), "target": index(record.Target48), "spell": record.Spell, "level": record.Level, "flag": record.Flag20, "timer": record.Frame68, "frame": frame, "tick": tick, "cache": *cache, "positions": pos, "classes": cls, "flags": flags, "buffs": buffs, "gates": [2]int{index(updates[0].SoulGate), index(updates[1].SoulGate)}, "wake_expiry": objects[4].Field34, "wake_data": dataIndex, "wake_destinations": [2][2]uint32{words(data[0].Destination), words(data[1].Destination)}, "destination": words(*dest)}
	}
	calls := map[string]int{}
	mutated := false
	mutate := func(kind string) {
		at := calls[kind]
		calls[kind]++
		m := input.Mutation
		if mutated || m.Stage != kind || m.At != at {
			return
		}
		mutated = true
		switch m.Action {
		case "target-b":
			record.Target48 = objects[3]
		case "caster-b":
			record.Caster16 = objects[0]
		case "source-b":
			record.Obj12 = objects[0]
		case "spell":
			record.Spell = 43
		case "position-a":
			objects[2].PosVec = types.Ptf(-17.5, 23.25)
		case "position-b":
			objects[3].PosVec = types.Ptf(-77.5, 83.25)
		case "invisible-a":
			objects[2].Buffs |= 1
		case "visible-a":
			objects[2].Buffs &^= 1
		case "collision-b":
			objects[4].CollideData = unsafe.Pointer(data[1])
		case "frame":
			frame = 0xfffffffd
		case "tick":
			tick = 73
		case "gate-b":
			updates[0].SoulGate = objects[3]
		case "destination":
			*dest = types.Ptf(-27.5, 33.25)
		default:
			t.Fatal("unknown mutation", m.Action)
		}
	}
	events := make([]map[string]any, 0)
	event := func(kind string, args []any) {
		events = append(events, map[string]any{"kind": kind, "args": args, "state": state()})
		mutate(kind)
	}
	destination := types.Ptf(math.Float32frombits(input.Destination[0]), math.Float32frombits(input.Destination[1]))
	next := func(id uint32) *Object {
		event("new", []any{id})
		if !input.Wake {
			return nil
		}
		return objects[4]
	}
	rt := SpellBlinkRuntime530310{
		QuestMode: func() bool { return input.Quest }, CoopMode: func() bool { return input.Coop }, Frame: func() uint32 { return frame }, TickRate: func() uint32 { return tick },
		Waypoint: func(out *types.Pointf, target *Object) int {
			event("waypoint", []any{index(target)})
			*out = destination
			return input.Waypoint
		},
		PlayerStart: func(target *Object) types.Pointf { event("start", []any{index(target)}); return destination },
		RandomReachable: func(radius float32, center types.Pointf) types.Pointf {
			p := words(center)
			event("random", []any{math.Float32bits(radius), p[0], p[1]})
			return destination
		},
		NewObject: func(name string) *Object {
			if name != "TeleportWake" {
				t.Fatal("wake name", name)
			}
			return next(*cache)
		},
		CreateAt: func(wake, owner *Object, at types.Pointf) {
			p := words(at)
			event("create", []any{index(wake), index(owner), p[0], p[1], 0})
			wake.PosVec = at
		},
		SendPointFX: func(code netmsg.Op, at types.Pointf) { p := words(at); event("fx", []any{code, p[0], p[1]}) },
		CastSound: func(id spell.ID) sound.ID {
			event("sound", []any{id, 0})
			return sound.ID(input.Sound + uint32(calls["sound"]-1))
		},
		Audio: func(id sound.ID, target *Object, kind int, code uint32) {
			event("audio", []any{id, index(target), kind, code})
		},
		Teleport: func(target *Object, at types.Pointf) {
			p := words(at)
			event("teleport", []any{index(target), p[0], p[1]})
			target.PosVec = at
		},
		Attribution: func(caster, target *Object) { event("attribution", []any{index(caster), index(target)}) },
	}
	raw, err := hex.DecodeString(input.Delay)
	if err != nil || len(raw) != 8 {
		t.Fatal("delay binary64", err)
	}
	value := math.Float64frombits(binary.LittleEndian.Uint64(raw))
	// Bind the declared balance service to the actual runtime return width.
	// This exposes the old binary32 API loss before its binary64 correction.
	delay := reflect.ValueOf(&rt).Elem().FieldByName("TeleportDelay")
	typ := delay.Type()
	delay.Set(reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		event("delay", []any{uint32(args[0].Uint())})
		return []reflect.Value{reflect.ValueOf(value).Convert(typ.Out(0))}
	}))
	// The reached original helper's typed cache/allocation binding is new;
	// the warm-cache root before fixture still uses its existing name factory.
	fields := reflect.ValueOf(&rt).Elem()
	if f := fields.FieldByName("WakeType"); f.IsValid() {
		f.Set(reflect.ValueOf(cache))
		fields.FieldByName("LookupType").Set(reflect.ValueOf(func(name string) uint32 {
			if name != "TeleportWake" {
				t.Fatal("lookup", name)
			}
			event("lookup", []any{})
			return input.Lookup
		}))
		fields.FieldByName("NewTypedObject").Set(reflect.ValueOf(next))
	}
	for _, b := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(b.pointer) <= 0xffffffff {
			t.Fatal("native Blink frame below4GiB")
		}
		b.before = append([]byte(nil), unsafe.Slice((*byte)(b.pointer), b.size)...)
	}
	var result any
	switch input.Kind {
	case "create":
		result = SpellBlinkCreate530310(record, rt)
	case "update":
		result = SpellBlinkUpdate530380(record, rt)
	case "wake":
		if wakeHelper == nil {
			t.Fatal("wake helper not bound")
		}
		destination := dest
		if input.Alias {
			destination = &objects[2].PosVec
		}
		wakeHelper(objects[2], &objects[2].PosVec, destination, rt)
	default:
		t.Fatal("unknown root")
	}
	if input.Mutation.Stage != "" && !mutated {
		t.Fatal("declared mutation not reached")
	}
	for n, b := range blocks {
		after := unsafe.Slice((*byte)(b.pointer), b.size)
		for i, v := range b.before {
			if v != after[i] && !b.allowed[i] {
				t.Fatalf("Blink native frame%d nonoutput/canary byte%d changed", n, i)
			}
		}
	}
	resultRaw, err := json.Marshal(map[string]any{"ret": result, "state": state(), "events": events})
	if err != nil {
		t.Fatal(err)
	}
	return resultRaw
}
