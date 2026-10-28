package server

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type tagOriginalInput530160 struct {
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Caster      int       `json:"caster"`
	Target      int       `json:"target"`
	Level       uint32    `json:"level"`
	Timer       uint32    `json:"timer"`
	Frame       uint32    `json:"frame"`
	Classes     [4]uint32 `json:"classes"`
	Flags       [4]uint32 `json:"flags"`
	Types       [2]uint16 `json:"types"`
	Codes       [2]uint16 `json:"codes"`
	Indexes     [2]byte   `json:"indexes"`
	BalanceBits string    `json:"balance_bits64"`
	Mutation    string    `json:"mutation"`
}

type tagOriginalBlock530160 struct {
	pointer unsafe.Pointer
	size    int
	allowed []bool
	before  []byte
}

func tagOriginalAllocate530160[T comparable](t *testing.T, blocks *[]*tagOriginalBlock530160) (*T, *tagOriginalBlock530160) {
	t.Helper()
	type frame struct {
		Before [32]byte
		Value  T
		After  [32]byte
	}
	p, free := alloc.New(frame{})
	t.Cleanup(free)
	*p = frame{}
	for i := range p.Before {
		p.Before[i], p.After[i] = byte(0x55+i), byte(0xf1-i)
	}
	b := &tagOriginalBlock530160{pointer: unsafe.Pointer(p), size: int(unsafe.Sizeof(*p))}
	b.allowed = make([]bool, b.size)
	*blocks = append(*blocks, b)
	return &p.Value, b
}
func (b *tagOriginalBlock530160) allow(p unsafe.Pointer, n uintptr) {
	off := int(uintptr(p) - uintptr(b.pointer))
	for i := off; i < off+int(n); i++ {
		b.allowed[i] = true
	}
}

