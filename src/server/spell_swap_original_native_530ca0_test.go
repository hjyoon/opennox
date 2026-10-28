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

type swapOriginalInput530CA0 struct {
	Name       string       `json:"name"`
	Kind       string       `json:"kind"`
	Coop       bool         `json:"coop"`
	Flag       uint32       `json:"flag"`
	Flags88    uint32       `json:"flags88"`
	Level      uint32       `json:"level"`
	Timer      uint32       `json:"timer"`
	Frame      uint32       `json:"frame"`
	Source     int          `json:"source"`
	Caster     int          `json:"caster"`
	Target     int          `json:"target"`
	Classes    [4]uint32    `json:"classes"`
	Subclasses [4]uint32    `json:"subclasses"`
	Flags      [4]uint32    `json:"flags"`
	Buffs      [4]uint32    `json:"buffs"`
	Positions  [4][2]uint32 `json:"positions"`
	Dimensions [4][2]uint16 `json:"dimensions"`
	Interacts  bool         `json:"interacts"`
	Delay      string       `json:"delay_bits64"`
	Sound      uint32       `json:"sound"`
	Mutation   struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}

func TestSpellSwapOriginalNative530CA0(t *testing.T) {
	raw, err := os.ReadFile("testdata/spell_swap_native_original_530ca0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA     string `json:"original_sha256"`
		Control uint32 `json:"control_word"`
		Rows    []struct {
			Input  swapOriginalInput530CA0 `json:"input"`
			Result json.RawMessage         `json:"result"`
		} `json:"rows"`
	}
	if json.Unmarshal(raw, &fixture) != nil || fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0xe3f || len(fixture.Rows) != 169 {
		t.Fatal("original Swap identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			var got, want any
			actual := swapOriginalRun530CA0(t, row.Input)
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original native Swap mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}

func swapOriginalRun530CA0(t *testing.T, input swapOriginalInput530CA0) json.RawMessage {
	t.Helper()
	var blocks []*tagOriginalBlock530160
	record, rb := tagOriginalAllocate530160[DurSpell](t, &blocks)
	var objects [4]*Object
	var updates [4]*PlayerUpdateData
	var players [4]*Player
	var ob, ub, pb [4]*tagOriginalBlock530160
	for i := range objects {
		objects[i], ob[i] = tagOriginalAllocate530160[Object](t, &blocks)
		updates[i], ub[i] = tagOriginalAllocate530160[PlayerUpdateData](t, &blocks)
		players[i], pb[i] = tagOriginalAllocate530160[Player](t, &blocks)
		*objects[i] = Object{ObjClass: object.Class(input.Classes[i]), ObjSubClass: object.SubClass(input.Subclasses[i]), ObjFlags: object.Flags(input.Flags[i]), Buffs: input.Buffs[i], PosVec: types.Ptf(math.Float32frombits(input.Positions[i][0]), math.Float32frombits(input.Positions[i][1])), UpdateData: unsafe.Pointer(updates[i])}
		updates[i].Player = players[i]
		players[i].Field10, players[i].Field12 = input.Dimensions[i][0], input.Dimensions[i][1]
		if input.Kind == "update" {
			ob[i].allow(unsafe.Pointer(&objects[i].PosVec), unsafe.Sizeof(objects[i].PosVec))
		}
	}
	obj := func(i int) *Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	*record = DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_SWAP), Level: input.Level, Obj12: obj(input.Source), Caster16: obj(input.Caster), Target48: obj(input.Target), Frame68: input.Timer, Flag20: input.Flag, Flags88: input.Flags88}
	if input.Kind == "create" {
		rb.allow(unsafe.Pointer(&record.Frame68), unsafe.Sizeof(record.Frame68))
	}
	switch input.Mutation.Action {
	case "target-b":
		rb.allow(unsafe.Pointer(&record.Target48), unsafe.Sizeof(record.Target48))
	case "caster-b":
		rb.allow(unsafe.Pointer(&record.Caster16), unsafe.Sizeof(record.Caster16))
	case "spell":
		rb.allow(unsafe.Pointer(&record.Spell), unsafe.Sizeof(record.Spell))
	case "invisible-target", "visible-target":
		ob[2].allow(unsafe.Pointer(&objects[2].Buffs), unsafe.Sizeof(objects[2].Buffs))
	case "invisible-caster":
		ob[1].allow(unsafe.Pointer(&objects[1].Buffs), unsafe.Sizeof(objects[1].Buffs))
	case "update-caster":
		ob[1].allow(unsafe.Pointer(&objects[1].UpdateData), unsafe.Sizeof(objects[1].UpdateData))
	case "player-caster":
		ub[1].allow(unsafe.Pointer(&updates[1].Player), unsafe.Sizeof(updates[1].Player))
	case "dimensions":
		pb[1].allow(unsafe.Pointer(&players[1].Field10), unsafe.Sizeof(players[1].Field10))
		pb[1].allow(unsafe.Pointer(&players[1].Field12), unsafe.Sizeof(players[1].Field12))
	case "", "frame", "position-target", "position-caster":
	default:
		t.Fatal("unknown mutation", input.Mutation)
	}
	index := func(p *Object) int {
		if p == nil {
			return -1
		}
		for i, o := range objects {
			if o == p {
				return i
			}
		}
		t.Fatal("unknown object")
		return -2
	}
	updateIndex := func(p unsafe.Pointer) int {
		for i, u := range updates {
			if unsafe.Pointer(u) == p {
				return i
			}
		}
		t.Fatal("unknown update data")
		return -2
	}
	playerIndex := func(p *Player) int {
		for i, v := range players {
			if p == v {
				return i
			}
		}
		t.Fatal("unknown player")
		return -2
	}
	words := func(p types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	frame := input.Frame
	state := func() map[string]any {
		var positions [4][2]uint32
		var classes, subclasses, flags, buffs [4]uint32
		var objectUpdates, updatePlayers [4]int
		var dimensions [4][2]uint16
		for i, o := range objects {
			positions[i], classes[i], subclasses[i], flags[i], buffs[i] = words(o.PosVec), uint32(o.ObjClass), uint32(o.ObjSubClass), uint32(o.ObjFlags), o.Buffs
			objectUpdates[i], updatePlayers[i] = updateIndex(o.UpdateData), playerIndex(updates[i].Player)
			dimensions[i] = [2]uint16{players[i].Field10, players[i].Field12}
		}
		return map[string]any{"source": index(record.Obj12), "caster": index(record.Caster16), "target": index(record.Target48), "spell": record.Spell, "level": record.Level, "flag": record.Flag20, "flags88": record.Flags88, "timer": record.Frame68, "frame": frame, "positions": positions, "classes": classes, "subclasses": subclasses, "flags": flags, "buffs": buffs, "object_updates": objectUpdates, "update_players": updatePlayers, "dimensions": dimensions}
	}
	calls := make(map[string]int)
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
		case "spell":
			record.Spell = 43
		case "position-target":
			objects[2].PosVec = types.Ptf(-17.5, 23.25)
		case "position-caster":
			objects[1].PosVec = types.Ptf(-77.5, 83.25)
		case "invisible-target":
			objects[2].Buffs |= 1
		case "invisible-caster":
			objects[1].Buffs |= 1
		case "visible-target":
			objects[2].Buffs &^= 1
		case "frame":
			frame = 0xfffffffd
		case "update-caster":
			objects[1].UpdateData = unsafe.Pointer(updates[0])
		case "player-caster":
			updates[1].Player = players[0]
		case "dimensions":
			players[1].Field10, players[1].Field12 = 65535, 1
		default:
			t.Fatal("unknown mutation")
		}
	}
	events := make([]map[string]any, 0)
	event := func(kind string, args []any) {
		events = append(events, map[string]any{"kind": kind, "args": args, "state": state()})
		mutate(kind)
	}
	rt := SpellSwapRuntime530CA0{
		CoopMode: func() bool { return input.Coop }, Frame: func() uint32 { return frame },
		CanInteract: func(caster, target *Object) bool {
			event("interact", []any{index(caster), index(target), 0})
			return input.Interacts
		},
		InformNoLOS: func(caster *Object) { event("inform", []any{index(caster)}) },
		Teleport: func(target *Object, at types.Pointf) {
			p := words(at)
			event("teleport", []any{index(target), p[0], p[1]})
			target.PosVec = at
		},
		SendPointFX: func(code netmsg.Op, at types.Pointf) { p := words(at); event("fx", []any{uint32(code), p[0], p[1]}) },
		CastSound: func(id spell.ID) sound.ID {
			event("sound", []any{uint32(id), 0})
			return sound.ID(input.Sound + uint32(calls["sound"]-1))
		},
		Audio: func(id sound.ID, target *Object, kind int, code uint32) {
			event("audio", []any{uint32(id), index(target), kind, code})
		},
		Attribution: func(caster, target *Object) { event("attribution", []any{index(caster), index(target)}) },
	}
	raw, err := hex.DecodeString(input.Delay)
	if err != nil || len(raw) != 8 {
		t.Fatal("balance binary64")
	}
	value := math.Float64frombits(binary.LittleEndian.Uint64(raw))
	// Follow the actual service return width so the unchanged fixture exposes
	// the old binary32 API loss before its binary64 correction.
	f := reflect.ValueOf(&rt).Elem().FieldByName("TeleportDelay")
	f.Set(reflect.MakeFunc(f.Type(), func(args []reflect.Value) []reflect.Value {
		event("delay", []any{uint32(args[0].Uint())})
		return []reflect.Value{reflect.ValueOf(value).Convert(f.Type().Out(0))}
	}))
	for _, b := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(b.pointer) <= 0xffffffff {
			t.Fatal("native Swap frame below4GiB")
		}
		b.before = append([]byte(nil), unsafe.Slice((*byte)(b.pointer), b.size)...)
	}
	var result int32
	switch input.Kind {
	case "create":
		result = SpellSwapCreate530CA0(record, rt)
	case "update":
		result = SpellSwapUpdate530D30(record, rt)
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
				t.Fatalf("Swap native frame%d nonoutput/canary byte%d changed", n, i)
			}
		}
	}
	out, err := json.Marshal(map[string]any{"ret": result, "state": state(), "events": events})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
