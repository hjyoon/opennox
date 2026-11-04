package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"math"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"
)

type golemRawContract549E90 interface {
	GolemDeathNative549E90(bool, *server.GolemDeathState549E90, server.GolemDeathRuntime549E90) int32
}
type golemPlan549E90 struct {
	Position, Flags  []uint32
	Cache, FPS, Coop *uint32
	Divisor          *int32
	Mech, Rock, Mass []int
	Reenter          []any
	ReturnNil        bool `json:"return_nil"`
}
type golemInput549E90 struct {
	Name, Entry             string
	Unit                    int
	Coop, Cache, FPS        uint32
	Divisor, Parts, Seconds int32
	MechStop                int    `json:"mech_stop"`
	FloatBits               uint64 `json:"float_bits"`
	New                     map[string]int
	Script                  map[string]golemPlan549E90
}
type golemOutput549E90 struct {
	Name    string           `json:"name"`
	Result  *uint32          `json:"result"`
	Fault   map[string]any   `json:"fault"`
	Events  []map[string]any `json:"events"`
	Cache   uint32           `json:"cache"`
	Divisor uint32           `json:"divisor"`
	FPS     uint32           `json:"fps"`
	Coop    uint32           `json:"coop"`
	Heap    string           `json:"heap_sha256"`
	Tables  map[string]any   `json:"tables"`
}

