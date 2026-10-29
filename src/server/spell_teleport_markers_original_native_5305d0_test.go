package server

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

type teleportMarkersInput5305D0 struct {
	Name          string       `json:"name"`
	Kind          string       `json:"kind"`
	Coop          bool         `json:"coop"`
	Spell         uint32       `json:"spell"`
	Level         uint32       `json:"level"`
	Timer         uint32       `json:"timer"`
	Frame         uint32       `json:"frame"`
	Source        int          `json:"source"`
	Caster        int          `json:"caster"`
	Target        int          `json:"target"`
	Flag          uint32       `json:"flag"`
	Flags88       uint32       `json:"flags88"`
	Classes       [8]uint32    `json:"classes"`
	Flags         [8]uint32    `json:"flags"`
	Buffs         [8]uint32    `json:"buffs"`
	Netcodes      [8]uint32    `json:"netcodes"`
	Field34       [8]uint32    `json:"field34"`
	Positions     [8][2]uint32 `json:"positions"`
	ObjectUpdates [8]int       `json:"object_updates"`
	Markers       [4][4]int    `json:"markers"`
	Charges       [4]uint32    `json:"charges"`
	Random        []int        `json:"random"`
	Delay         string       `json:"delay_bits64"`
	Sound         uint32       `json:"sound"`
	Mutation      struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}

