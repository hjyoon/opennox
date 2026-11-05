package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type specialDeathContract54E460 interface {
	SpecialDeathNative54E460(int, *server.GolemDeathState549E90, server.SpecialDeathRuntime54E460)
}

//go:noinline
func specialDeathContractOf54E460(value any) (specialDeathContract54E460, bool) {
	v, ok := value.(specialDeathContract54E460)
	return v, ok
}

type specialDeathOutput54E460 struct {
	Name           string
	Fault          *generatorCreditFault4D61B0
	Events         []map[string]json.RawMessage
	PartIndex      uint32 `json:"part_index"`
	Divisor        uint32
	FPS, Frame     uint32
	Quest          int32
	OwnerReturn    int `json:"owner_return"`
	Slots          [4]*int
	HeapSHA256     string                       `json:"heap_sha256"`
	ImageSHA256    string                       `json:"image_sha256"`
	OriginalWrites []generatorCreditWrite4D61B0 `json:"original_writes"`
}

type specialDeathBoundary54E460 struct {
	EventCursor    int                          `json:"event_cursor"`
	HeapSHA256     string                       `json:"heap_sha256"`
	ImageSHA256    string                       `json:"image_sha256"`
	ImageRegions   []string                     `json:"image_regions"`
	OriginalWrites []generatorCreditWrite4D61B0 `json:"original_writes"`
}

type specialDeathBoundaries54E460 struct {
	OriginalImageSHA256 string `json:"original_image_sha256"`
	Regions             []struct {
		Address uint32
		Bytes   string
	}
	Names          map[string]string
	NamedAddresses [4]uint32 `json:"named_addresses"`
	Outputs        []struct {
		Name       string
		Boundaries []specialDeathBoundary54E460
		ImageReads []struct {
			PC, Address string
			Data        *string
		} `json:"image_reads"`
	}
}

type specialDeathPlan54E460 struct {
	Index, FPS, Frame *uint32
	Divisor           *int32
	Quest             *int32
	OwnerReturn       *int `json:"owner_return"`
	ReturnRole        *int `json:"return_role"`
	Position          []uint32
	Flags, Class      []uint32
	Update, Owner     []int
	Slot, Table, Mass []int
	Reenter           []json.RawMessage
}

type specialDeathGuardedImage54E460 struct {
	buf, expected []byte
	base          unsafe.Pointer
}

func specialDeathImageAllocation54E460(t *testing.T, size int) specialDeathGuardedImage54E460 {
	p, free := alloc.Malloc(uintptr(size + 64))
	t.Cleanup(free)
	b := unsafe.Slice((*byte)(p), size+64)
	for i := range b {
		b[i] = byte(i*19 + 0x69)
	}
	return specialDeathGuardedImage54E460{buf: b, base: unsafe.Add(p, 32)}
}

type specialDeathWorld54E460 struct {
	t              *testing.T
	m              *specialDeathMemory54E460
	input          specialDeathInput54E460
	boundaries     []specialDeathBoundary54E460
	boundaryCursor int
	writesCursor   int
	events         []map[string]any
	occurrences    map[string]int
	serviceCount   int
	pending        string
	quest          int32
	ownerReturn    int
	state          server.GolemDeathState549E90
	rt             server.SpecialDeathRuntime54E460
	stateImage     specialDeathGuardedImage54E460
	massImage      specialDeathGuardedImage54E460
	tableImage     specialDeathGuardedImage54E460
	nameImages     []specialDeathGuardedImage54E460
	peNames        map[uint32]unsafe.Pointer
	nativeNames    map[uintptr]uint32
	regions        [5][]byte
	namedAddresses [4]uint32
	certificates   map[string]string
	run            func(string, *server.Object)
}

func specialDeathDecode54E460(t *testing.T, value json.RawMessage, target any) {
	t.Helper()
	d := json.NewDecoder(bytes.NewReader(value))
	d.UseNumber()
	if err := d.Decode(target); err != nil {
		t.Fatal(err)
	}
}

func specialDeathRegionKey54E460(regions []string) string {
	return strings.Join(regions, ":")
}