func TestGolemDeathOriginal549E90(t *testing.T) {
	TestGolemDeathNativeLayout549E90(t)
	if _, ok := any((*server.Object)(nil)).(golemRawContract549E90); !ok {
		if os.Getenv("OPENNOX_GOLEM_RAW_REQUIRED") == "1" {
			t.Fatal("raw native contract not implemented")
		}
		t.Log("complete frozen original references compile before engine; no leaf comparison credited")
		return
	}
	var inputs []golemInput549E90
	trollDeathFixture54A270(t, "golem_death_independent_inputs_549e90.json.gz", "c9c5c2b03c74b9d6bd30c7e18e55b2edee5fde4f0627d8997c4da3fef8f0a791", "3f4255a6e51c9ef5e75b41d3e53d06a27be54cb701b33b66654032eeed36c0b8", &inputs)
	if len(inputs) != 164 {
		t.Fatal("complete independent input count")
	}
	for _, p := range []struct {
		name, encoded, decoded string
		nearest                bool
	}{
		{"pc53_tozero", "d8889990ae60f2c903b251f3d1fa110c4e313cbd83343b23cf12f19ce8564182", "f7a03b5cd44fd16045b22487d150b905c54a8e97e5b8ebd5e2d7e5a670abf2c7", false},
		{"pc53_nearest", "62b1e1875cbcbc996487eb6d3876cf5d3075040c6d59ca1672317a71f9563f64", "c56137eefe174e1826c66fdbe1574c40da5b17737edc4ef21eb81439488f7d78", true},
	} {
		t.Run(p.name, func(t *testing.T) {
			var outputs []golemOutput549E90
			trollDeathFixture54A270(t, "golem_death_native_original_549e90_"+p.name+".json.gz", p.encoded, p.decoded, &outputs)
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
							got := golemRunOriginal549E90(t, in, mode, p.nearest)
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
func golemRunOriginal549E90(t *testing.T, in golemInput549E90, mode string, nearest bool) (out golemOutput549E90) {
	t.Helper()
	base, free := alloc.Malloc(0x10000)
	defer free()
	buf := unsafe.Slice((*byte)(base), 0x10000)
	for i := range buf {
		buf[i] = byte(i*17 + (i>>7)*29)
	}
	canonical := bytes.Clone(buf)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(base) <= math.MaxUint32 {
		t.Fatal("C fixture must exceed4GiB")
	}
	offsets := [8]int{0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x9000, 0xa000}
	ptr := func(n int) unsafe.Pointer {
		if n < 0 {
			return golemOriginalBad549E90(n)
		}
		return unsafe.Add(base, offsets[n])
	}
	role := func(p unsafe.Pointer) int {
		for n := -2; n < 8; n++ {
			if p == ptr(n) {
				return n
			}
		}
		t.Fatalf("undeclared pointer %p", p)
		return -99
	}
	nativeOff := [4]int{int(unsafe.Offsetof(server.Object{}.ObjFlags)), int(unsafe.Offsetof(server.Object{}.PosVec)), int(unsafe.Offsetof(server.Object{}.Field27)), int(unsafe.Offsetof(server.Object{}.Field29))}
	originalOff := [4]int{16, 56, 108, 116}
	store := func(at int, v uint32) { binary.LittleEndian.PutUint32(buf[at:], v) }
	for n := range 6 {
		store(offsets[n]+nativeOff[0], 0x11223344+uint32(n))
		binary.LittleEndian.PutUint32(canonical[offsets[n]+16:], 0x11223344+uint32(n))
		for j, v := range []uint32{0x40e00000 + uint32(n), 0x41000000 + uint32(n)} {
			store(offsets[n]+nativeOff[1]+4*j, v)
			binary.LittleEndian.PutUint32(canonical[offsets[n]+56+4*j:], v)
		}
		for _, k := range []int{2, 3} {
			copy(buf[offsets[n]+nativeOff[k]:offsets[n]+nativeOff[k]+4], canonical[offsets[n]+originalOff[k]:offsets[n]+originalOff[k]+4])
		}
	}
	expected := bytes.Clone(buf)
	type guardedState struct {
		Before [16]byte
		Index  uint32
		After  [16]byte
	}
	g, freeState := alloc.New(guardedState{})
	defer freeState()
	*g = guardedState{Before: [16]byte{1, 2, 3}, Index: in.Cache, After: [16]byte{4, 5, 6}}
	beforeState := *g
	// Stock byte image includes the adjacent divisor, so index12 aliases its
	// least significant byte exactly, while negative/huge reads remain raw.
	massStorage, freeMass := alloc.Malloc(80)
	defer freeMass()
	massBuf := unsafe.Slice((*byte)(massStorage), 80)
	for i := range massBuf {
		massBuf[i] = 0x69
	}
	copy(massBuf[16:64], golemOriginalMassBytes549E90())
	massBase := unsafe.Add(massStorage, 32)
	divisor := (*uint32)(unsafe.Add(massBase, 12))
	*divisor = uint32(in.Divisor)
	expectedMass := bytes.Clone(massBuf)
	state := server.GolemDeathState549E90{Index: &g.Index, Divisor: divisor}
	fps, coop := in.FPS, in.Coop
	slotSize := int(unsafe.Sizeof(uintptr(0)))
	makeTable := func(ids []int) (unsafe.Pointer, []byte, []byte, func()) {
		address, release := alloc.Malloc(uintptr((len(ids) + 4) * slotSize))
		data := unsafe.Slice((*byte)(address), (len(ids)+4)*slotSize)
		for i := range data {
			data[i] = 0x96
		}
		for i, id := range ids {
			value := golemOriginalName549E90(id)
			if id == -3 {
				value = golemOriginalBadName549E90()
			}
			at := (i + 2) * slotSize
			if slotSize == 4 {
				binary.LittleEndian.PutUint32(data[at:], uint32(uintptr(value)))
			} else {
				binary.LittleEndian.PutUint64(data[at:], uint64(uintptr(value)))
			}
		}
		return unsafe.Add(address, 2*slotSize), data, bytes.Clone(data), release
	}
	mechIDs := []int{0, 1, 2, 0, 3, 3, 0, 0, -1}
	mechIDs[in.MechStop] = -1
	mech, mechBuf, expectedMech, freeMech := makeTable(mechIDs)
	defer freeMech()
	// The preceding non-pointer word denotes the original loaded-image empty
	// name in these controlled fixtures; this is not a native stock alias claim.
	rockSlots, rockBuf, expectedRock, freeRock := makeTable([]int{7, 4, 5, 6, 5, 6, 6, 5, 6, 6, 6, 6, 6, -3})
	defer freeRock()
	rock := unsafe.Add(rockSlots, slotSize)
	names := []string{"MechGear", "MechHead", "MechArm", "MechLeg", "BigRock", "MediumRock", "SmallRock", ""}
	pending := ""
	nameValue := func(p unsafe.Pointer) any {
		if p == nil {
			return nil
		}
		for i, name := range names {
			if p == golemOriginalName549E90(i) {
				return name
			}
		}
		pending = "name"
		raw := unsafe.Slice((*byte)(p), 48)
		end := bytes.IndexByte(raw, 0)
		if end < 0 {
			t.Fatal("unterminated name")
		}
		return string(raw[:end])
	}
	tableWrite := func(address unsafe.Pointer, data, expected []byte, index, name int) {
		value := golemOriginalName549E90(name)
		at := int(uintptr(golemOriginalSlotAddress549E90(address, int32(index))) - uintptr(unsafe.Pointer(&data[0])))
		if slotSize == 4 {
			binary.LittleEndian.PutUint32(data[at:], uint32(uintptr(value)))
		} else {
			binary.LittleEndian.PutUint64(data[at:], uint64(uintptr(value)))
		}
		copy(expected[at:at+slotSize], data[at:at+slotSize])
	}
	events := []map[string]any{}
	occ := map[string]int{}
	nextIndex := 0
	var invoke func(string, int) int32
	mutate := func(plan golemPlan549E90) {
		for k, values := range [][]uint32{plan.Flags, plan.Position} {
			if len(values) > 0 {
				n := int(values[0])
				for j, v := range values[1:] {
					at := offsets[n] + nativeOff[k] + 4*j
					store(at, v)
					binary.LittleEndian.PutUint32(expected[at:], v)
					binary.LittleEndian.PutUint32(canonical[offsets[n]+originalOff[k]+4*j:], v)
				}
			}
		}
		if plan.Cache != nil {
			g.Index = *plan.Cache
		}
		if plan.Divisor != nil {
			*divisor = uint32(*plan.Divisor)
			copy(expectedMass[44:48], massBuf[44:48])
		}
		if plan.FPS != nil {
			fps = *plan.FPS
		}
		if plan.Coop != nil {
			coop = *plan.Coop
		}
		if len(plan.Mech) > 0 {
			tableWrite(mech, mechBuf, expectedMech, plan.Mech[0], plan.Mech[1])
		}
		if len(plan.Rock) > 0 {
			tableWrite(rock, rockBuf, expectedRock, plan.Rock[0], plan.Rock[1])
		}
		if len(plan.Mass) > 0 {
			at := 32 + plan.Mass[0]
			massBuf[at] = byte(plan.Mass[1])
			expectedMass[at] = massBuf[at]
		}
	}
	begin := func(kind string) (map[string]any, golemPlan549E90, int) {
		seq := occ[kind]
		occ[kind] = seq + 1
		at := nextIndex
		nextIndex++
		return map[string]any{"kind": kind, "index": at, "cache": g.Index, "divisor": *divisor, "fps": fps, "coop": coop}, in.Script[fmt.Sprintf("%s:%d", kind, seq)], seq
	}
	finish := func(e map[string]any, plan golemPlan549E90, result any) {
		events = append(events, e)
		mutate(plan)
		if len(plan.Reenter) > 0 {
			entry := plan.Reenter[0].(string)
			n, err := plan.Reenter[1].(json.Number).Int64()
			if err != nil {
				t.Fatal(err)
			}
			unit := int(n)
			events = append(events, map[string]any{"kind": "nested", "index": e["index"], "entry": entry, "unit": unit})
			r := invoke(entry, unit)
			events = append(events, map[string]any{"kind": "nested-return", "index": e["index"], "result": r})
		}
		events = append(events, map[string]any{"kind": e["kind"].(string) + "-return", "index": e["index"], "result": result})
		pending = ""
	}
	sentinel := func(e map[string]any) uint32 { return 0xa5a50000 + uint32(e["index"].(int)) }
	checkOrigin := func(p server.GolemOrigin549E90) {
		if p.Address() != uintptr(unsafe.Pointer(p.Source()))+uintptr(nativeOff[1]) {
			t.Fatal("independent C position offset")
		}
		r := role(unsafe.Pointer(p.Source()))
		if r != in.Unit && r != 1 {
			t.Fatal("cached source identity", r)
		}
	}
	readOrigin := func(p server.GolemOrigin549E90) [2]uint32 {
		checkOrigin(p)
		pending = "origin"
		pos := p.Read()
		return [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}
	}
	motion := func(o *server.Object) [3]uint32 {
		pending = "object"
		return [3]uint32{uint32(o.ObjFlags), math.Float32bits(o.Field27), o.Field29}
	}
	rt := server.GolemDeathRuntime549E90{RoundToNearest: nearest}
	rt.FPS = func() uint32 { return fps }
	rt.MechName = func(index int32) unsafe.Pointer {
		pending = "mech-table"
		value := *(*unsafe.Pointer)(golemOriginalSlotAddress549E90(mech, index))
		pending = ""
		return value
	}
	rt.RockName = func(index int32) unsafe.Pointer {
		pending = "rock-table"
		value := *(*unsafe.Pointer)(golemOriginalSlotAddress549E90(rock, index))
		pending = ""
		return value
	}
	rt.Mass = func(index int32) byte {
		pending = "mass-table"
		value := *(*byte)(golemOriginalByteAddress549E90(massBase, index))
		pending = ""
		return value
	}
	rt.Flags = func() uint32 { e, p, _ := begin("flags"); r := coop; finish(e, p, r); return r }
	rt.RandomInt = func(lo, hi int32, site server.GolemRandomSite549E90) int32 {
		e, p, _ := begin("int")
		e["min"], e["max"] = lo, hi
		r := in.Seconds
		if site == server.GolemPartCount549E90 {
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
		if !p.ReturnNil && value != nil {
			if v, ok := in.New[value.(string)]; ok {
				r = v
			}
		}
		finish(e, p, r)
		return (*server.Object)(ptr(r))
	}
	rt.Point = func(radius float32, origin server.GolemOrigin549E90, output *types.Pointf) {
		e, p, seq := begin("point")
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
	rt.Force = func(origin server.GolemOrigin549E90, o *server.Object, force float32) uint32 {
		e, p, _ := begin("force")
		e["object"], e["origin"], e["force"] = role(unsafe.Pointer(o)), readOrigin(origin), math.Float32bits(force)
		e["motion"] = motion(o)
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.Audio = func(id uint32, o *server.Object, kind, arg uint32) uint32 {
		e, p, _ := begin("audio")
		if kind != 0 || arg != 0 {
			t.Fatal("audio arguments")
		}
		e["sound"], e["unit"] = id, role(unsafe.Pointer(o))
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.PointFX = func(id uint32, origin server.GolemOrigin549E90) uint32 {
		e, p, _ := begin("pointfx")
		checkOrigin(origin)
		e["effect"], e["origin_pointer"] = id, role(unsafe.Pointer(origin.Source()))
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	rt.Decay = func(o *server.Object, duration uint32) uint32 {
		e, p, _ := begin("decay")
		e["object"], e["duration"], e["motion"] = role(unsafe.Pointer(o)), duration, motion(o)
		r := sentinel(e)
		finish(e, p, r)
		return r
	}
	oldRT, oldState := golemOriginalRuntimeOverride549E90, golemOriginalStateOverride549E90
	golemOriginalRuntimeOverride549E90, golemOriginalStateOverride549E90 = &rt, &state
	defer func() { golemOriginalRuntimeOverride549E90, golemOriginalStateOverride549E90 = oldRT, oldState }()
	invoke = func(entry string, unit int) int32 {
		mech := entry == "549e90"
		obj := (*server.Object)(ptr(unit))
		if mode == "C" {
			if golemOriginalABI549E90(mech) != 64 {
				t.Fatal("typed public C entry required")
			}
			return golemOriginalC549E90(mech, obj)
		}
		return any(obj).(golemRawContract549E90).GolemDeathNative549E90(mech, &state, rt)
	}
	oldFault := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(oldFault)
	var panicValue any
	func() {
		defer func() { panicValue = recover() }()
		r := uint32(invoke(in.Entry, in.Unit))
		out.Result = &r
	}()
	// Keep recovery free of allocation until every raw callback frame unwinds.
	if r := panicValue; r != nil {
		kind, site := "memory", pending
		text := fmt.Sprint(r)
		if strings.Contains(text, "divide") || strings.Contains(text, "modulo") {
			kind, site = "arithmetic", "rock-modulo"
		} else if site == "" {
			site = "object"
		}
		if kind == "memory" && !strings.Contains(text, "invalid memory") && !strings.Contains(text, "fault") {
			t.Fatalf("not an actual native memory fault: %T %v", r, r)
		}
		out.Fault = map[string]any{"kind": kind, "site": site}
	}
	for n := range 6 {
		for _, k := range []int{0, 2, 3} {
			at := offsets[n] + nativeOff[k]
			v := binary.LittleEndian.Uint32(buf[at:])
			binary.LittleEndian.PutUint32(expected[at:], v)
			binary.LittleEndian.PutUint32(canonical[offsets[n]+originalOff[k]:], v)
		}
	}
	if !bytes.Equal(buf, expected) {
		t.Fatal("whole64KiB native heap or guards changed outside original fields")
	}
	if g.Before != beforeState.Before || g.After != beforeState.After {
		t.Fatal("native cache state guards")
	}
	if !bytes.Equal(massBuf, expectedMass) || !bytes.Equal(mechBuf, expectedMech) || !bytes.Equal(rockBuf, expectedRock) {
		t.Fatal("complete native tables/mass/divisor guards")
	}
	mechNames, rockNames := []any{}, []any{}
	for index := range 9 {
		mechNames = append(mechNames, nameValue(rt.MechName(int32(index))))
	}
	for index := range 12 {
		rockNames = append(rockNames, nameValue(rt.RockName(int32(index))))
	}
	masses := []int{}
	for index := range 12 {
		masses = append(masses, int(rt.Mass(int32(index))))
	}
	out.Name = in.Name
	out.Events = events
	out.Cache = g.Index
	out.Divisor = *divisor
	out.FPS = fps
	out.Coop = coop
	out.Heap = fmt.Sprintf("%x", sha256.Sum256(canonical))
	out.Tables = map[string]any{"mech": mechNames, "rock": rockNames, "mass": masses}
	return
}
