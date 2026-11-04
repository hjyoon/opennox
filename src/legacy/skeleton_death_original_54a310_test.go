package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type skeletonRawContract54A310 interface {
	SkeletonDeathNative54A310(bool, *server.SkeletonDeathState54A310, server.SkeletonDeathRuntime54A310) int32
	SkeletonDebrisNative54A4C0(*server.SkeletonDeathState54A310, server.SkeletonDeathRuntime54A310) int32
	SkeletonDropNative54A390(unsafe.Pointer, [4]unsafe.Pointer, int32, server.SkeletonDeathRuntime54A310)
}
type skeletonPlan54A310 struct {
	Position, Class, Subclass, Flags []uint32
	Use                              []int
	Cache, FPS                       *uint32
	Divisor                          *int32
	Coop                             *uint32
	Reenter                          []any
	ReturnNil                        bool `json:"return_nil"`
}
type skeletonInput54A310 struct {
	Name, Entry                   string
	Unit, Type                    int
	Mods                          [4]int
	Charge                        int32
	Coop, Cache, FPS              uint32
	Divisor, Roll, Parts, Seconds int32
	FloatBits                     uint64 `json:"float_bits"`
	New                           map[string]int
	Classes, Subclasses           [6]uint32
	Uses                          [6]int
	Script                        map[string]skeletonPlan54A310
}
type skeletonOutput54A310 struct {
	Name                      string           `json:"name"`
	Result                    *uint32          `json:"result"`
	Fault                     map[string]any   `json:"fault"`
	Events                    []map[string]any `json:"events"`
	Cache, Divisor, FPS, Coop uint32
	Heap                      string `json:"heap_sha256"`
}

// Explicit JSON names keep this complete-output comparison identical to the
// producer, including all state, event prefixes, faults and the canonical heap.
func (o skeletonOutput54A310) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"name": o.Name, "result": o.Result, "fault": o.Fault, "events": o.Events, "cache": o.Cache, "divisor": o.Divisor, "fps": o.FPS, "coop": o.Coop, "heap_sha256": o.Heap})
}
func TestSkeletonDeathOriginal54A310(t *testing.T) {
	TestSkeletonDeathNativeLayout54A310(t)
	if _, ok := any((*server.Object)(nil)).(skeletonRawContract54A310); !ok {
		if os.Getenv("OPENNOX_SKELETON_RAW_REQUIRED") == "1" {
			t.Fatal("raw native contract not implemented")
		}
		t.Log("frozen complete original references compile before engine")
		return
	}
	var inputs []skeletonInput54A310
	trollDeathFixture54A270(t, "skeleton_death_independent_inputs_54a310.json.gz", "70a545ed723b29909a80fa5dc084869fad81d26566b316d395c44152f4cc4776", "8bef0035c28d964690e708763c3cc433731b68d77ea9e8168dc08019c6b0268f", &inputs)
	if len(inputs) != 169 {
		t.Fatal("independent complete input count")
	}
	for _, p := range []struct {
		name, encoded, decoded string
		nearest                bool
	}{
		{"pc53_tozero", "e0901b3aa56d2badd8dbe6225409be2d3f60f0a3388b056aa6ecb1ca0fea19d6", "9251f28ef713cfb8bb6a1ec7e37311a779411fad307ad6fb45d3a812737f7f7a", false},
		{"pc53_nearest", "3bdf9e9b9cda3474025e434e710c1ba859c687c5a38dc306c00afc7c76ab5245", "3b08c58856895bf3566123435f8e4f62fbb9088520c19403f516ba96587cd28e", true},
	} {
		t.Run(p.name, func(t *testing.T) {
			var outputs []skeletonOutput54A310
			trollDeathFixture54A270(t, "skeleton_death_native_original_54a310_"+p.name+".json.gz", p.encoded, p.decoded, &outputs)
			if len(outputs) != len(inputs) {
				t.Fatal("complete output count")
			}
			for _, mode := range []string{"Go", "C"} {
				t.Run(mode, func(t *testing.T) {
					for i, in := range inputs {
						t.Run(in.Name, func(t *testing.T) {
							want := outputs[i]
							if want.Name != in.Name {
								t.Fatal("input/output identity")
							}
							got := skeletonRunOriginal54A310(t, in, mode, p.nearest)
							a, err := json.Marshal(got)
							if err != nil {
								t.Fatal(err)
							}
							b, err := json.Marshal(want)
							if err != nil {
								t.Fatal(err)
							}
							if !bytes.Equal(a, b) {
								t.Fatalf("complete original output differs\ngot %s\nwant%s", a, b)
							}
						})
					}
				})
			}
		})
	}
}