func TestSpellTeleportMarkersOriginalNative5305D0(t *testing.T) {
	raw, err := os.ReadFile("testdata/spell_teleport_markers_native_original_5305d0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA     string `json:"original_sha256"`
		Control uint32 `json:"control_word"`
		Rows    []struct {
			Input  teleportMarkersInput5305D0 `json:"input"`
			Result json.RawMessage            `json:"result"`
		} `json:"rows"`
	}
	if json.Unmarshal(raw, &fixture) != nil || fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0xe3f || len(fixture.Rows) != 324 {
		t.Fatal("original marker duration identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			actual := teleportMarkersRun5305D0(t, row.Input)
			var got, want any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original native marker duration mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}

func teleportMarkersRun5305D0(t *testing.T, input teleportMarkersInput5305D0) json.RawMessage {
	t.Helper()
	var blocks []*tagOriginalBlock530160
	record, rb := tagOriginalAllocate530160[DurSpell](t, &blocks)
	var objects [8]*Object
	var updates [4]*PlayerUpdateData
	var ob [8]*tagOriginalBlock530160
	for i := range objects {
		objects[i], ob[i] = tagOriginalAllocate530160[Object](t, &blocks)
	}
	for i := range updates {
		var block *tagOriginalBlock530160
		updates[i], block = tagOriginalAllocate530160[PlayerUpdateData](t, &blocks)
		if strings.HasSuffix(input.Kind, "update") {
			block.allow(unsafe.Pointer(&updates[i].Field29), unsafe.Sizeof(updates[i].Field29))
			block.allow(unsafe.Pointer(&updates[i].Field39), unsafe.Sizeof(updates[i].Field39))
		}
	}
	obj := func(i int) *Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	for i, o := range objects {
		*o = Object{ObjClass: object.Class(input.Classes[i]), ObjFlags: object.Flags(input.Flags[i]), Buffs: input.Buffs[i], NetCode: input.Netcodes[i], Field34: input.Field34[i],
			PosVec: types.Ptf(math.Float32frombits(input.Positions[i][0]), math.Float32frombits(input.Positions[i][1]))}
		if input.ObjectUpdates[i] >= 0 {
			o.UpdateData = unsafe.Pointer(updates[input.ObjectUpdates[i]])
		}
		if strings.HasSuffix(input.Kind, "update") {
			ob[i].allow(unsafe.Pointer(&o.PosVec), unsafe.Sizeof(o.PosVec))
		}
		if input.Kind == "mark-update" {
			ob[i].allow(unsafe.Pointer(&o.Field34), unsafe.Sizeof(o.Field34))
		}
	}
	for i, data := range updates {
		data.Field39 = input.Charges[i]
		for j, n := range input.Markers[i] {
			data.Field29[j] = obj(n)
		}
	}
	*record = DurSpell{ID: 0xabcd, Spell: input.Spell, Level: input.Level, Obj12: obj(input.Source), Caster16: obj(input.Caster), Target48: obj(input.Target),
		Flag20: input.Flag, Flags88: input.Flags88, Frame68: input.Timer, Pos: types.Ptf(11.5, -22.25)}
	if strings.HasSuffix(input.Kind, "create") {
		rb.allow(unsafe.Pointer(&record.Frame68), unsafe.Sizeof(record.Frame68))
	}
	owner := input.Source
	if strings.HasPrefix(input.Kind, "pop") {
		owner = input.Caster
	}
	switch input.Mutation.Action {
	case "target-b":
		rb.allow(unsafe.Pointer(&record.Target48), unsafe.Sizeof(record.Target48))
	case "source-b":
		rb.allow(unsafe.Pointer(&record.Obj12), unsafe.Sizeof(record.Obj12))
	case "caster-b", "caster-nil":
		rb.allow(unsafe.Pointer(&record.Caster16), unsafe.Sizeof(record.Caster16))
	case "caster-destroyed":
		ob[1].allow(unsafe.Pointer(&objects[1].ObjFlags), unsafe.Sizeof(objects[1].ObjFlags))
	case "spell":
		rb.allow(unsafe.Pointer(&record.Spell), unsafe.Sizeof(record.Spell))
	case "position-target":
		ob[2].allow(unsafe.Pointer(&objects[2].PosVec), unsafe.Sizeof(objects[2].PosVec))
	case "position-marker":
		ob[4].allow(unsafe.Pointer(&objects[4].PosVec), unsafe.Sizeof(objects[4].PosVec))
	case "invisible-target", "visible-target":
		ob[2].allow(unsafe.Pointer(&objects[2].Buffs), unsafe.Sizeof(objects[2].Buffs))
	case "netcode-target":
		ob[2].allow(unsafe.Pointer(&objects[2].NetCode), unsafe.Sizeof(objects[2].NetCode))
	case "update-owner":
		ob[owner].allow(unsafe.Pointer(&objects[owner].UpdateData), unsafe.Sizeof(objects[owner].UpdateData))
	case "", "marker-b", "charge", "frame":
	default:
		t.Fatal("unknown mutation")
	}
	index := func(o *Object) int {
		if o == nil {
			return -1
		}
		for i, p := range objects {
			if o == p {
				return i
			}
		}
		t.Fatal("unknown object")
		return -2
	}
	words := func(p types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	frame, lastRandom := input.Frame, 0
	state := func() map[string]any {
		var positions [8][2]uint32
		var classes, flags, buffs, netcodes, field34 [8]uint32
		var objectUpdates [8]int
		var markers [4][4]int
		var charges [4]uint32
		for i, o := range objects {
			positions[i], classes[i], flags[i], buffs[i], netcodes[i], field34[i] = words(o.PosVec), uint32(o.ObjClass), uint32(o.ObjFlags), o.Buffs, o.NetCode, o.Field34
			objectUpdates[i] = -1
			for j, d := range updates {
				if unsafe.Pointer(d) == o.UpdateData {
					objectUpdates[i] = j
				}
			}
		}
		for i, d := range updates {
			charges[i] = d.Field39
			for j, o := range d.Field29 {
				markers[i][j] = index(o)
			}
		}
		return map[string]any{"source": index(record.Obj12), "caster": index(record.Caster16), "target": index(record.Target48), "spell": record.Spell, "level": record.Level, "flag": record.Flag20, "flags88": record.Flags88, "timer": record.Frame68, "frame": frame,
			"positions": positions, "classes": classes, "flags": flags, "buffs": buffs, "netcodes": netcodes, "field34": field34, "object_updates": objectUpdates, "markers": markers, "charges": charges}
	}
	ownerData := func() *PlayerUpdateData { return updates[input.ObjectUpdates[owner]] }
	selected := func() int {
		if strings.HasPrefix(input.Kind, "mark") {
			return int(input.Spell) - 122
		}
		return lastRandom
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
		case "source-b":
			record.Obj12 = objects[3]
		case "caster-b":
			record.Caster16 = objects[3]
		case "caster-nil":
			record.Caster16 = nil
		case "caster-destroyed":
			objects[1].ObjFlags |= object.Flags(0x20)
		case "spell":
			record.Spell = 43
		case "position-target":
			objects[2].PosVec = types.Ptf(-17.5, 23.25)
		case "position-marker":
			objects[4].PosVec = types.Ptf(-37.5, 43.25)
		case "invisible-target":
			objects[2].Buffs |= 1
		case "visible-target":
			objects[2].Buffs &^= 1
		case "netcode-target":
			objects[2].NetCode = 0xdeadbeef
		case "marker-b":
			ownerData().Field29[selected()] = objects[7]
		case "charge":
			ownerData().Field39 = 0xfedcba01
		case "update-owner":
			objects[owner].UpdateData = unsafe.Pointer(updates[3])
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
	var runtime any
	if strings.HasPrefix(input.Kind, "pop") {
		runtime = &SpellTeleportPopRuntime530820{}
	} else {
		runtime = &SpellTeleportToMarkRuntime5305D0{}
	}
	v := reflect.ValueOf(runtime).Elem()
	set := func(name string, callback any) { v.FieldByName(name).Set(reflect.ValueOf(callback)) }
	set("CoopMode", func() bool { return input.Coop })
	set("Frame", func() uint32 { return frame })
	set("TickRate", func() uint32 { return 30 })
	if random := v.FieldByName("RandomIndex"); random.IsValid() {
		set("RandomIndex", func() int {
			n := calls["random"]
			if n >= len(input.Random) {
				n = len(input.Random) - 1
			}
			lastRandom = input.Random[n]
			event("random", []any{lastRandom})
			return lastRandom
		})
	}
	set("SendPointFX", func(op netmsg.Op, p types.Pointf) { w := words(p); event("fx", []any{uint32(op), w[0], w[1]}) })
	soundField, mode := "OnSound", uint32(1)
	if strings.HasPrefix(input.Kind, "pop") {
		soundField, mode = "CastSound", 0
	}
	set(soundField, func(id spell.ID) sound.ID {
		event("sound", []any{uint32(id), mode})
		return sound.ID(input.Sound + uint32(calls["sound"]-1))
	})
	set("Audio", func(id sound.ID, target *Object, kind int, code uint32) {
		event("audio", []any{uint32(id), index(target), kind, code})
	})
	set("DelayedDelete", func(marker *Object) { event("delete", []any{index(marker)}) })
	set("Attribution", func(caster, target *Object) { event("attribution", []any{index(caster), index(target)}) })
	pointIndex := func(p *types.Pointf) int {
		for i, o := range objects {
			if p == &o.PosVec {
				return i
			}
		}
		t.Fatal("point is not the actual object position")
		return -2
	}
	teleport := v.FieldByName("Teleport")
	teleport.Set(reflect.MakeFunc(teleport.Type(), func(args []reflect.Value) []reflect.Value {
		target := args[0].Interface().(*Object)
		var p types.Pointf
		marker := -1
		if point, ok := args[1].Interface().(*types.Pointf); ok {
			p, marker = *point, pointIndex(point)
		} else {
			p = args[1].Interface().(types.Pointf)
			// Before's value-only API has no point identity. These controlled
			// marker positions are unique, so record their unambiguous source.
			for i := 4; i < len(objects); i++ {
				if objects[i].PosVec == p {
					if marker != -1 {
						t.Fatal("ambiguous old value point")
					}
					marker = i
				}
			}
			if marker == -1 {
				t.Fatal("old value point has no marker")
			}
		}
		w := words(p)
		event("teleport", []any{index(target), marker, w[0], w[1]})
		target.PosVec = p
		return nil
	}))
	if wake := v.FieldByName("Wake"); wake.IsValid() {
		set("Wake", func(target *Object, from, to *types.Pointf) {
			if from != &target.PosVec {
				t.Fatal("wake origin pointer")
			}
			f, d := words(*from), words(*to)
			event("wake", []any{index(target), pointIndex(to), f[0], f[1], d[0], d[1]})
		})
	}
	// The unchanged before body caches its target at entry and, for Pop,
	// its origin after random selection. A failed allocation gives the same
	// declared no-output wake service without executing transitive callees.
	set("NewObject", func(name string) *Object {
		if name != "TeleportWake" {
			t.Fatal("wake type")
		}
		target := obj(input.Target)
		marker := ownerData().Field29[selected()]
		f, d := words(target.PosVec), words(marker.PosVec)
		event("wake", []any{index(target), index(marker), f[0], f[1], d[0], d[1]})
		return nil
	})
	raw, err := hex.DecodeString(input.Delay)
	if err != nil || len(raw) != 8 {
		t.Fatal("binary64 balance")
	}
	delay := math.Float64frombits(binary.LittleEndian.Uint64(raw))
	delayField := v.FieldByName("TeleportDelay")
	delayField.Set(reflect.MakeFunc(delayField.Type(), func(args []reflect.Value) []reflect.Value {
		event("delay", []any{uint32(args[0].Uint())})
		return []reflect.Value{reflect.ValueOf(delay).Convert(delayField.Type().Out(0))}
	}))
	for _, b := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(b.pointer) <= 0xffffffff {
			t.Fatal("native frame below4GiB")
		}
		b.before = append([]byte(nil), unsafe.Slice((*byte)(b.pointer), b.size)...)
	}
	var result int32
	switch input.Kind {
	case "pop-create":
		result = SpellTeleportPopCreate530820(record, *runtime.(*SpellTeleportPopRuntime530820))
	case "pop-update":
		result = SpellTeleportPopUpdate530880(record, *runtime.(*SpellTeleportPopRuntime530820))
	case "mark-create":
		result = SpellTeleportToMarkCreate5305D0(record, *runtime.(*SpellTeleportToMarkRuntime5305D0))
	case "mark-update":
		result = SpellTeleportToMarkUpdate530650(record, *runtime.(*SpellTeleportToMarkRuntime5305D0))
	default:
		t.Fatal("unknown entry")
	}
	// Missing before services serialize as actual mismatches. Reachability is
	// independently asserted by the unchanged original producer.
	for n, b := range blocks {
		for i, value := range b.before {
			if value != unsafe.Slice((*byte)(b.pointer), b.size)[i] && !b.allowed[i] {
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