func newSpecialDeathWorld54E460(t *testing.T, input specialDeathInput54E460, initial specialDeathInitial54E460, reference specialDeathBoundaries54E460, index int, nearest bool) *specialDeathWorld54E460 {
	w := &specialDeathWorld54E460{t: t, m: newSpecialDeathMemory54E460(t, input, initial), input: input,
		boundaries: reference.Outputs[index].Boundaries, events: make([]map[string]any, 0),
		occurrences: make(map[string]int), quest: input.Quest, ownerReturn: input.OwnerReturn,
		peNames: make(map[uint32]unsafe.Pointer), nativeNames: make(map[uintptr]uint32), certificates: make(map[string]string)}
	if reference.OriginalImageSHA256 != "66eb0df5e37c97beb6c9b68fafa5a1d966cd4271347bd6c3453215684ff1a1c5" || len(reference.Regions) != 5 || len(reference.Outputs) != 200 || reference.Outputs[index].Name != input.Name {
		t.Fatal("independent original image transport and complete boundary identities")
	}
	for _, row := range reference.Outputs {
		for _, b := range row.Boundaries {
			key := specialDeathRegionKey54E460(b.ImageRegions)
			if old, ok := w.certificates[key]; ok && old != b.ImageSHA256 {
				t.Fatal("same complete original image patch has two image digests")
			}
			w.certificates[key] = b.ImageSHA256
		}
	}
	for i, region := range reference.Regions {
		data, err := hex.DecodeString(region.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		w.regions[i] = data
	}
	w.peNames[0], w.nativeNames[0] = nil, 0
	for name, payload := range reference.Names {
		var pe uint32
		if _, err := fmt.Sscanf(name, "0x%x", &pe); err != nil {
			t.Fatal(err)
		}
		data, err := hex.DecodeString(payload)
		if err != nil || len(data) != 48 {
			t.Fatal("complete original declared name-read payload", err)
		}
		a := specialDeathImageAllocation54E460(t, 48)
		copy(a.buf[32:], data)
		a.expected = bytes.Clone(a.buf)
		w.nameImages = append(w.nameImages, a)
		w.peNames[pe], w.nativeNames[uintptr(a.base)] = a.base, pe
	}
	w.stateImage = specialDeathImageAllocation54E460(t, 12)
	w.massImage = specialDeathImageAllocation54E460(t, len(w.regions[0]))
	w.tableImage = specialDeathImageAllocation54E460(t, len(w.regions[1])/4*int(w.m.off[28]))
	copy(w.massImage.buf[32:], w.regions[0])
	for slot := 0; slot < len(w.regions[1])/4; slot++ {
		pe := binary.LittleEndian.Uint32(w.regions[1][slot*4:])
		w.putPointer(w.tableImage.buf, 32+slot*int(w.m.off[28]), w.nameFromPE(pe))
	}
	w.state.Index = (*uint32)(w.stateImage.base)
	w.state.Divisor = (*uint32)(unsafe.Add(w.massImage.base, 28))
	*w.state.Index, *w.state.Divisor = input.Index, uint32(input.Divisor)
	*(*uint32)(unsafe.Add(w.stateImage.base, 4)), *(*uint32)(unsafe.Add(w.stateImage.base, 8)) = input.FPS, input.Frame
	w.stateImage.expected, w.massImage.expected, w.tableImage.expected = bytes.Clone(w.stateImage.buf), bytes.Clone(w.massImage.buf), bytes.Clone(w.tableImage.buf)
	w.namedAddresses = reference.NamedAddresses
	w.rt.RoundToNearest = nearest
	w.rt.DestroyedName = w.nameFromPE(reference.NamedAddresses[3])
	w.installRuntime()
	return w
}

func (w *specialDeathWorld54E460) putPointer(buf []byte, at int, pointer unsafe.Pointer) {
	for i := 0; i < int(w.m.off[28]); i++ {
		buf[at+i] = byte(uintptr(pointer) >> (8 * i))
	}
}

func (w *specialDeathWorld54E460) pointer(buf []byte, at int) uintptr {
	var p uintptr
	for i := 0; i < int(w.m.off[28]); i++ {
		p |= uintptr(buf[at+i]) << (8 * i)
	}
	return p
}

func (w *specialDeathWorld54E460) nameFromPE(pe uint32) unsafe.Pointer {
	if p, ok := w.peNames[pe]; ok {
		return p
	}
	p := w.m.fromPE(pe)
	w.peNames[pe], w.nativeNames[uintptr(p)] = p, pe
	return p
}

func (w *specialDeathWorld54E460) role(pointer unsafe.Pointer) any {
	pe := w.m.toPE(uintptr(pointer))
	for i := -2; i < 8; i++ {
		if specialDeathRolePE54E460(i) == pe {
			return i
		}
	}
	return nil
}

func (w *specialDeathWorld54E460) slots() [4]any {
	var out [4]any
	s := specialDeathInitialSlotsC54E460(unsafe.Add(w.m.heap, 0x9000))
	for i, v := range s {
		out[i] = w.role(w.m.fromPE(w.m.toPE(v)))
	}
	return out
}

func (w *specialDeathWorld54E460) imageRegions() []string {
	regions := []string{hex.EncodeToString(w.massImage.buf[32 : len(w.massImage.buf)-32])}
	table := bytes.Clone(w.regions[1])
	for slot := 0; slot < len(table)/4; slot++ {
		p := w.pointer(w.tableImage.buf, 32+slot*int(w.m.off[28]))
		pe, ok := w.nativeNames[p]
		if !ok {
			w.t.Fatal("undeclared native image name pointer", p)
		}
		binary.LittleEndian.PutUint32(table[slot*4:], pe)
	}
	regions = append(regions, hex.EncodeToString(table))
	for i := 0; i < 3; i++ {
		regions = append(regions, hex.EncodeToString(w.stateImage.buf[32+4*i:36+4*i]))
	}
	return regions
}

// Original image bytes outside the five declared regions are immutable in
// the independently executed oracle. A frozen, case-independent patch-to-full-
// image certificate projects native region bytes onto that original template.
// This does not hash unrelated native process/code memory as if it were PE32.
func (w *specialDeathWorld54E460) imageDigest() string {
	key := specialDeathRegionKey54E460(w.imageRegions())
	digest, ok := w.certificates[key]
	if !ok {
		w.t.Fatal("native image projection has no independently observed complete-image certificate", key)
	}
	return digest
}

func (w *specialDeathWorld54E460) checkBoundary() {
	if w.boundaryCursor >= len(w.boundaries) {
		w.t.Fatal("extra native world boundary")
	}
	b := w.boundaries[w.boundaryCursor]
	w.boundaryCursor++
	if b.EventCursor != len(w.events) || w.writesCursor > len(b.OriginalWrites) {
		w.t.Fatal("original/native write-prefix or world boundary cursor", b.EventCursor, len(w.events))
	}
	for _, write := range b.OriginalWrites[w.writesCursor:] {
		var address uint32
		if _, err := fmt.Sscanf(write.Address, "0x%x", &address); err != nil {
			w.t.Fatal(err)
		}
		if write.PC == "0x54e5fd" {
			if address != 0x834acc || *w.state.Index != write.Value {
				w.t.Fatal("original shared index write", write)
			}
			copy(w.stateImage.expected[32:36], w.stateImage.buf[32:36])
			continue
		}
		if address < 0x21000000 || address >= 0x21010000 {
			// Unicorn records the attempted invalid store before its fault. It
			// is not a write permission into any native mapped allocation.
			if write.PC != "0x54e57a" || address != 0x4100006c {
				w.t.Fatal("undeclared original invalid store", write)
			}
			continue
		}
		found := false
		for _, slot := range w.m.slots {
			if slot.pe != int(address-0x21000000) {
				continue
			}
			var got uint32
			if write.PC == "0x54e48c" {
				got = w.m.toPE(w.pointer(w.m.buf, slot.native))
			} else if write.PC == "0x54e57a" || write.PC == "0x54e5b0" || write.PC == "0x54e5b3" {
				got = binary.LittleEndian.Uint32(w.m.buf[slot.native:])
			} else {
				w.t.Fatal("undeclared original write instruction", write)
			}
			if got != write.Value {
				w.t.Fatal("native value at exact original prefix write destination", write, got)
			}
			copy(w.m.before[slot.native:slot.native+slot.width], w.m.buf[slot.native:slot.native+slot.width])
			found = true
			break
		}
		if !found {
			w.t.Fatal("original write has no independently frozen native field", write)
		}
	}
	w.writesCursor = len(b.OriginalWrites)
	if !bytes.Equal(w.m.buf, w.m.before) || w.m.project() != b.HeapSHA256 {
		w.t.Fatal("complete native heap guards or whole original heap projection at boundary", len(w.events), w.m.project(), b.HeapSHA256)
	}
	for _, a := range append([]specialDeathGuardedImage54E460{w.stateImage, w.massImage, w.tableImage}, w.nameImages...) {
		if !bytes.Equal(a.buf, a.expected) {
			w.t.Fatal("complete native image/name allocation guard at boundary", len(w.events))
		}
	}
	if !reflect.DeepEqual(w.imageRegions(), b.ImageRegions) || w.imageDigest() != b.ImageSHA256 {
		w.t.Fatal("complete declared image patches and independently certified original image digest", len(w.events))
	}
}

func (w *specialDeathWorld54E460) sequence(kind string) int {
	seq := w.occurrences[kind]
	w.occurrences[kind]++
	return seq
}

func (w *specialDeathWorld54E460) plan(kind string, seq int) specialDeathPlan54E460 {
	var p specialDeathPlan54E460
	if raw, ok := w.input.Script[fmt.Sprintf("%s:%d", kind, seq)]; ok {
		specialDeathDecode54E460(w.t, raw, &p)
	}
	return p
}

func (w *specialDeathWorld54E460) world(kind string, seq int, extra map[string]any, result any) {
	w.checkBoundary()
	at := w.serviceCount
	w.serviceCount++
	e := map[string]any{"kind": kind, "index": at, "part_index": *w.state.Index, "divisor": *w.state.Divisor,
		"fps": *(*uint32)(unsafe.Add(w.stateImage.base, 4)), "frame": *(*uint32)(unsafe.Add(w.stateImage.base, 8)), "quest": w.quest, "owner_return": w.ownerReturn}
	for k, v := range extra {
		e[k] = v
	}
	w.events = append(w.events, e)
	p := w.plan(kind, seq)
	w.mutate(p)
	if len(p.Reenter) != 0 {
		var entry string
		var unit int
		if len(p.Reenter) != 2 {
			w.t.Fatal("original reentry arity")
		}
		specialDeathDecode54E460(w.t, p.Reenter[0], &entry)
		specialDeathDecode54E460(w.t, p.Reenter[1], &unit)
		w.events = append(w.events, map[string]any{"kind": "nested", "index": at, "entry": entry, "unit": unit})
		w.run(entry, (*server.Object)(w.m.fromPE(specialDeathRolePE54E460(unit))))
		w.events = append(w.events, map[string]any{"kind": "nested-return", "index": at})
	}
	w.events = append(w.events, map[string]any{"kind": kind + "-return", "index": at, "result": result})
}

func (w *specialDeathWorld54E460) result() uint32 { return 0xa5a50000 + uint32(w.serviceCount) }

func (w *specialDeathWorld54E460) mutate(p specialDeathPlan54E460) {
	word := func(buf, expected []byte, at int, value uint32) {
		binary.LittleEndian.PutUint32(buf[at:], value)
		binary.LittleEndian.PutUint32(expected[at:], value)
	}
	for i, value := range []*uint32{p.Index, p.FPS, p.Frame} {
		if value != nil {
			word(w.stateImage.buf, w.stateImage.expected, 32+4*i, *value)
		}
	}
	if p.Divisor != nil {
		word(w.massImage.buf, w.massImage.expected, 60, uint32(*p.Divisor))
	}
	for _, f := range []struct {
		values []uint32
		offset uintptr
	}{{p.Position, w.m.off[3]}, {p.Flags, w.m.off[2]}, {p.Class, w.m.off[1]}} {
		if len(f.values) == 0 {
			continue
		}
		at := specialDeathRoleOffsets54E460[f.values[0]] + int(f.offset)
		for i, value := range f.values[1:] {
			word(w.m.buf, w.m.before, at+4*i, value)
		}
	}
	for _, f := range []struct {
		values []int
		offset uintptr
	}{{p.Update, w.m.off[9]}, {p.Owner, w.m.off[8]}} {
		if len(f.values) == 0 {
			continue
		}
		at := specialDeathRoleOffsets54E460[f.values[0]] + int(f.offset)
		pointer := w.m.fromPE(specialDeathRolePE54E460(f.values[1]))
		w.putPointer(w.m.buf, at, pointer)
		w.putPointer(w.m.before, at, pointer)
	}
	if len(p.Slot) != 0 {
		at := 0x9000 + int(w.m.off[11]) + p.Slot[0]*int(w.m.off[28])
		pointer := w.m.fromPE(specialDeathRolePE54E460(p.Slot[1]))
		w.putPointer(w.m.buf, at, pointer)
		w.putPointer(w.m.before, at, pointer)
	}
	if len(p.Table) != 0 {
		pe := uint32(0)
		if p.Table[1] >= 0 {
			pe = w.namedAddresses[p.Table[1]]
		} else if p.Table[1] == -2 {
			pe = 0x41000000
		}
		at := 32 + (p.Table[0]+4)*int(w.m.off[28])
		pointer := w.nameFromPE(pe)
		w.putPointer(w.tableImage.buf, at, pointer)
		w.putPointer(w.tableImage.expected, at, pointer)
	}
	if len(p.Mass) != 0 {
		at := 48 + p.Mass[0]
		w.massImage.buf[at], w.massImage.expected[at] = byte(p.Mass[1]), byte(p.Mass[1])
	}
	if p.Quest != nil {
		w.quest = *p.Quest
	}
	if p.OwnerReturn != nil {
		w.ownerReturn = *p.OwnerReturn
	}
}

func (w *specialDeathWorld54E460) installRuntime() {
	w.rt.FPS = func() uint32 { return *(*uint32)(unsafe.Add(w.stateImage.base, 4)) }
	w.rt.Frame = func() uint32 { return *(*uint32)(unsafe.Add(w.stateImage.base, 8)) }
	w.rt.RockName = func(index int32) unsafe.Pointer {
		w.pending = "name-table"
		address := uint32(0x5ce2b8) + uint32(index)*4
		var out unsafe.Pointer
		if address >= 0x5ce2a8 && address+4 <= 0x5ce300 {
			out = *(*unsafe.Pointer)(unsafe.Add(w.tableImage.base, int(address-0x5ce2a8)/4*int(w.m.off[28])))
		} else {
			out = *(*unsafe.Pointer)(w.m.fromPE(address))
		}
		w.pending = ""
		return out
	}
	w.rt.Mass = func(index int32) byte {
		w.pending = "mass-table"
		address := uint32(0x5cde18) + uint32(index)
		var out byte
		if address >= 0x5cde08 && address < 0x5cde48 {
			out = *(*byte)(unsafe.Add(w.massImage.base, address-0x5cde08))
		} else {
			out = *(*byte)(w.m.fromPE(address))
		}
		w.pending = ""
		return out
	}
	w.rt.Parent = func(unit *server.Object) *server.Object {
		seq := w.sequence("owner")
		role := w.ownerReturn
		if p := w.plan("owner", seq); p.ReturnRole != nil {
			role = *p.ReturnRole
		}
		w.world("owner", seq, map[string]any{"unit": w.role(unsafe.Pointer(unit))}, role)
		return (*server.Object)(w.m.fromPE(specialDeathRolePE54E460(role)))
	}
	w.rt.New = func(name unsafe.Pointer) *server.Object {
		seq := w.sequence("new")
		var value any
		if name != nil {
			w.pending = "name"
			data := unsafe.Slice((*byte)(name), 48)
			end := bytes.IndexByte(data, 0)
			if end < 0 {
				w.t.Fatal("unterminated original name")
			}
			value = string(data[:end])
			w.pending = ""
		}
		role := -1
		if text, ok := value.(string); ok {
			if n, ok := w.input.New[text]; ok {
				role = n
			}
		}
		if p := w.plan("new", seq); p.ReturnRole != nil {
			role = *p.ReturnRole
		}
		w.world("new", seq, map[string]any{"name": value}, role)
		return (*server.Object)(w.m.fromPE(specialDeathRolePE54E460(role)))
	}
	w.rt.RandomInt = func(lo, hi int32, _ server.GolemRandomSite549E90) int32 {
		seq := w.sequence("int")
		result := w.input.Seconds
		if lo == 20 && hi == 30 {
			result = w.input.Parts
		} else if lo != 45 || hi != 75 {
			w.t.Fatal("original RNG bounds", lo, hi)
		}
		w.world("int", seq, map[string]any{"min": lo, "max": hi}, result)
		return result
	}
	w.rt.RandomFloat = func(lo, hi float32) float64 {
		seq := w.sequence("float")
		w.world("float", seq, map[string]any{"min": math.Float32bits(lo), "max": math.Float32bits(hi)}, w.input.FloatBits)
		return math.Float64frombits(w.input.FloatBits)
	}
	w.rt.PointFX = func(effect uint32, origin server.GolemOrigin549E90) uint32 {
		seq, result := w.sequence("pointfx"), w.result()
		w.world("pointfx", seq, map[string]any{"effect": effect, "origin_pointer": w.role(unsafe.Pointer(origin.Source())), "slots": w.slots()}, result)
		return result
	}
	w.rt.GeneratorFX = func(origin server.SkeletonOrigin54A310, intensity uint32) uint32 {
		seq, result := w.sequence("generatorfx"), w.result()
		w.world("generatorfx", seq, map[string]any{"intensity": intensity, "origin_pointer": w.role(unsafe.Pointer(origin.Source()))}, result)
		return result
	}
	w.rt.Audio = func(id uint32, unit *server.Object, kind, arg uint32) uint32 {
		seq, result := w.sequence("audio"), w.result()
		if kind != 0 || arg != 0 {
			w.t.Fatal("original audio arguments")
		}
		w.world("audio", seq, map[string]any{"sound": id, "unit": w.role(unsafe.Pointer(unit))}, result)
		return result
	}
	w.rt.Delete = func(unit *server.Object) uint32 {
		seq, result := w.sequence("delete"), w.result()
		w.world("delete", seq, map[string]any{"unit": w.role(unsafe.Pointer(unit))}, result)
		return result
	}
	originValue := func(origin server.GolemOrigin549E90) [2]uint32 {
		w.pending = "origin"
		pos := origin.Read()
		w.pending = ""
		return [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}
	}
	motion := func(unit *server.Object) [3]uint32 {
		w.pending = "part"
		out := [3]uint32{uint32(unit.ObjFlags), math.Float32bits(unit.Field27), unit.Field29}
		w.pending = ""
		return out
	}
	w.rt.Point = func(radius float32, origin server.GolemOrigin549E90, out *types.Pointf) {
		seq, result := w.sequence("point"), w.result()
		pos := originValue(origin)
		w.world("point", seq, map[string]any{"radius": math.Float32bits(radius), "origin": pos}, result)
		*out = types.Ptf(math.Float32frombits(0x42c80000+uint32(seq)), math.Float32frombits(0x43480000+uint32(seq)))
	}
	w.rt.Create = func(unit, owner *server.Object, pos types.Pointf, extra uint32) uint32 {
		seq, result := w.sequence("create"), w.result()
		w.world("create", seq, map[string]any{"object": w.role(unsafe.Pointer(unit)), "owner": w.role(unsafe.Pointer(owner)), "position": [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}, "extra": extra}, result)
		return result
	}
	w.rt.Raise = func(unit *server.Object, height float32) uint32 {
		seq, result := w.sequence("raise"), w.result()
		w.world("raise", seq, map[string]any{"object": w.role(unsafe.Pointer(unit)), "height": math.Float32bits(height)}, result)
		return result
	}
	w.rt.Force = func(origin server.GolemOrigin549E90, unit *server.Object, force float32) uint32 {
		seq, result := w.sequence("force"), w.result()
		pos, state := originValue(origin), motion(unit)
		w.world("force", seq, map[string]any{"object": w.role(unsafe.Pointer(unit)), "origin": pos, "force": math.Float32bits(force), "motion": state}, result)
		return result
	}
	w.rt.Decay = func(unit *server.Object, duration uint32) uint32 {
		seq, result := w.sequence("decay"), w.result()
		state := motion(unit)
		w.world("decay", seq, map[string]any{"object": w.role(unsafe.Pointer(unit)), "duration": duration, "motion": state}, result)
		return result
	}
	w.rt.Timer = func(value uint32) uint32 {
		seq, result := w.sequence("timer"), w.result()
		w.world("timer", seq, map[string]any{"value": value}, result)
		return result
	}
	w.rt.Mode = func(value int32) uint32 {
		seq, result := w.sequence("mode"), w.result()
		w.world("mode", seq, map[string]any{"value": value}, result)
		return result
	}
	w.rt.Script = func(data server.SpecialDeathCallback54E460, caller, trigger *server.Object, event uint32) uint32 {
		seq, result := w.sequence("script"), w.result()
		w.world("script", seq, map[string]any{"callback_data_role": w.role(data.Data), "caller": w.role(unsafe.Pointer(caller)), "trigger": w.role(unsafe.Pointer(trigger)), "event": event}, result)
		return result
	}
	w.rt.Quest = func() int32 {
		seq, result := w.sequence("flags"), w.quest
		w.world("flags", seq, nil, result)
		return result
	}
	w.rt.Credit = func(unit *server.Object) uint32 {
		seq, result := w.sequence("credit"), w.result()
		w.world("credit", seq, map[string]any{"unit": w.role(unsafe.Pointer(unit))}, result)
		return result
	}
}

func (w *specialDeathWorld54E460) invoke(run func()) (fault *generatorCreditFault4D61B0) {
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if value := recover(); value != nil {
			stack := string(debug.Stack())
			if strings.Contains(stack, ".specialDeathBoulderModulo54E460(") {
				fault = &generatorCreditFault4D61B0{"arithmetic", "boulder-modulo"}
				return
			}
			if _, ok := value.(interface{ RuntimeError() }); !ok {
				panic(value)
			}
			if w.pending != "" {
				fault = &generatorCreditFault4D61B0{"memory", w.pending}
				return
			}
			for _, site := range []struct{ helper, site string }{
				{"specialDeathParentUpdate54E460", "parent-update"}, {"specialDeathMarkerSlot54E460", "marker-slot"},
				{"specialDeathSourceUpdate54E460", "source-update"}, {"specialDeathSourceOwner54E460", "source-owner"},
				{"specialDeathParentClass54E460", "parent-class"}, {"specialDeathSourceY54E460", "source-y"}, {"specialDeathSourceX54E460", "source-x"},
				{"specialDeathPartVertical54E460", "part"}, {"specialDeathPartFlags54E460", "part"}, {"specialDeathPartMass54E460", "part"},
			} {
				if strings.Contains(stack, "."+site.helper+"(") {
					fault = &generatorCreditFault4D61B0{"memory", site.site}
					return
				}
			}
			panic(fmt.Sprintf("unclassified native memory failure: %v\n%s", value, stack))
		}
	}()
	run()
	return
}