func skeletonRunOriginal54A310(t *testing.T, in skeletonInput54A310, mode string, nearest bool) (out skeletonOutput54A310) {
	t.Helper()
	base, free := alloc.Malloc(0x10000)
	defer free()
	buf := unsafe.Slice((*byte)(base), 0x10000)
	canonical := make([]byte, len(buf))
	for i := range buf {
		buf[i] = byte(i*17 + (i>>7)*29)
	}
	copy(canonical, buf)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(base) <= math.MaxUint32 {
		t.Fatal("C-owned fixture must exceed4GiB")
	}
	offsets := [8]int{0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x9000, 0xa000}
	ptr := func(n int) unsafe.Pointer {
		if n < 0 {
			return skeletonOriginalBad54A310(n)
		}
		return unsafe.Add(base, offsets[n])
	}
	role := func(p unsafe.Pointer) int {
		for n := -2; n < 8; n++ {
			if p == ptr(n) {
				return n
			}
		}
		t.Fatalf("undeclared native pointer %p", p)
		return -99
	}
	pePtr := func(n int) uint32 {
		if n == -1 {
			return 0
		}
		if n == -2 {
			return 0x41000000
		}
		return 0x21000000 + uint32(offsets[n])
	}
	objs := [6]*server.Object{}
	for i := range objs {
		objs[i] = (*server.Object)(ptr(i))
	}
	nativeOff := [7]int{int(unsafe.Offsetof(server.Object{}.ObjClass)), int(unsafe.Offsetof(server.Object{}.ObjSubClass)), int(unsafe.Offsetof(server.Object{}.ObjFlags)), int(unsafe.Offsetof(server.Object{}.PosVec)), int(unsafe.Offsetof(server.Object{}.Field27)), int(unsafe.Offsetof(server.Object{}.Field29)), int(unsafe.Offsetof(server.Object{}.UseData))}
	originalOff := [7]int{8, 12, 16, 56, 108, 116, 736}
	store := func(at int, v uint32) { binary.LittleEndian.PutUint32(buf[at:], v) }
	storePtr := func(at int, p unsafe.Pointer) {
		if unsafe.Sizeof(uintptr(0)) == 4 {
			binary.LittleEndian.PutUint32(buf[at:], uint32(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint64(buf[at:], uint64(uintptr(p)))
		}
	}
	for n := range objs {
		for k, v := range []uint32{in.Classes[n], in.Subclasses[n], 0x11223344 + uint32(n)} {
			store(offsets[n]+nativeOff[k], v)
			binary.LittleEndian.PutUint32(canonical[offsets[n]+originalOff[k]:], v)
		}
		for j, v := range []uint32{0x40e00000 + uint32(n), 0x41000000 + uint32(n)} {
			store(offsets[n]+nativeOff[3]+4*j, v)
			binary.LittleEndian.PutUint32(canonical[offsets[n]+56+4*j:], v)
		}
		for _, k := range []int{4, 5} {
			copy(buf[offsets[n]+nativeOff[k]:offsets[n]+nativeOff[k]+4], canonical[offsets[n]+originalOff[k]:offsets[n]+originalOff[k]+4])
		}
		storePtr(offsets[n]+nativeOff[6], ptr(in.Uses[n]))
		binary.LittleEndian.PutUint32(canonical[offsets[n]+736:], pePtr(in.Uses[n]))
	}
	// C poison bytes are never assigned through Go pointer-typed fields: this
	// avoids the write barrier interpreting old C guard words as Go pointers.
	expected := bytes.Clone(buf)
	type guardedState struct {
		Before         [16]byte
		Index, Divisor uint32
		After          [16]byte
	}
	g, freeState := alloc.New(guardedState{Before: [16]byte{1, 2, 3}, Index: in.Cache, Divisor: uint32(in.Divisor), After: [16]byte{4, 5, 6}})
	defer freeState()
	beforeState := *g
	state := server.SkeletonDeathState54A310{Index: &g.Index, Divisor: &g.Divisor}
	fps, coop := in.FPS, in.Coop
	rt := server.SkeletonDeathRuntime54A310{RoundToNearest: nearest}
	rt.Names = server.SkeletonDeathNames54A310{Skull: skeletonOriginalName54A310(5), Sword: skeletonOriginalName54A310(8), WoodenShield: skeletonOriginalName54A310(9), SteelShield: skeletonOriginalName54A310(10), Power: skeletonOriginalName54A310(1), Material: skeletonOriginalName54A310(2)}
	names := []string{"TestItem", "WeaponPower1", "Material2", "FireEffect", "OtherEffect", "Skull", "ArmBone", "LegBone", "Sword", "WoodenShield", "SteelShield"}
	nameValue := func(p unsafe.Pointer) any {
		for i := -2; i < len(names); i++ {
			if p == skeletonOriginalName54A310(i) {
				if i == -1 {
					return nil
				}
				if i == -2 {
					t.Fatal("unexpected invalid name dereference")
				}
				return names[i]
			}
		}
		t.Fatal("undeclared name pointer", p)
		return nil
	}
	// The DWORD-scaled PE table aliases index7fffffff and ffffffff to its
	// preceding nil slot. Use native C pointer slots while keeping that wrap.
	table, freeTable := alloc.Malloc(4 * unsafe.Sizeof(uintptr(0)))
	defer freeTable()
	tableBuf := unsafe.Slice((*byte)(table), 4*unsafe.Sizeof(uintptr(0)))
	for i := range tableBuf {
		tableBuf[i] = 0
	}
	tablePtr := unsafe.Add(table, unsafe.Sizeof(uintptr(0)))
	for i, index := range []int{6, 7} {
		off := (i + 1) * int(unsafe.Sizeof(uintptr(0)))
		if unsafe.Sizeof(uintptr(0)) == 4 {
			binary.LittleEndian.PutUint32(tableBuf[off:], uint32(uintptr(skeletonOriginalName54A310(index))))
		} else {
			binary.LittleEndian.PutUint64(tableBuf[off:], uint64(uintptr(skeletonOriginalName54A310(index))))
		}
	}
	rt.BoneName = func(index int32) unsafe.Pointer { return skeletonOriginalBone54A310(tablePtr, index) }
	descriptors, freeDescriptors := alloc.Malloc(1024)
	defer freeDescriptors()
	descriptor := func(id int32) unsafe.Pointer { return unsafe.Add(descriptors, int(id)*8) }
	descriptorID := func(p unsafe.Pointer) int32 {
		if p == nil {
			return -1
		}
		for i := int32(0); i < 100; i++ {
			if p == descriptor(i) {
				return i
			}
		}
		t.Fatal("undeclared modifier descriptor")
		return -99
	}
	events := []map[string]any{}
	occ := map[string]int{}
	nextIndex := 0
	pending := ""
	var invoke func(string, int) int32
	var mutate func(skeletonPlan54A310)
	mutate = func(plan skeletonPlan54A310) {
		for k, vals := range [][]uint32{plan.Class, plan.Subclass, plan.Flags, plan.Position} {
			if len(vals) > 0 {
				n := int(vals[0])
				for j, v := range vals[1:] {
					at := offsets[n] + nativeOff[k] + 4*j
					store(at, v)
					binary.LittleEndian.PutUint32(expected[at:], v)
					binary.LittleEndian.PutUint32(canonical[offsets[n]+originalOff[k]+4*j:], v)
				}
			}
		}
		if len(plan.Use) > 0 {
			n, r := plan.Use[0], plan.Use[1]
			at := offsets[n] + nativeOff[6]
			storePtr(at, ptr(r))
			copy(expected[at:at+int(unsafe.Sizeof(uintptr(0)))], buf[at:at+int(unsafe.Sizeof(uintptr(0)))])
			binary.LittleEndian.PutUint32(canonical[offsets[n]+736:], pePtr(r))
		}
		if plan.Cache != nil {
			g.Index = *plan.Cache
		}
		if plan.Divisor != nil {
			g.Divisor = uint32(*plan.Divisor)
		}
		if plan.FPS != nil {
			fps = *plan.FPS
		}
		if plan.Coop != nil {
			coop = *plan.Coop
		}
	}
	begin := func(kind string) (map[string]any, skeletonPlan54A310, int) {
		seq := occ[kind]
		occ[kind] = seq + 1
		at := nextIndex
		nextIndex++
		return map[string]any{"kind": kind, "index": at, "cache": g.Index, "divisor": g.Divisor, "fps": fps, "coop": coop}, in.Script[fmt.Sprintf("%s:%d", kind, seq)], seq
	}
	finish := func(e map[string]any, plan skeletonPlan54A310, result any) {
		events = append(events, e)
		mutate(plan)
		if len(plan.Reenter) > 0 {
			entry := plan.Reenter[0].(string)
			number, err := plan.Reenter[1].(json.Number).Int64()
			if err != nil {
				t.Fatal(err)
			}
			unit := int(number)
			events = append(events, map[string]any{"kind": "nested", "index": e["index"], "entry": entry, "unit": unit})
			r := invoke(entry, unit)
			var ret any
			if entry != "54a390" {
				ret = r
			}
			events = append(events, map[string]any{"kind": "nested-return", "index": e["index"], "result": ret})
		}
		events = append(events, map[string]any{"kind": e["kind"].(string) + "-return", "index": e["index"], "result": result})
		pending = ""
	}
	sentinel := func(e map[string]any) uint32 { return uint32(0xa5a50000) + uint32(e["index"].(int)) }
	readOrigin := func(p *types.Pointf) [2]uint32 {
		pending = "origin"
		return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)}
	}
	checkOrigin := func(p *types.Pointf) {
		t.Helper()
		r := role(unsafe.Add(unsafe.Pointer(p), -int(nativeOff[3])))
		if r != in.Unit && r != 1 {
			t.Fatal("origin pointer identity", r)
		}
	}
	rt.FPS = func() uint32 { return fps }
	rt.Flags = func() uint32 { e, p, _ := begin("flags"); r := coop; finish(e, p, r); return r }
	rt.RandomInt = func(lo, hi int32, site server.SkeletonRandomSite54A310) int32 {
		e, p, _ := begin("int")
		e["min"], e["max"] = lo, hi
		r := in.Seconds
		if site == server.SkeletonRoll54A310 {
			r = in.Roll
		} else if site == server.SkeletonPartCount54A310 {
			r = in.Parts
		}
		finish(e, p, r)
		return r
	}
	rt.RandomFloat = func(lo, hi float32) float64 {
		e, p, _ := begin("float")
		e["min"], e["max"] = math.Float32bits(lo), math.Float32bits(hi)
		finish(e, p, in.FloatBits)
		return math.Float64frombits(in.FloatBits)
	}
	rt.New = func(name unsafe.Pointer) *server.Object {
		e, p, _ := begin("new")
		value := nameValue(name)
		e["name"] = value
		r := -1
		if value != nil {
			if v, ok := in.New[value.(string)]; ok {
				r = v
			}
		}
		finish(e, p, r)
		return (*server.Object)(ptr(r))
	}
	rt.Point = func(radius float32, origin, output *types.Pointf) {
		e, p, seq := begin("point")
		checkOrigin(origin)
		e["radius"], e["origin"] = math.Float32bits(radius), readOrigin(origin)
		output.X = math.Float32frombits(0x42c80000 + uint32(seq))
		output.Y = math.Float32frombits(0x43480000 + uint32(seq))
		finish(e, p, sentinel(e))
	}
	rt.Create = func(o, owner *server.Object, pos types.Pointf, extra uint32) uint32 {
		e, p, _ := begin("create")
		e["object"], e["owner"], e["position"], e["extra"] = role(unsafe.Pointer(o)), role(unsafe.Pointer(owner)), [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}, extra
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.Raise = func(o *server.Object, height float32) uint32 {
		e, p, _ := begin("raise")
		e["object"], e["height"] = role(unsafe.Pointer(o)), math.Float32bits(height)
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.Force = func(origin *types.Pointf, o *server.Object, force float32) uint32 {
		e, p, _ := begin("force")
		checkOrigin(origin)
		e["object"], e["origin"], e["force"] = role(unsafe.Pointer(o)), readOrigin(origin), math.Float32bits(force)
		pending = "object"
		e["motion"] = [3]uint32{uint32(o.ObjFlags), math.Float32bits(o.Field27), o.Field29}
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.Audio = func(id uint32, o *server.Object, kind, arg uint32) uint32 {
		e, p, _ := begin("audio")
		if kind != 0 || arg != 0 {
			t.Fatal("audio args")
		}
		e["sound"], e["unit"] = id, role(unsafe.Pointer(o))
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.PointFX = func(id uint32, origin *types.Pointf) uint32 {
		e, p, _ := begin("pointfx")
		checkOrigin(origin)
		e["effect"], e["origin_pointer"] = id, role(unsafe.Add(unsafe.Pointer(origin), -nativeOff[3]))
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.Decay = func(o *server.Object, duration uint32) uint32 {
		e, p, _ := begin("decay")
		e["object"], e["duration"] = role(unsafe.Pointer(o)), duration
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.ModifierID = func(name unsafe.Pointer) int32 {
		e, p, _ := begin("modid")
		v := nameValue(name)
		e["name"] = v
		r := int32(31)
		if v != nil {
			for i, s := range names {
				if v == s {
					r = int32(i) + 40
				}
			}
		}
		finish(e, p, r)
		return r
	}
	rt.ModifierDesc = func(id int32) unsafe.Pointer {
		e, p, _ := begin("moddesc")
		e["id"] = id
		r := id
		if p.ReturnNil {
			r = -1
		}
		finish(e, p, r)
		if r == -1 {
			return nil
		}
		return descriptor(r)
	}
	rt.Apply = func(o *server.Object, attrs *server.ModifierInitData) uint32 {
		e, p, _ := begin("apply")
		e["object"] = role(unsafe.Pointer(o))
		var ids [4]int32
		for i, v := range attrs.Modifiers {
			ids[i] = descriptorID(unsafe.Pointer(v))
		}
		e["modifiers"] = ids
		if attrs.Field16 != 0 {
			t.Fatal("uninitialized original scratch is now explicitly zero")
		}
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	sound := func(kind string, o *server.Object) uint32 {
		e, p, _ := begin(kind)
		e["object"] = role(unsafe.Pointer(o))
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.ArmorSound = func(o *server.Object) uint32 { return sound("armor", o) }
	rt.WeaponSound = func(o *server.Object) uint32 { return sound("weapon", o) }
	oldRT, oldState := skeletonOriginalRuntimeOverride54A310, skeletonOriginalStateOverride54A310
	skeletonOriginalRuntimeOverride54A310 = &rt
	skeletonOriginalStateOverride54A310 = &state
	defer func() { skeletonOriginalRuntimeOverride54A310 = oldRT; skeletonOriginalStateOverride54A310 = oldState }()
	invoke = func(entry string, unit int) int32 {
		obj := (*server.Object)(ptr(unit))
		name := skeletonOriginalName54A310(in.Type)
		var mods [4]unsafe.Pointer
		for i, n := range in.Mods {
			mods[i] = skeletonOriginalName54A310(n)
		}
		if mode == "C" {
			number := map[string]int{"54a310": 0, "54a390": 1, "54a4c0": 2, "54a750": 3}[entry]
			if skeletonOriginalABI54A310(number) != 64 {
				t.Fatal("typed actual C entry required")
			}
			return skeletonOriginalC54A310(number, obj, name, mods, in.Charge)
		}
		raw := any(obj).(skeletonRawContract54A310)
		switch entry {
		case "54a310":
			return raw.SkeletonDeathNative54A310(false, &state, rt)
		case "54a750":
			return raw.SkeletonDeathNative54A310(true, &state, rt)
		case "54a4c0":
			return raw.SkeletonDebrisNative54A4C0(&state, rt)
		case "54a390":
			raw.SkeletonDropNative54A390(name, mods, in.Charge, rt)
			return 1
		}
		t.Fatal("entry")
		return 0
	}
	oldFault := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(oldFault)
	func() {
		defer func() {
			if r := recover(); r != nil {
				kind, site := "memory", pending
				text := fmt.Sprint(r)
				if strings.Contains(text, "divide") || strings.Contains(text, "modulo") {
					kind, site = "arithmetic", "bone-modulo"
				} else if site == "" {
					site = "object"
					if in.Entry == "54a390" {
						site = "charge"
					}
				}
				if kind == "memory" && !strings.Contains(text, "invalid memory") && !strings.Contains(text, "fault") {
					t.Fatalf("not an actual native memory fault: %T %v", r, r)
				}
				out.Fault = map[string]any{"kind": kind, "site": site}
			}
		}()
		r := invoke(in.Entry, in.Unit)
		if in.Entry != "54a390" {
			v := uint32(r)
			out.Result = &v
		}
	}()
	// Reconcile only the original stores, then compare every byte of the entire
	// native buffer and the complete independently expected PE32 heap digest.
	for n := range objs {
		for _, k := range []int{2, 4, 5} {
			at := offsets[n] + nativeOff[k]
			v := binary.LittleEndian.Uint32(buf[at:])
			binary.LittleEndian.PutUint32(expected[at:], v)
			binary.LittleEndian.PutUint32(canonical[offsets[n]+originalOff[k]:], v)
		}
	}
	for _, n := range []int{6, 7} {
		at := offsets[n] + 1
		expected[at] = buf[at]
		canonical[at] = buf[at]
	}
	if !bytes.Equal(buf, expected) {
		t.Fatal("whole64KiB native heap or guards changed outside declared fields")
	}
	if g.Before != beforeState.Before || g.After != beforeState.After {
		t.Fatal("whole native state guard")
	}
	out.Name = in.Name
	out.Events = events
	out.Cache = g.Index
	out.Divisor = g.Divisor
	out.FPS = fps
	out.Coop = coop
	out.Heap = fmt.Sprintf("%x", sha256.Sum256(canonical))
	return
}
