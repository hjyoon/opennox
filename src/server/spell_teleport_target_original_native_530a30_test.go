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

type teleportTargetOriginalInput530A30 struct {
	Name          string       `json:"name"`
	Kind          string       `json:"kind"`
	Coop          bool         `json:"coop"`
	Blocked       bool         `json:"blocked"`
	Clear         bool         `json:"clear"`
	Flag          uint32       `json:"flag"`
	Flags88       uint32       `json:"flags88"`
	Level         uint32       `json:"level"`
	Timer         uint32       `json:"timer"`
	Frame         uint32       `json:"frame"`
	Source        int          `json:"source"`
	Caster        int          `json:"caster"`
	Target        int          `json:"target"`
	Classes       [4]uint32    `json:"classes"`
	Flags         [4]uint32    `json:"flags"`
	Buffs         [4]uint32    `json:"buffs"`
	Netcodes      [4]uint32    `json:"netcodes"`
	Positions     [4][2]uint32 `json:"positions"`
	Destination   [2]uint32    `json:"destination"`
	ObjectUpdates [4]int       `json:"object_updates"`
	UpdatePlayers [4]int       `json:"update_players"`
	PlayerIndices [4]uint8     `json:"player_indices"`
	Delay         string       `json:"delay_bits64"`
	Sound         uint32       `json:"sound"`
	Mutation      struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}