func TestSpecialDeathOriginalConsumersBefore54E460(t *testing.T) {
	if os.Getenv("OPENNOX_SPECIAL_DEATH_CONSUMERS_BEFORE") != "1" {
		return
	}
	if _, ok := specialDeathContractOf54E460((*server.Object)(nil)); ok {
		t.Fatal("whole consumers must be frozen before all three engines")
	}
	for i := 0; i < 3; i++ {
		if specialDeathOriginalABI54E460(i) != 32 {
			t.Fatal("whole consumers must precede all three typed C entries")
		}
	}
}

func TestSpecialDeathOriginal54E460(t *testing.T) {
	inputs, initial := specialDeathInputs54E460(t)
	for _, profile := range []struct{ name, encoded, decoded, boundaryEncoded, boundaryDecoded string }{
		{"tozero", "c575d3d1f06b4b963fc759609be895e71883b1473112780602c921d7865eeb5a", "6f790dea714ab3061aecdeed56eab8cea3d15c1c0ded1a90454795046cc6f297", "064c2fbf2ad02b904bd98cb33560c81d76bc36835ef83328c81b74ec64e16e96", "6cdb3f63caf36f14a3fa72ec9b249c3576c12a7325f821984fcb7828bd6e752a"},
		{"nearest", "7095f87e5535ebbf97792ea37d9fd573c0274ed9955cba062653e8ab694b1557", "542f20f9b67ab5fc1ae66d856d2771d8a839c7194ee8adb04ae30b11c32cbeb7", "0e034279c38a99d88431431bc050f9d849f6aad47d07146d4d4a081700482c77", "cb57267cdbc9d2f0ff68e6ac488e10a45cec8eb13714556c1b48ca2e0a31fee2"},
	} {
		t.Run(profile.name, func(t *testing.T) {
			var outputs []specialDeathOutput54E460
			var boundaries specialDeathBoundaries54E460
			trollDeathFixture54A270(t, "marker_boulder_generator_original_54e460_pc53_"+profile.name+".json.gz", profile.encoded, profile.decoded, &outputs)
			trollDeathFixture54A270(t, "marker_boulder_generator_boundaries_54e460_pc53_"+profile.name+".json.gz", profile.boundaryEncoded, profile.boundaryDecoded, &boundaries)
			if len(outputs) != len(inputs) || len(boundaries.Outputs) != len(inputs) {
				t.Fatal("all original complete outputs and whole-memory boundaries required")
			}
			for _, mode := range []string{"Go", "C"} {
				t.Run(mode, func(t *testing.T) {
					_, present := specialDeathContractOf54E460((*server.Object)(nil))
					if !present || specialDeathOriginalABI54E460(0) != 64 || specialDeathOriginalABI54E460(1) != 64 || specialDeathOriginalABI54E460(2) != 64 {
						t.Log("complete whole-output consumer compiled before raw native engines; no native comparison credit")
						return
					}
					for i, input := range inputs {
						t.Run(input.Name, func(t *testing.T) {
							want := outputs[i]
							if want.Name != input.Name || boundaries.Outputs[i].Name != input.Name {
								t.Fatal("independent input/output/boundary identity")
							}
							w := newSpecialDeathWorld54E460(t, input, initial[i], boundaries, i, profile.name == "nearest")
							oldState, oldRuntime := specialDeathOriginalStateOverride54E460, specialDeathOriginalRuntimeOverride54E460
							specialDeathOriginalStateOverride54E460, specialDeathOriginalRuntimeOverride54E460 = &w.state, &w.rt
							defer func() {
								specialDeathOriginalStateOverride54E460, specialDeathOriginalRuntimeOverride54E460 = oldState, oldRuntime
							}()
							w.run = func(entry string, unit *server.Object) {
								id, ok := map[string]int{"54e460": 0, "54e4b0": 1, "54e630": 2}[entry]
								if !ok {
									t.Fatal("undeclared raw reentry", entry)
								}
								if mode == "C" {
									specialDeathOriginalC54E460(id, unit)
								} else {
									runner, _ := specialDeathContractOf54E460(unit)
									runner.SpecialDeathNative54E460(id, &w.state, w.rt)
								}
							}
							fault := w.invoke(func() { w.run(input.Entry, w.m.argument) })
							if !reflect.DeepEqual(fault, want.Fault) {
								t.Fatal("actual native fault/site and original fault prefix", fault, want.Fault)
							}
							w.checkBoundary()
							if w.boundaryCursor != len(w.boundaries) || w.m.project() != want.HeapSHA256 || w.imageDigest() != want.ImageSHA256 || *w.state.Index != want.PartIndex || *w.state.Divisor != want.Divisor || w.rt.FPS() != want.FPS || w.rt.Frame() != want.Frame || w.quest != want.Quest || w.ownerReturn != want.OwnerReturn {
								t.Fatal("complete final native memory/image/world output", input.Name)
							}
							got, err := json.Marshal(w.events)
							if err != nil {
								t.Fatal(err)
							}
							var actual, expected any
							specialDeathDecode54E460(t, got, &actual)
							raw, _ := json.Marshal(want.Events)
							specialDeathDecode54E460(t, raw, &expected)
							if !reflect.DeepEqual(actual, expected) {
								t.Fatalf("complete ordered original world trace differs\ngot %s\nwant%s", got, raw)
							}
							var slots any
							raw, _ = json.Marshal(w.slots())
							specialDeathDecode54E460(t, raw, &slots)
							raw, _ = json.Marshal(want.Slots)
							specialDeathDecode54E460(t, raw, &expected)
							if !reflect.DeepEqual(slots, expected) {
								t.Fatal("complete final marker slots", slots, expected)
							}
						})
					}
				})
			}
		})
	}
}

