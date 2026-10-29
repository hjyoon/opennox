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
	"github.com/opennox/libs/types"
)

type turnUndeadOriginalInput531310 struct {
	Name        string        `json:"name"`
	Kind        string        `json:"kind"`
	Level       uint32        `json:"level"`
	Caster      int           `json:"caster"`
	Anchor      int           `json:"anchor"`
	Flag        uint32        `json:"flag"`
	Budget      uint32        `json:"budget"`
	Caches      [2]uint32     `json:"caches"`
	Lookup      uint32        `json:"lookup"`
	Failures    []int         `json:"failures"`
	Head        int           `json:"head"`
	Positions   [46][2]uint32 `json:"positions"`
	Types       [46]uint16    `json:"types"`
	Spells      [46]int       `json:"spells"`
	Links       [46]int       `json:"links"`
	BalanceBits string        `json:"balance_bits64"`
	Mutation    struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}

func TestSpellTurnUndeadOriginalNative531310(t *testing.T) {
	raw, err := os.ReadFile("testdata/spell_turn_undead_native_original_531310.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA     string `json:"original_sha256"`
		Control uint32 `json:"control_word"`
		Rows    []struct {
			Input  turnUndeadOriginalInput531310 `json:"input"`
			Result json.RawMessage               `json:"result"`
		} `json:"rows"`
	}
	if json.Unmarshal(raw, &fixture) != nil || fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0xe3f || len(fixture.Rows) != 125 {
		t.Fatal("original TurnUndead identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			actual := turnUndeadOriginalRun531310(t, row.Input)
			var got, want any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original native TurnUndead mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}

func turnUndeadOriginalRun531310(t *testing.T, in turnUndeadOriginalInput531310) json.RawMessage {
	t.Helper()
	var blocks []*tagOriginalBlock530160
	record, rb := tagOriginalAllocate530160[DurSpell](t, &blocks)
	other, _ := tagOriginalAllocate530160[DurSpell](t, &blocks)
	var objects [46]*Object
	var ob [46]*tagOriginalBlock530160
	var collisions [46]*UndeadKillerCollideData
	var cb [46]*tagOriginalBlock530160
	for i := range objects {
		objects[i], ob[i] = tagOriginalAllocate530160[Object](t, &blocks)
		collisions[i], cb[i] = tagOriginalAllocate530160[UndeadKillerCollideData](t, &blocks)
	}
	caches, kb := tagOriginalAllocate530160[[2]uint32](t, &blocks)
	*caches = in.Caches
	kb.allow(unsafe.Pointer(caches), unsafe.Sizeof(*caches))
	obj := func(i int) *Object {
		if i < 0 {
			return nil
		}
		return objects[i]
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
		return -99
	}
	spellIndex := func(p *DurSpell) int {
		if p == nil {
			return -1
		}
		if p == record {
			return 0
		}
		if p == other {
			return 1
		}
		t.Fatal("unknown duration")
		return -99
	}
	for i, o := range objects {
		o.TypeInd = in.Types[i]
		o.PosVec = types.Ptf(math.Float32frombits(in.Positions[i][0]), math.Float32frombits(in.Positions[i][1]))
		o.Direction1 = Dir16(0xa100 + i)
		o.Direction2 = Dir16(0xb200 + i)
		o.VelVec = types.Ptf(1.25, -2.5)
		o.Float28 = 7
		o.ObjNext = obj(in.Links[i])
		o.CollideData = unsafe.Pointer(collisions[i])
		if in.Spells[i] == 0 {
			collisions[i].Spell = record
		} else if in.Spells[i] == 1 {
			collisions[i].Spell = other
		}
	}
	*record = DurSpell{ID: 0xabcd, Spell: 0x66, Level: in.Level, Caster16: obj(in.Caster), Obj24: obj(in.Anchor), Flag20: in.Flag, Field72: int32(in.Budget), Pos: types.Ptf(11.5, -22.25), Frame68: 0x87654321, Flags88: 0x77556601}
	head, current := in.Head, 3
	lookups, attempts := 0, 0
	coreState := func() map[string]any {
		var positions [3][2]uint32
		for i := range positions {
			positions[i] = [2]uint32{math.Float32bits(objects[i].PosVec.X), math.Float32bits(objects[i].PosVec.Y)}
		}
		return map[string]any{"level": record.Level, "caster": index(record.Caster16), "anchor": index(record.Obj24), "flag": record.Flag20, "budget": uint32(record.Field72), "caches": *caches, "source_positions": positions, "lookups": lookups, "attempts": attempts}
	}
	state := func() map[string]any {
		s := coreState()
		rows := make([][]int64, 46)
		for i, o := range objects {
			rows[i] = []int64{int64(o.TypeInd), int64(spellIndex(collisions[i].Spell)), int64(o.Direction1), int64(o.Direction2), int64(math.Float32bits(o.VelVec.X)), int64(math.Float32bits(o.VelVec.Y)), int64(math.Float32bits(o.Float28)), int64(math.Float32bits(o.PosVec.X)), int64(math.Float32bits(o.PosVec.Y)), int64(index(o.ObjNext)), int64(index(o.ObjOwner))}
		}
		s["head"], s["objects"] = head, rows
		return s
	}
	if in.Kind == "create" {
		rb.allow(unsafe.Pointer(&record.Field72), unsafe.Sizeof(record.Field72))
		for i := 3; i < 46; i++ {
			o := objects[i]
			ob[i].allow(unsafe.Pointer(&o.TypeInd), unsafe.Sizeof(o.TypeInd))
			ob[i].allow(unsafe.Pointer(&o.PosVec), unsafe.Sizeof(o.PosVec))
			ob[i].allow(unsafe.Pointer(&o.ObjOwner), unsafe.Sizeof(o.ObjOwner))
			ob[i].allow(unsafe.Pointer(&o.Direction1), unsafe.Sizeof(o.Direction1))
			ob[i].allow(unsafe.Pointer(&o.Direction2), unsafe.Sizeof(o.Direction2))
			ob[i].allow(unsafe.Pointer(&o.VelVec), unsafe.Sizeof(o.VelVec))
			ob[i].allow(unsafe.Pointer(&o.Float28), unsafe.Sizeof(o.Float28))
			cb[i].allow(unsafe.Pointer(&collisions[i].Spell), unsafe.Sizeof(collisions[i].Spell))
		}
	}
	switch in.Mutation.Action {
	case "caster-b":
		rb.allow(unsafe.Pointer(&record.Caster16), unsafe.Sizeof(record.Caster16))
	case "anchor-b":
		rb.allow(unsafe.Pointer(&record.Obj24), unsafe.Sizeof(record.Obj24))
	case "flag":
		rb.allow(unsafe.Pointer(&record.Flag20), unsafe.Sizeof(record.Flag20))
	case "source-pos":
		ob[0].allow(unsafe.Pointer(&objects[0].PosVec), unsafe.Sizeof(objects[0].PosVec))
	case "budget":
		rb.allow(unsafe.Pointer(&record.Field72), unsafe.Sizeof(record.Field72))
	case "level":
		rb.allow(unsafe.Pointer(&record.Level), unsafe.Sizeof(record.Level))
	case "next-nil", "next-skip":
		for i := range objects {
			ob[i].allow(unsafe.Pointer(&objects[i].ObjNext), unsafe.Sizeof(objects[i].ObjNext))
		}
	case "next-type":
		ob[6].allow(unsafe.Pointer(&objects[6].TypeInd), unsafe.Sizeof(objects[6].TypeInd))
	case "next-spell":
		cb[6].allow(unsafe.Pointer(&collisions[6].Spell), unsafe.Sizeof(collisions[6].Spell))
	case "", "create-cache", "destroy-cache", "head-b", "projectile-fields", "projectile-spell":
	default:
		t.Fatal("unknown mutation", in.Mutation.Action)
	}
	changed := false
	calls := map[string]int{}
	mutate := func(stage string) {
		at := calls[stage]
		calls[stage]++
		if stage == "lookup" {
			lookups++
		}
		if stage == "new" {
			attempts++
		}
		if changed || stage != in.Mutation.Stage || at != in.Mutation.At {
			return
		}
		changed = true
		switch in.Mutation.Action {
		case "caster-b":
			record.Caster16 = objects[1]
		case "anchor-b":
			record.Obj24 = objects[2]
		case "flag":
			record.Flag20 ^= 1
		case "source-pos":
			objects[0].PosVec = types.Ptf(777.25, -888.5)
		case "budget":
			record.Field72 = 0x13579bdf
		case "level":
			record.Level = 0xfffffffe
		case "create-cache":
			caches[0] = 99
		case "destroy-cache":
			caches[1] = 20
		case "head-b":
			head = 6
		case "next-nil":
			objects[current].ObjNext = nil
		case "next-skip":
			objects[current].ObjNext = objects[8]
		case "next-type":
			objects[6].TypeInd = 20
		case "next-spell":
			collisions[6].Spell = other
		case "projectile-fields":
			objects[current].Direction1 = 0x1234
			objects[current].Direction2 = 0x5678
			objects[current].VelVec = types.Ptf(1.25, 2.5)
			objects[current].Float28 = 9
		case "projectile-spell":
			collisions[current].Spell = other
		default:
			t.Fatal("unknown callback mutation")
		}
	}
	events := make([]map[string]any, 0)
	event := func(kind string, args []any) {
		events = append(events, map[string]any{"kind": kind, "args": args, "state": coreState()})
		mutate(kind)
	}
	resolve := func(name string) uint32 {
		if name != "UndeadKiller" {
			t.Fatal("type name", name)
		}
		mutate("lookup")
		return in.Lookup
	}
	allocate := func(kind uint32) *Object {
		attempt := attempts
		current = 3 + attempt
		ok := kind != 0 && kind <= 0xffff
		for _, failed := range in.Failures {
			if failed == attempt {
				ok = false
			}
		}
		out := obj(current)
		if !ok {
			out = nil
		}
		event("new", []any{kind, index(out)})
		if out != nil {
			out.TypeInd = uint16(kind)
		}
		return out
	}
	bits, err := hex.DecodeString(in.BalanceBits)
	if err != nil || len(bits) != 8 {
		t.Fatal("binary64 balance", err)
	}
	balance := func(level uint32) float64 {
		event("balance", []any{level})
		return math.Float64frombits(binary.LittleEndian.Uint64(bits))
	}
	rt := SpellTurnUndeadRuntime531310{
		TypeInd: func(name string) int { return int(resolve(name)) },
		CreateAt: func(o, owner *Object, p types.Pointf) {
			current = index(o)
			event("create", []any{current, index(owner), math.Float32bits(p.X), math.Float32bits(p.Y), spellIndex((*UndeadKillerCollideData)(o.CollideData).Spell)})
			o.PosVec = p
			o.ObjOwner = owner
		},
		SendPointFX: func(effect netmsg.Op, p types.Pointf) {
			event("fx", []any{effect, math.Float32bits(p.X), math.Float32bits(p.Y)})
		},
		FirstObject:   func() *Object { return obj(head) },
		DelayedDelete: func(o *Object) { current = index(o); event("delete", []any{current}) },
	}
	// The old production name allocator resolves its name on every attempt;
	// the new typed allocator receives the caller's actual cache dword.
	switch field := any(&rt.NewObject).(type) {
	case *func(string) *Object:
		*field = func(name string) *Object { return allocate(resolve(name)) }
	case *func(uint32) *Object:
		*field = allocate
	default:
		t.Fatal("unexpected allocator API")
	}
	switch field := any(&rt.KillPoints).(type) {
	case *func(uint32) float32:
		*field = func(level uint32) float32 { return float32(balance(level)) }
	case *func(uint32) float64:
		*field = balance
	default:
		t.Fatal("unexpected balance API")
	}
	value := reflect.ValueOf(&rt).Elem()
	for name, pointer := range map[string]*uint32{"CreateType": &caches[0], "DestroyType": &caches[1]} {
		if field := value.FieldByName(name); field.IsValid() {
			field.Set(reflect.ValueOf(pointer))
		}
	}
	for _, b := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(b.pointer) <= 0xffffffff {
			t.Fatal("native C-owned TurnUndead frame must be above4GiB")
		}
		b.before = append([]byte(nil), unsafe.Slice((*byte)(b.pointer), b.size)...)
	}
	var result any
	if in.Kind == "create" {
		result = SpellTurnUndeadCreate531310(record, rt)
	} else if in.Kind == "destroy" {
		SpellTurnUndeadDestroy531420(record, rt)
	} else {
		t.Fatal("root", in.Kind)
	}
	if in.Mutation.Stage != "" && !changed {
		t.Fatal("mutation not reached")
	}
	for _, b := range blocks {
		after := unsafe.Slice((*byte)(b.pointer), b.size)
		for i, v := range after {
			if v != b.before[i] && !b.allowed[i] {
				t.Fatalf("nonoutput TurnUndead canary: block=%p offset=%d before=%02x after=%02x", b.pointer, i, b.before[i], v)
			}
		}
	}
	encoded, err := json.Marshal(map[string]any{"ret": result, "state": state(), "events": events})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