func TestSpellTagOriginalNative530160(t *testing.T) {
	raw, err := os.ReadFile("testdata/spell_tag_native_original_530160.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Control     uint32 `json:"control_word"`
		Rows        []struct {
			Input  tagOriginalInput530160 `json:"input"`
			Result json.RawMessage        `json:"result"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0xe3f || len(fixture.Rows) != 76 {
		t.Fatal("original Tag fixture identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			actual := tagOriginalRunNative530160(t, row.Input)
			var got, want any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(row.Result, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("original native Tag mismatch\ngot %s\nwant %s", actual, row.Result)
			}
		})
	}
}

func tagOriginalRunNative530160(t *testing.T, input tagOriginalInput530160) json.RawMessage {
	t.Helper()
	var blocks []*tagOriginalBlock530160
	record, rb := tagOriginalAllocate530160[DurSpell](t, &blocks)
	var objects [4]*Object
	var ob [4]*tagOriginalBlock530160
	for i := range objects {
		objects[i], ob[i] = tagOriginalAllocate530160[Object](t, &blocks)
		objects[i].ObjClass = object.Class(input.Classes[i])
		objects[i].ObjFlags = object.Flags(input.Flags[i])
	}
	obj := func(i int) *Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	var data [2]*PlayerUpdateData
	var db [2]*tagOriginalBlock530160
	var infos [2]*Player
	var ib [2]*tagOriginalBlock530160
	for i := range data {
		data[i], db[i] = tagOriginalAllocate530160[PlayerUpdateData](t, &blocks)
		infos[i], ib[i] = tagOriginalAllocate530160[Player](t, &blocks)
		objects[i].UpdateData = unsafe.Pointer(data[i])
		data[i].Player = infos[i]
		infos[i].PlayerInd = input.Indexes[i]
		objects[i+2].TypeInd = input.Types[i]
	}
	*record = DurSpell{ID: 0xabcd, Spell: 116, Level: input.Level, Caster16: obj(input.Caster), Target48: obj(input.Target), Frame68: input.Timer, Field36: 0x12345678, Field40: 0x87654321, Flags88: 0xa1b2c3d4}
	switch input.Kind {
	case "create":
		rb.allow(unsafe.Pointer(&record.Frame68), unsafe.Sizeof(record.Frame68))
	case "update", "destroy":
	default:
		t.Fatal("unknown root", input.Kind)
	}
	switch input.Mutation {
	case "balance-caster-b", "mark-caster-b":
		rb.allow(unsafe.Pointer(&record.Caster16), unsafe.Sizeof(record.Caster16))
	case "balance-data-b", "mark-data-b", "code-data-b", "unmark-data-b":
		ob[0].allow(unsafe.Pointer(&objects[0].UpdateData), unsafe.Sizeof(objects[0].UpdateData))
	case "balance-info-b", "mark-info-b", "unmark-info-b":
		db[0].allow(unsafe.Pointer(&data[0].Player), unsafe.Sizeof(data[0].Player))
	case "balance-target-b", "mark-target-b", "code-target-b", "unmark-target-b":
		rb.allow(unsafe.Pointer(&record.Target48), unsafe.Sizeof(record.Target48))
	case "balance-level":
		rb.allow(unsafe.Pointer(&record.Level), unsafe.Sizeof(record.Level))
	case "mark-index", "unmark-index":
		ib[0].allow(unsafe.Pointer(&infos[0].PlayerInd), unsafe.Sizeof(infos[0].PlayerInd))
	case "code-type-b":
		ob[2].allow(unsafe.Pointer(&objects[2].TypeInd), unsafe.Sizeof(objects[2].TypeInd))
	case "send-timer":
		rb.allow(unsafe.Pointer(&record.Frame68), unsafe.Sizeof(record.Frame68))
	case "none", "balance-frame":
	default:
		t.Fatal("unknown mutation", input.Mutation)
	}
	objectIndex := func(p *Object) int {
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
	dataIndex := func(p unsafe.Pointer) int {
		for i, d := range data {
			if unsafe.Pointer(d) == p {
				return i
			}
		}
		t.Fatal("unknown player data identity")
		return -2
	}
	infoIndex := func(p *Player) int {
		for i, d := range infos {
			if d == p {
				return i
			}
		}
		t.Fatal("unknown player identity")
		return -2
	}
	s := &Server{}
	s.SetFrame(input.Frame)
	state := func() map[string]any {
		return map[string]any{
			"caster": objectIndex(record.Caster16), "target": objectIndex(record.Target48), "level": record.Level, "timer": record.Frame68, "frame": s.Frame(),
			"data": [2]int{dataIndex(objects[0].UpdateData), dataIndex(objects[1].UpdateData)}, "players": [2]int{infoIndex(data[0].Player), infoIndex(data[1].Player)}, "indexes": [2]byte{infos[0].PlayerInd, infos[1].PlayerInd},
			"classes": [4]uint32{uint32(objects[0].ObjClass), uint32(objects[1].ObjClass), uint32(objects[2].ObjClass), uint32(objects[3].ObjClass)},
			"flags":   [4]uint32{uint32(objects[0].ObjFlags), uint32(objects[1].ObjFlags), uint32(objects[2].ObjFlags), uint32(objects[3].ObjFlags)}, "types": [2]uint16{objects[2].TypeInd, objects[3].TypeInd},
		}
	}
	changed := false
	mutate := func(service string) {
		if changed || !strings.HasPrefix(input.Mutation, service+"-") {
			return
		}
		changed = true
		switch input.Mutation {
		case "balance-caster-b", "mark-caster-b":
			record.Caster16 = objects[1]
		case "balance-data-b", "mark-data-b", "code-data-b", "unmark-data-b":
			objects[0].UpdateData = unsafe.Pointer(data[1])
		case "balance-info-b", "mark-info-b", "unmark-info-b":
			data[0].Player = infos[1]
		case "balance-target-b", "mark-target-b", "code-target-b", "unmark-target-b":
			record.Target48 = objects[3]
		case "balance-level":
			record.Level = 0xffffffff
		case "balance-frame":
			s.SetFrame(0xfffffff9)
		case "mark-index", "unmark-index":
			infos[0].PlayerInd = 255
		case "code-type-b":
			objects[2].TypeInd = 0xfe12
		case "send-timer":
			record.Frame68 = 0x12345678
		default:
			t.Fatal("unknown callback mutation")
		}
	}
	events := make([]map[string]any, 0)
	event := func(kind string, args []any) {
		events = append(events, map[string]any{"kind": kind, "args": args, "state": state()})
	}
	bits, err := hex.DecodeString(input.BalanceBits)
	if err != nil || len(bits) != 8 {
		t.Fatal("invalid binary64 balance", err)
	}
	h := spellTagServerHooks530160(s)
	h.loadBalance = func(key string) float64 {
		if key != "TagDurationPerLevel" {
			t.Fatal("balance key", key)
		}
		event("balance", []any{})
		mutate("balance")
		return math.Float64frombits(binary.LittleEndian.Uint64(bits))
	}
	h.mark = func(index int32, target *Object, flags uint32) {
		event("mark", []any{index, objectIndex(target), flags})
		mutate("mark")
	}
	h.unmark = func(index int32, target *Object, flags uint32) {
		event("unmark", []any{index, objectIndex(target), flags})
		mutate("unmark")
	}
	h.unitCode = func(target *Object) uint16 {
		i := objectIndex(target)
		event("code", []any{i})
		mutate("code")
		return input.Codes[i-2]
	}
	h.send = func(index int32, packet [7]byte) {
		event("send", []any{index, hex.EncodeToString(packet[:]), 0, 1})
		mutate("send")
	}
	for _, b := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(b.pointer) <= 0xffffffff {
			t.Fatal("C-owned Tag reference frame below4GiB")
		}
		b.before = append([]byte(nil), unsafe.Slice((*byte)(b.pointer), b.size)...)
	}
	var result any
	switch input.Kind {
	case "create":
		result = spellTagCreate530160(record, h)
	case "update":
		result = SpellTagUpdate530250(record)
	case "destroy":
		spellTagDestroy530270(record, h)
	}
	if input.Mutation != "none" && !changed {
		t.Fatal("declared mutation was not reached")
	}
	for n, b := range blocks {
		after := unsafe.Slice((*byte)(b.pointer), b.size)
		for i, value := range b.before {
			if value != after[i] && !b.allowed[i] {
				t.Fatal(fmt.Sprintf("native Tag frame%d nonoutput/canary byte%d changed", n, i))
			}
		}
	}
	raw, err := json.Marshal(map[string]any{"ret": result, "state": state(), "events": events})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