func TestSpecialDeathWholeConsumerInitialTransport54E460(t *testing.T) {
	inputs, initial := specialDeathInputs54E460(t)
	for _, profile := range []struct{ name, encoded, decoded string }{
		{"tozero", "064c2fbf2ad02b904bd98cb33560c81d76bc36835ef83328c81b74ec64e16e96", "6cdb3f63caf36f14a3fa72ec9b249c3576c12a7325f821984fcb7828bd6e752a"},
		{"nearest", "0e034279c38a99d88431431bc050f9d849f6aad47d07146d4d4a081700482c77", "cb57267cdbc9d2f0ff68e6ac488e10a45cec8eb13714556c1b48ca2e0a31fee2"},
	} {
		t.Run(profile.name, func(t *testing.T) {
			var reference specialDeathBoundaries54E460
			trollDeathFixture54A270(t, "marker_boulder_generator_boundaries_54e460_pc53_"+profile.name+".json.gz", profile.encoded, profile.decoded, &reference)
			for i, input := range inputs {
				t.Run(input.Name, func(t *testing.T) {
					w := newSpecialDeathWorld54E460(t, input, initial[i], reference, i, profile.name == "nearest")
					if len(w.boundaries[0].OriginalWrites) != 0 || w.boundaries[0].EventCursor != 0 {
						t.Fatal("first original observation must precede native writes/worlds")
					}
					w.checkBoundary()
				})
			}
		})
	}
}