func TestSpellTeleportTargetOriginalNative530A30(t *testing.T) {
	raw, err := os.ReadFile("testdata/spell_teleport_target_native_original_530a30.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA     string `json:"original_sha256"`
		Control uint32 `json:"control_word"`
		Rows    []struct {
			Input  teleportTargetOriginalInput530A30 `json:"input"`
			Result json.RawMessage                   `json:"result"`
		} `json:"rows"`
	}
	if json.Unmarshal(raw, &fixture) != nil || fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0xe3f || len(fixture.Rows) != 102 {
		t.Fatal("original TeleportToTarget identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			var got, want any
			actual := teleportTargetOriginalRun530A30(t, row.Input)
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original native TeleportToTarget mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}

func teleportTargetOriginalRun530A30(t *testing.T, input teleportTargetOriginalInput530A30) json.RawMessage {
	t.Helper()
	var blocks []*tagOriginalBlock530160
	record, rb := tagOriginalAllocate530160[DurSpell](t, &blocks)
	var objects [4]*Object
	var updates [4]*PlayerUpdateData
	var players [4]*Player
	var ob, ub [4]*tagOriginalBlock530160
	for i := range objects {
		objects[i], ob[i] = tagOriginalAllocate530160[Object](t, &blocks)
		updates[i], ub[i] = tagOriginalAllocate530160[PlayerUpdateData](t, &blocks)
		players[i], _ = tagOriginalAllocate530160[Player](t, &blocks)
	}
	obj := func(i int) *Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	for i := range objects {
		*objects[i] = Object{ObjClass: object.Class(input.Classes[i]), ObjFlags: object.Flags(input.Flags[i]), Buffs: input.Buffs[i], NetCode: input.Netcodes[i],
			PosVec: types.Ptf(math.Float32frombits(input.Positions[i][0]), math.Float32frombits(input.Positions[i][1]))}
		if input.ObjectUpdates[i] >= 0 {
			objects[i].UpdateData = unsafe.Pointer(updates[input.ObjectUpdates[i]])
		}
		updates[i].Player = players[input.UpdatePlayers[i]]
		players[i].PlayerInd = input.PlayerIndices[i]
		if input.Kind == "update" {
			ob[i].allow(unsafe.Pointer(&objects[i].PosVec), unsafe.Sizeof(objects[i].PosVec))
		}
	}
	*record = DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_TELEPORT_TO_TARGET), Level: input.Level, Obj12: obj(input.Source), Caster16: obj(input.Caster), Target48: obj(input.Target),
		Pos2: types.Ptf(math.Float32frombits(input.Destination[0]), math.Float32frombits(input.Destination[1])), Frame68: input.Timer, Flag20: input.Flag, Flags88: input.Flags88}
	if input.Kind == "create" {
		rb.allow(unsafe.Pointer(&record.Target48), unsafe.Sizeof(record.Target48))
		rb.allow(unsafe.Pointer(&record.Frame68), unsafe.Sizeof(record.Frame68))
	}
	switch input.Mutation.Action {
	case "target-b":
		rb.allow(unsafe.Pointer(&record.Target48), unsafe.Sizeof(record.Target48))
	case "caster-b", "caster-nil":
		rb.allow(unsafe.Pointer(&record.Caster16), unsafe.Sizeof(record.Caster16))
	case "spell":
		rb.allow(unsafe.Pointer(&record.Spell), unsafe.Sizeof(record.Spell))
	case "destination":
		rb.allow(unsafe.Pointer(&record.Pos2), unsafe.Sizeof(record.Pos2))
	case "position-target":
		ob[2].allow(unsafe.Pointer(&objects[2].PosVec), unsafe.Sizeof(objects[2].PosVec))
	case "invisible-target", "visible-target":
		ob[2].allow(unsafe.Pointer(&objects[2].Buffs), unsafe.Sizeof(objects[2].Buffs))
	case "netcode-target":
		ob[2].allow(unsafe.Pointer(&objects[2].NetCode), unsafe.Sizeof(objects[2].NetCode))
	case "update-caster":
		ob[1].allow(unsafe.Pointer(&objects[1].UpdateData), unsafe.Sizeof(objects[1].UpdateData))
	case "player-caster":
		ub[1].allow(unsafe.Pointer(&updates[1].Player), unsafe.Sizeof(updates[1].Player))
	case "", "frame":
	default:
		t.Fatal("unknown mutation")
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
		if p == nil {
			return -1
		}
		for i, u := range updates {
			if unsafe.Pointer(u) == p {
				return i
			}
		}
		t.Fatal("unknown update")
		return -2
	}
	playerIndex := func(p *Player) int {
		for i, v := range players {
			if v == p {
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
		var classes, flags, buffs, netcodes [4]uint32
		var objectUpdates, updatePlayers [4]int
		var indices [4]uint8
		for i, o := range objects {
			positions[i], classes[i], flags[i], buffs[i], netcodes[i] = words(o.PosVec), uint32(o.ObjClass), uint32(o.ObjFlags), o.Buffs, o.NetCode
			objectUpdates[i], updatePlayers[i], indices[i] = updateIndex(o.UpdateData), playerIndex(updates[i].Player), players[i].PlayerInd
		}
		return map[string]any{"source": index(record.Obj12), "caster": index(record.Caster16), "target": index(record.Target48), "spell": record.Spell, "level": record.Level, "flag": record.Flag20, "flags88": record.Flags88,
			"destination": words(record.Pos2), "timer": record.Frame68, "frame": frame, "positions": positions, "classes": classes, "flags": flags, "buffs": buffs, "netcodes": netcodes,
			"object_updates": objectUpdates, "update_players": updatePlayers, "player_indices": indices}
	}
	calls, mutated := map[string]int{}, false
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
		case "caster-nil":
			record.Caster16 = nil
		case "spell":
			record.Spell = 43
		case "destination":
			record.Pos2 = types.Ptf(-37.5, 43.25)
		case "position-target":
			objects[2].PosVec = types.Ptf(-17.5, 23.25)
		case "invisible-target":
			objects[2].Buffs |= 1
		case "visible-target":
			objects[2].Buffs &^= 1
		case "netcode-target":
			objects[2].NetCode = 0xdeadbeef
		case "update-caster":
			objects[1].UpdateData = unsafe.Pointer(updates[0])
		case "player-caster":
			updates[1].Player = players[0]
		case "frame":
			frame = 0xfffffffd
		default:
			t.Fatal("unknown mutation")
		}
	}
	events := make([]map[string]any, 0)
	event := func(kind string, args []any) {
		events = append(events, map[string]any{"kind": kind, "args": args, "state": state()})
		mutate(kind)
	}
	rt := SpellTeleportToTargetRuntime530A30{
		CoopMode: func() bool { return input.Coop }, Frame: func() uint32 { return frame }, TickRate: func() uint32 { return 30 },
		TileBlocksTeleport: func(p *types.Pointf) bool {
			if p != &record.Pos2 {
				t.Fatal("tile pointer")
			}
			w := words(*p)
			event("tile", []any{w[0], w[1]})
			return input.Blocked
		},
		TraceRay9: func(from, to types.Pointf) bool {
			f, d := words(from), words(to)
			event("ray", []any{f[0], f[1], d[0], d[1]})
			return input.Clear
		},
		SendPointFX: func(op netmsg.Op, p types.Pointf) { w := words(p); event("fx", []any{uint32(op), w[0], w[1]}) },
		CastSound: func(id spell.ID) sound.ID {
			event("sound", []any{uint32(id), 0})
			return sound.ID(input.Sound + uint32(calls["sound"]-1))
		},
		Audio: func(id sound.ID, target *Object, kind int, code uint32) {
			event("audio", []any{uint32(id), index(target), kind, code})
		},
		Teleport: func(target *Object, p types.Pointf) {
			w := words(p)
			event("teleport", []any{index(target), w[0], w[1]})
			target.PosVec = p
		},
		Attribution: func(caster, target *Object) { event("attribution", []any{index(caster), index(target)}) },
	}
	value := reflect.ValueOf(&rt).Elem()
	lookup := value.FieldByName("LookupUnseenTarget")
	if lookup.IsValid() {
		lookup.Set(reflect.ValueOf(func() string { event("lookup", []any{}); return "UnseenTarget" }))
	}
	sender := value.FieldByName("SendUnseenTarget")
	sender.Set(reflect.MakeFunc(sender.Type(), func(args []reflect.Value) []reflect.Value {
		if !lookup.IsValid() {
			event("lookup", []any{})
		} // Existing API performs lookup inside the already-bound target service.
		if len(args) == 2 && args[1].String() != "UnseenTarget" {
			t.Fatal("message token")
		}
		event("unseen", []any{index(args[0].Interface().(*Object))})
		return nil
	}))
	inform := value.FieldByName("InformNoLOS")
	inform.Set(reflect.MakeFunc(inform.Type(), func(args []reflect.Value) []reflect.Value {
		var data *PlayerUpdateData
		if unit, ok := args[0].Interface().(*Object); ok {
			if unit == nil || unit.UpdateData == nil {
				return nil
			}
			data = unit.UpdateDataPlayer() // Match the unchanged before production binding, including its nil guard.
		} else {
			data = args[0].Interface().(*PlayerUpdateData)
		}
		if data == nil || data.Player == nil {
			return nil
		}
		event("inform", []any{uint32(data.Player.PlayerInd), 0, 2})
		return nil
	}))
	wake := value.FieldByName("Wake")
	if wake.IsValid() {
		wake.Set(reflect.ValueOf(func(target *Object, from, to *types.Pointf) {
			if from != &target.PosVec || to != &record.Pos2 {
				t.Fatal("actual live wake pointers")
			}
			f, d := words(*from), words(*to)
			event("wake", []any{index(target), f[0], f[1], d[0], d[1]})
		}))
	}
	// Before the Wake service exists, the body saves this target/origin before
	// its first FX and inlines the helper. A nil allocation models the declared
	// no-output wake service without inventing a different target/origin.
	cachedTarget, cachedOrigin := obj(input.Target), types.Pointf{}
	if cachedTarget != nil {
		cachedOrigin = cachedTarget.PosVec
	}
	rt.NewObject = func(name string) *Object {
		if name != "TeleportWake" {
			t.Fatal("wake type")
		}
		f, d := words(cachedOrigin), words(record.Pos2)
		event("wake", []any{index(cachedTarget), f[0], f[1], d[0], d[1]})
		return nil
	}
	raw, err := hex.DecodeString(input.Delay)
	if err != nil || len(raw) != 8 {
		t.Fatal("binary64 balance")
	}
	delay := math.Float64frombits(binary.LittleEndian.Uint64(raw))
	delayField := value.FieldByName("TeleportDelay")
	delayField.Set(reflect.MakeFunc(delayField.Type(), func(args []reflect.Value) []reflect.Value {
		event("delay", []any{uint32(args[0].Uint())})
		return []reflect.Value{reflect.ValueOf(delay).Convert(delayField.Type().Out(0))}
	}))
	for _, block := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(block.pointer) <= 0xffffffff {
			t.Fatal("native frame below4GiB")
		}
		block.before = append([]byte(nil), unsafe.Slice((*byte)(block.pointer), block.size)...)
	}
	var result int32
	if input.Kind == "create" {
		result = SpellTeleportToTargetCreate530A30(record, rt)
	} else {
		result = SpellTeleportToTargetUpdate530B70(record, rt)
	}
	// The original producer proves mutation reachability. A before Go body
	// may omit that service entirely; compare its actual events/state instead
	// of aborting before serializing the semantic mismatch.
	for n, block := range blocks {
		after := unsafe.Slice((*byte)(block.pointer), block.size)
		for i, b := range block.before {
			if b != after[i] && !block.allowed[i] {
				t.Fatalf("native frame%d nonoutput/canary byte%d changed", n, i)
			}
		}
	}
	out, err := json.Marshal(map[string]any{"ret": result, "state": state(), "events": events})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
