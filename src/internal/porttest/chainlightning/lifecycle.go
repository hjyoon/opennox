// Package chainlightning projects original PE32 lifecycle records onto guarded,
// C-owned native objects. Only the three declared original services are modeled.
package chainlightning

import (
	"encoding/json"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type Input struct {
	Name       string    `json:"name"`
	Kind       string    `json:"kind"`
	Mutation   string    `json:"mutation"`
	Caster     int       `json:"caster"`
	Flag       uint32    `json:"flag"`
	Classes    [2]uint32 `json:"classes"`
	Equipped   [2]int    `json:"equipped"`
	Subclasses [2]uint32 `json:"subclasses"`
	WandFlags  [2]uint32 `json:"wand_flags"`
	Stored     int       `json:"stored"`
	Carry      uint32    `json:"carry"`
	Flags      uint32    `json:"flags"`
	Pos        [2]uint32 `json:"pos"`
	Sub104     []int     `json:"sub104"`
	Sub108     []int     `json:"sub108"`
}

type Fixture struct {
	OriginalSHA256 string `json:"original_sha256"`
	Rows           []struct {
		Input  Input           `json:"input"`
		Result json.RawMessage `json:"result"`
	} `json:"rows"`
}

type block struct {
	base    unsafe.Pointer
	size    int
	before  []byte
	allowed []bool
	free    func()
}

func allocate[T comparable](blocks *[]*block) (*T, *block) {
	type frame struct {
		Before [32]byte
		Value  T
		After  [32]byte
	}
	p, free := alloc.New(frame{})
	*p = frame{}
	for i := range p.Before {
		p.Before[i], p.After[i] = byte(0x55+i), byte(0xf1-i)
	}
	b := &block{base: unsafe.Pointer(p), size: int(unsafe.Sizeof(*p)), free: free}
	b.allowed = make([]bool, b.size)
	*blocks = append(*blocks, b)
	return &p.Value, b
}

func (b *block) allow(pointer unsafe.Pointer, size uintptr) {
	off := int(uintptr(pointer) - uintptr(b.base))
	for i := off; i < off+int(size); i++ {
		b.allowed[i] = true
	}
}

// Run invokes either the Go root or the actual native C entry against the same
// original inputs. Native pointer relocation and post-destroy sidecar release
// are explicit; all other record, object and adjacent bytes are checked.
func Run(input Input, invoke func(string, *server.DurSpell, server.SpellChainLightningRuntime52F820) uint32) (json.RawMessage, error) {
	var blocks []*block
	defer func() {
		for i := len(blocks) - 1; i >= 0; i-- {
			blocks[i].free()
		}
	}()
	record, rb := allocate[server.DurSpell](&blocks)
	var objects [4]*server.Object
	for i := range objects {
		objects[i], _ = allocate[server.Object](&blocks)
	}
	obj := func(i int) *server.Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	var players [2]*server.PlayerUpdateData
	var wands [2]*server.WandUseData
	for i := range players {
		var pb, wb *block
		players[i], pb = allocate[server.PlayerUpdateData](&blocks)
		wands[i], wb = allocate[server.WandUseData](&blocks)
		players[i].EquippedWeapon = obj(input.Equipped[i])
		pb.allow(unsafe.Pointer(&players[i].EquippedWeapon), unsafe.Sizeof(players[i].EquippedWeapon))
		objects[i].ObjClass, objects[i].UpdateData = object.Class(input.Classes[i]), unsafe.Pointer(players[i])
		objects[i+2].ObjSubClass = object.SubClass(input.Subclasses[i])
		objects[i+2].UseData.Ptr = unsafe.Pointer(wands[i])
		wands[i].Flags, wands[i].Charge, wands[i].MaxCharge = input.WandFlags[i], 7, 11
		wb.allow(unsafe.Pointer(&wands[i].Flags), unsafe.Sizeof(wands[i].Flags))
	}
	var rays [4]*server.DurSpell
	for i := range rays {
		var b *block
		rays[i], b = allocate[server.DurSpell](&blocks)
		b.allow(unsafe.Pointer(&rays[i].Next), unsafe.Sizeof(rays[i].Next))
	}
	link := func(chain []int) *server.DurSpell {
		if len(chain) == 0 {
			return nil
		}
		for i := 0; i+1 < len(chain); i++ {
			rays[chain[i]].Next = rays[chain[i+1]]
		}
		return rays[chain[0]]
	}
	*record = server.DurSpell{ID: 0xabcd, Spell: 43, Level: 5,
		Caster16: obj(input.Caster), Flag20: input.Flag,
		Pos:     types.Ptf(math.Float32frombits(input.Pos[0]), math.Float32frombits(input.Pos[1])),
		Field36: 0x13579bdf, Field40: 0x2468ace0, Frame60: 0x12345678,
		Field76: uintptr(input.Carry), Flags88: input.Flags,
		Sub104: link(input.Sub104), Sub108: link(input.Sub108)}
	for _, field := range []struct {
		p unsafe.Pointer
		n uintptr
	}{
		{unsafe.Pointer(&record.Caster16), unsafe.Sizeof(record.Caster16)},
		{unsafe.Pointer(&record.Flag20), unsafe.Sizeof(record.Flag20)},
		{unsafe.Pointer(&record.Pos), unsafe.Sizeof(record.Pos)},
		{unsafe.Pointer(&record.Field76), unsafe.Sizeof(record.Field76)},
		{unsafe.Pointer(&record.Flags88), unsafe.Sizeof(record.Flags88)},
		{unsafe.Pointer(&record.Sub104), unsafe.Sizeof(record.Sub104)},
		{unsafe.Pointer(&record.Sub108), unsafe.Sizeof(record.Sub108)},
	} {
		rb.allow(field.p, field.n)
	}
	stored := obj(input.Stored)
	objIndex := func(value *server.Object) int {
		if value == nil {
			return -1
		}
		for i, o := range objects {
			if o == value {
				return i
			}
		}
		panic("unexpected object reference")
	}
	rayIndex := func(value *server.DurSpell) int {
		if value == nil {
			return -1
		}
		for i, ray := range rays {
			if ray == value {
				return i
			}
		}
		panic("unexpected duration reference")
	}
	state := func(final bool) map[string]any {
		value := map[string]any{
			"caster": objIndex(record.Caster16), "flag": record.Flag20,
			"pos":   [2]uint32{math.Float32bits(record.Pos.X), math.Float32bits(record.Pos.Y)},
			"carry": uint32(record.Field76), "flags": record.Flags88,
			"sub104": rayIndex(record.Sub104), "sub108": rayIndex(record.Sub108),
			"classes":    [2]uint32{uint32(objects[0].ObjClass), uint32(objects[1].ObjClass)},
			"equipped":   [2]int{objIndex(players[0].EquippedWeapon), objIndex(players[1].EquippedWeapon)},
			"subclasses": [2]uint32{uint32(objects[2].ObjSubClass), uint32(objects[3].ObjSubClass)},
			"wand_flags": [2]uint32{wands[0].Flags, wands[1].Flags},
			"next":       [4]int{rayIndex(rays[0].Next), rayIndex(rays[1].Next), rayIndex(rays[2].Next), rayIndex(rays[3].Next)},
		}
		if !final || input.Kind != "destroy" {
			value["stored"] = objIndex(stored)
		}
		return value
	}
	mutated := false
	mutate := func(service string, ray *server.DurSpell) {
		if mutated || len(input.Mutation) <= len(service) || input.Mutation[:len(service)+1] != service+"-" {
			return
		}
		mutated = true
		switch input.Mutation {
		case "cancel-caster-b":
			record.Caster16 = objects[1]
		case "cancel-trap-on":
			record.Flag20 = 1
		case "cancel-trap-off":
			record.Flag20 = 0
		case "cancel-equipped-none":
			players[0].EquippedWeapon = nil
		case "cancel-equipped-b":
			players[0].EquippedWeapon = objects[3]
		case "cancel-clear-flag":
			wands[0].Flags &^= 4
		case "cancel-pos":
			record.Pos = types.Ptf(math.Float32frombits(0xc49a5000), math.Float32frombits(0x447a1000))
		case "cancel-flags":
			record.Flags88 = 0xa1b2c301
		case "free-next-nil":
			ray.Next = nil
		case "free-old-head":
			record.Sub104 = rays[3]
		case "free-weapon-b":
			stored = objects[3]
		case "free-weapon-nil":
			stored = nil
		case "free-flags":
			wands[0].Flags = 0xf1223344
		default:
			panic("unknown mutation")
		}
	}
	events := make([]map[string]any, 0)
	rt := server.SpellChainLightningRuntime52F820{
		CancelSpell: func(id int32, caster *server.Object) {
			events = append(events, map[string]any{"kind": "cancel", "args": []int{int(id), objIndex(caster)}, "state": state(false)})
			mutate("cancel", nil)
		},
		PointFX: func(code uint8, point types.Pointf) {
			events = append(events, map[string]any{"kind": "fx", "args": []uint32{uint32(code), math.Float32bits(point.X), math.Float32bits(point.Y)}, "state": state(false)})
		},
		FreeDuration: func(ray *server.DurSpell) {
			events = append(events, map[string]any{"kind": "free", "args": []int{rayIndex(ray)}, "state": state(false)})
			mutate("free", ray)
		},
		LoadWeapon:  func(*server.DurSpell) *server.Object { return stored },
		StoreWeapon: func(_ *server.DurSpell, wand *server.Object) { stored = wand },
	}
	for _, b := range blocks {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(b.base) <= 0xffffffff {
			return nil, fmt.Errorf("native frame below4GiB")
		}
		b.before = append([]byte(nil), unsafe.Slice((*byte)(b.base), b.size)...)
	}
	result := invoke(input.Kind, record, rt)
	if record.Field72 != 0 || uint64(record.Field76) > math.MaxUint32 {
		return nil, fmt.Errorf("native pointer/carry escaped PE32 scalar lanes")
	}
	if input.Kind == "destroy" && stored != nil {
		return nil, fmt.Errorf("destroy retained native wand sidecar")
	}
	if input.Mutation != "none" && !mutated {
		return nil, fmt.Errorf("declared service mutation was not reached")
	}
	for n, b := range blocks {
		after := unsafe.Slice((*byte)(b.base), b.size)
		for i, before := range b.before {
			if before != after[i] && !b.allowed[i] {
				return nil, fmt.Errorf("native frame%d nonoutput/canary byte%d changed", n, i)
			}
		}
	}
	return json.Marshal(map[string]any{"ret": result, "state": state(true), "events": events})
}
