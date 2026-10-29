package server

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type moonglowOriginalInput531A00 struct {
	Name          string       `json:"name"`
	Kind          string       `json:"kind"`
	Target        int          `json:"target"`
	Level         uint32       `json:"level"`
	Visual        int          `json:"visual"`
	Allocated     int          `json:"allocated"`
	Classes       [6]uint32    `json:"classes"`
	Flags         [6]uint32    `json:"flags"`
	Positions     [6][2]uint32 `json:"positions"`
	ObjectUpdates [6]int       `json:"object_updates"`
	Players       [3]int       `json:"players"`
	PlayerFlags   [3]uint32    `json:"player_flags"`
	Cursors       [3][2]int32  `json:"cursors"`
	BalanceBits   string       `json:"balance_bits64"`
	Mutation      struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}

func TestSpellMoonglowOriginalNative531A00(t *testing.T) {
	raw, err := os.ReadFile("testdata/spell_moonglow_native_original_531a00.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA     string `json:"original_sha256"`
		Control uint32 `json:"control_word"`
		Rows    []struct {
			Input  moonglowOriginalInput531A00 `json:"input"`
			Result json.RawMessage             `json:"result"`
		} `json:"rows"`
	}
	if json.Unmarshal(raw, &fixture) != nil || fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0xe3f || len(fixture.Rows) != 142 {
		t.Fatal("original Moonglow identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			actual := moonglowOriginalRun531A00(t, row.Input)
			var got, want any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original native Moonglow mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}

func moonglowOriginalRun531A00(t *testing.T, input moonglowOriginalInput531A00) json.RawMessage {
	t.Helper()
	var blocks []*tagOriginalBlock530160
	record, rb := tagOriginalAllocate530160[DurSpell](t, &blocks)
	var objects [6]*Object
	var ob [6]*tagOriginalBlock530160
	for i := range objects {
		objects[i], ob[i] = tagOriginalAllocate530160[Object](t, &blocks)
	}
	var updates [3]*PlayerUpdateData
	var ub [3]*tagOriginalBlock530160
	var players [3]*Player
	var pb [3]*tagOriginalBlock530160
	for i := range updates {
		updates[i], ub[i] = tagOriginalAllocate530160[PlayerUpdateData](t, &blocks)
		players[i], pb[i] = tagOriginalAllocate530160[Player](t, &blocks)
	}
	obj := func(i int) *Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	for i, o := range objects {
		o.ObjClass, o.ObjFlags = object.Class(input.Classes[i]), object.Flags(input.Flags[i])
		o.PosVec = types.Ptf(math.Float32frombits(input.Positions[i][0]), math.Float32frombits(input.Positions[i][1]))
		if input.ObjectUpdates[i] >= 0 {
			o.UpdateData = unsafe.Pointer(updates[input.ObjectUpdates[i]])
		}
	}
	for i := range updates {
		updates[i].Player = players[input.Players[i]]
		players[i].Field3680 = input.PlayerFlags[i]
		players[i].CursorVec = image.Pt(int(input.Cursors[i][0]), int(input.Cursors[i][1]))
	}
	*record = DurSpell{ID: 0xabcd, Spell: 0x66, Level: input.Level, Target48: obj(input.Target), Field72: 0x55aa, Pos: types.Ptf(11.5, -22.25), Frame68: 0x87654321, Flags88: 0x77556601}
	stored := obj(input.Visual)
	index := func(p *Object) int {
		if p == nil {
			return -1
		}
		for i, o := range objects {
			if o == p {
				return i
			}
		}
		t.Fatal("unknown object identity")
		return -2
	}
	udIndex := func(p unsafe.Pointer) int {
		if p == nil {
			return -1
		}
		for i, o := range updates {
			if unsafe.Pointer(o) == p {
				return i
			}
		}
		t.Fatal("unknown update identity")
		return -2
	}
	plIndex := func(p *Player) int {
		if p == nil {
			return -1
		}
		for i, o := range players {
			if o == p {
				return i
			}
		}
		t.Fatal("unknown player identity")
		return -2
	}
	state := func() map[string]any {
		var classes, flags [6]uint32
		var positions [6][2]uint32
		var objectUpdates [6]int
		var playerRefs [3]int
		var playerFlags [3]uint32
		var cursors [3][2]int32
		for i, o := range objects {
			classes[i], flags[i] = uint32(o.ObjClass), uint32(o.ObjFlags)
			positions[i] = [2]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y)}
			objectUpdates[i] = udIndex(o.UpdateData)
		}
		for i, o := range updates {
			playerRefs[i] = plIndex(o.Player)
			playerFlags[i] = players[i].Field3680
			cursors[i] = [2]int32{int32(players[i].CursorVec.X), int32(players[i].CursorVec.Y)}
		}
		return map[string]any{"target": index(record.Target48), "level": record.Level, "visual": index(stored), "classes": classes, "flags": flags, "positions": positions, "object_updates": objectUpdates, "players": playerRefs, "player_flags": playerFlags, "cursors": cursors}
	}
	if input.Kind == "create" && input.Allocated >= 0 {
		ob[input.Allocated].allow(unsafe.Pointer(&objects[input.Allocated].PosVec), unsafe.Sizeof(objects[input.Allocated].PosVec))
	}
	switch input.Mutation.Action {
	case "target-b", "target-monster", "target-nil":
		rb.allow(unsafe.Pointer(&record.Target48), unsafe.Sizeof(record.Target48))
	case "level":
		rb.allow(unsafe.Pointer(&record.Level), unsafe.Sizeof(record.Level))
	case "target-destroyed", "target-dead":
		ob[0].allow(unsafe.Pointer(&objects[0].ObjFlags), unsafe.Sizeof(objects[0].ObjFlags))
	case "update-b":
		ob[0].allow(unsafe.Pointer(&objects[0].UpdateData), unsafe.Sizeof(objects[0].UpdateData))
	case "player-b":
		ub[0].allow(unsafe.Pointer(&updates[0].Player), unsafe.Sizeof(updates[0].Player))
	case "cursor-mode":
		pb[0].allow(unsafe.Pointer(&players[0].Field3680), unsafe.Sizeof(players[0].Field3680))
	case "cursor":
		pb[0].allow(unsafe.Pointer(&players[0].CursorVec), unsafe.Sizeof(players[0].CursorVec))
	case "", "visual-b", "visual-nil":
	default:
		t.Fatal("unknown mutation", input.Mutation.Action)
	}
	changed := false
	calls := make(map[string]int)
	mutate := func(kind string) {
		at := calls[kind]
		calls[kind]++
		m := input.Mutation
		if changed || m.Stage != kind || m.At != at {
			return
		}
		changed = true
		switch m.Action {
		case "target-b":
			record.Target48 = objects[1]
		case "target-monster":
			record.Target48 = objects[2]
		case "target-nil":
			record.Target48 = nil
		case "level":
			record.Level = 0xfffffffe
		case "visual-b":
			stored = objects[5]
		case "visual-nil":
			stored = nil
		case "target-destroyed":
			objects[0].ObjFlags |= object.Flags(0x20)
		case "target-dead":
			objects[0].ObjFlags |= object.Flags(0x8000)
		case "update-b":
			objects[0].UpdateData = unsafe.Pointer(updates[1])
		case "player-b":
			updates[0].Player = players[1]
		case "cursor-mode":
			players[0].Field3680 ^= 0x10
		case "cursor":
			players[0].CursorVec = image.Pt(16777219, -16777219)
		default:
			t.Fatal("unknown callback mutation", m.Action)
		}
	}
	events := make([]map[string]any, 0)
	event := func(kind string, args []any) {
		events = append(events, map[string]any{"kind": kind, "args": args, "state": state()})
		mutate(kind)
	}
	bits, err := hex.DecodeString(input.BalanceBits)
	if err != nil || len(bits) != 8 {
		t.Fatal("invalid binary64 balance", err)
	}
	balance := func() float64 {
		event("balance", []any{})
		return math.Float64frombits(binary.LittleEndian.Uint64(bits))
	}
	runtime := SpellMoonglowRuntime531A00{
		NewObject: func(name string) *Object { event("new", []any{name}); return obj(input.Allocated) },
		CreateAt: func(visual, owner *Object, position types.Pointf) {
			event("create", []any{index(visual), index(owner), math.Float32bits(position.X), math.Float32bits(position.Y)})
			visual.PosVec = position
		},
		ApplyBuff: func(target *Object, buff EnchantID, duration int16, power int8) {
			event("apply", []any{index(target), buff, duration, power})
		},
		BuffOff:       func(target *Object, buff EnchantID) { event("off", []any{index(target), buff}) },
		DelayedDelete: func(visual *Object) { event("delete", []any{index(visual)}) },
		LoadVisual:    func(*DurSpell) *Object { return stored },
		StoreVisual:   func(_ *DurSpell, visual *Object) { stored = visual },
	}
	// Both API shapes keep this comparator byte-identical across the port.
	// The old float32 binding performs its actual nearest binary64 conversion;
	// the new binary64 binding lets the original directed spill happen in core.
	switch field := any(&runtime.EnchantmentDuration).(type) {
	case *func() float32:
		*field = func() float32 { return float32(balance()) }
	case *func() float64:
		*field = balance
	default:
		t.Fatal("unexpected Moonglow balance API")
	}
	for _, b := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(b.pointer) <= 0xffffffff {
			t.Fatal("native C-owned Moonglow block must be above4GiB")
		}
		b.before = append([]byte(nil), unsafe.Slice((*byte)(b.pointer), b.size)...)
	}
	var result any
	switch input.Kind {
	case "create":
		result = SpellMoonglowCreate531A00(record, runtime)
	case "destroy":
		SpellMoonglowDestroy531AF0(record, runtime)
	default:
		t.Fatal("unknown root", input.Kind)
	}
	if input.Mutation.Stage != "" && !changed {
		t.Fatal("callback mutation not reached")
	}
	for _, b := range blocks {
		after := unsafe.Slice((*byte)(b.pointer), b.size)
		for i, v := range after {
			if v != b.before[i] && !b.allowed[i] {
				t.Fatalf("nonoutput Moonglow canary changed: block=%p offset=%d before=%02x after=%02x", b.pointer, i, b.before[i], v)
			}
		}
	}
	encoded, err := json.Marshal(map[string]any{"ret": result, "state": state(), "events": events})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
