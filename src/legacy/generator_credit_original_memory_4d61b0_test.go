package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"runtime/debug"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type generatorCreditInput4D61B0 struct {
	Name, Unit, Update, Player, Alias string
	Flags, Counter                    uint32
	PlayerFlags                       uint32 `json:"player_flags"`
}

type generatorCreditInitial4D61B0 struct {
	Name              string
	ChosenPlayer      uint32 `json:"chosen_player"`
	Argument          uint32
	SelectedUpdate    uint32 `json:"selected_update"`
	SelectedPlayer    uint32 `json:"selected_player"`
	InitialHeapSHA256 string `json:"initial_heap_sha256"`
	Reads             map[string]*uint32
}

type generatorCreditSlot4D61B0 struct{ pe, native int }

// This explicitly declared PE32/native transport is frozen before the engine.
// It does not certify arbitrary aliases between differently widened records.
type generatorCreditMemory4D61B0 struct {
	t               *testing.T
	heap            unsafe.Pointer
	guard           uintptr
	buf, before, pe []byte
	chosenPE        uint32
	chosenNative    int
	off             [29]uintptr
	slots           []generatorCreditSlot4D61B0
	nativeToPE      map[uintptr]uint32
	peToNative      map[uint32]unsafe.Pointer
	argument        *server.Object
}

func generatorCreditInputs4D61B0(t *testing.T) ([]generatorCreditInput4D61B0, []generatorCreditInitial4D61B0) {
	t.Helper()
	var inputs []generatorCreditInput4D61B0
	var initial []generatorCreditInitial4D61B0
	trollDeathFixture54A270(t, "credit_independent_inputs_4d61b0.json.gz", "a086a617a8b90e951ac06183f77992dfb07f9c6b811757a36c668be3dff1e542", "4090b764c2022e8b8745aea57626e86eff624747d3f80518a366dff7aa4b1f0c", &inputs)
	trollDeathFixture54A270(t, "generator_credit_initial_memory_4d61b0.json.gz", "fc387a63cee081e37524766c90c9662b21380fac7ea79d5a82864b808e2d9887", "f308b03ed20bd665389fb663651cc9eed2079dff8cdb05c1bc839a13411de9f4", &initial)
	if len(inputs) != 286 || len(initial) != len(inputs) {
		t.Fatal("complete frozen independent initial memory")
	}
	return inputs, initial
}

func newGeneratorCreditMemory4D61B0(t *testing.T, c generatorCreditInput4D61B0, initial generatorCreditInitial4D61B0) *generatorCreditMemory4D61B0 {
	t.Helper()
	heap, guard, free := generatorCreditOriginalHeap4D61B0()
	t.Cleanup(free)
	if heap == nil || guard < unsafe.Sizeof(server.Player{})+256 || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(heap) <= math.MaxUint32) {
		t.Fatal("C-owned whole heap and predecessor guard large enough for a native Player")
	}
	m := &generatorCreditMemory4D61B0{t: t, heap: heap, guard: guard, buf: unsafe.Slice((*byte)(heap), 0x10000), pe: make([]byte, 0x10000), off: specialDeathOriginalLayout54E460(), chosenPE: initial.ChosenPlayer, nativeToPE: make(map[uintptr]uint32), peToNative: make(map[uint32]unsafe.Pointer)}
	for i := range m.pe {
		m.pe[i] = byte(i*17 + (i>>7)*29)
	}
	copy(m.buf, m.pe)
	m.chosenNative = 0x1000
	if c.Player == "unit" {
		m.chosenNative = 0x4000
	}
	switch c.Alias {
	case "counter-is-update-player":
		m.chosenNative = 0x9000 + int(m.off[12]) - int(m.off[17])
	case "counter-is-unit-update":
		m.chosenNative = 0x4000 + int(m.off[9]) - int(m.off[17])
	case "counter-is-unit-flags":
		m.chosenNative = 0x4000 + int(m.off[2]) - int(m.off[17])
	case "player-flags-is-update-player":
		m.chosenNative = 0x9000 + int(m.off[12]) - int(m.off[18])
	case "none":
	default:
		t.Fatal("undeclared alias", c.Alias)
	}
	register := func(pe uint32, offset int) {
		p := unsafe.Add(heap, offset)
		m.peToNative[pe], m.nativeToPE[uintptr(p)] = p, pe
	}
	register(0x21004000, 0x4000)
	register(0x21009000, 0x9000)
	register(initial.ChosenPlayer, m.chosenNative)
	m.peToNative[0], m.nativeToPE[0] = nil, 0
	putPE := func(address, value uint32) { binary.LittleEndian.PutUint32(m.pe[int(address-0x21000000):], value) }
	putNative := func(offset int, value uint32) { binary.LittleEndian.PutUint32(m.buf[offset:], value) }
	putPE(initial.ChosenPlayer+4668, c.Counter)
	putPE(initial.ChosenPlayer+4692, c.PlayerFlags)
	putNative(m.chosenNative+int(m.off[17]), c.Counter)
	putNative(m.chosenNative+int(m.off[18]), c.PlayerFlags)
	putPE(0x21009114, initial.SelectedPlayer)
	m.putPointer(0x9000+int(m.off[12]), m.fromPE(initial.SelectedPlayer))
	putPE(0x210042ec, initial.SelectedUpdate)
	m.putPointer(0x4000+int(m.off[9]), m.fromPE(initial.SelectedUpdate))
	putPE(0x21004010, c.Flags)
	putNative(0x4000+int(m.off[2]), c.Flags)
	m.slots = []generatorCreditSlot4D61B0{{0x9114, 0x9000 + int(m.off[12])}, {0x42ec, 0x4000 + int(m.off[9])}}
	if c.Update == "unit" {
		// The original reads an otherwise uninterpreted source word as a
		// Player pointer. Relocate its invalid value to a protected domain.
		m.slots = append(m.slots, generatorCreditSlot4D61B0{0x4114, 0x4000 + int(m.off[12])})
		m.putPointer(m.slots[2].native, m.fromPE(binary.LittleEndian.Uint32(m.pe[0x4114:])))
	}
	m.argument = (*server.Object)(m.fromPE(initial.Argument))
	m.before = bytes.Clone(m.buf)
	if c.Name != initial.Name || fmt.Sprintf("%x", sha256.Sum256(m.pe)) != initial.InitialHeapSHA256 || m.project() != initial.InitialHeapSHA256 {
		t.Fatal("independent whole initial PE32/native transport", c.Name)
	}
	return m
}

func (m *generatorCreditMemory4D61B0) putPointer(at int, p unsafe.Pointer) {
	v := uintptr(p)
	for i := 0; i < int(unsafe.Sizeof(v)); i++ {
		m.buf[at+i] = byte(v >> (8 * i))
	}
}

func (m *generatorCreditMemory4D61B0) fromPE(value uint32) unsafe.Pointer {
	if p, ok := m.peToNative[value]; ok {
		return p
	}
	p := generatorCreditOriginalBad4D61B0(m.heap, m.guard, len(m.peToNative))
	if uintptr(p)+m.off[18]+4 >= uintptr(m.heap) {
		m.t.Fatal("malformed pointer and full Player offset must remain protected")
	}
	m.peToNative[value], m.nativeToPE[uintptr(p)] = p, value
	return p
}

func (m *generatorCreditMemory4D61B0) toPE(value uintptr) uint32 {
	if pe, ok := m.nativeToPE[value]; ok {
		return pe
	}
	for _, basis := range []struct {
		native int
		pe     uint32
	}{{m.chosenNative, m.chosenPE}, {0x4000, 0x21004000}, {0x9000, 0x21009000}} {
		base := uintptr(m.heap) + uintptr(basis.native)
		if value >= base && value-base <= 32 {
			return basis.pe + uint32(value-base)
		}
	}
	m.t.Fatalf("undeclared native pointer value %#x", value)
	return 0
}

func (m *generatorCreditMemory4D61B0) pointerWord(at int) uintptr {
	var value uintptr
	for i := 0; i < int(unsafe.Sizeof(value)); i++ {
		value |= uintptr(m.buf[at+i]) << (8 * i)
	}
	return value
}

func (m *generatorCreditMemory4D61B0) project() string {
	pe := bytes.Clone(m.pe)
	for _, slot := range m.slots {
		binary.LittleEndian.PutUint32(pe[slot.pe:], m.toPE(m.pointerWord(slot.native)))
	}
	copy(pe[0x4010:0x4014], m.buf[0x4000+int(m.off[2]):0x4004+int(m.off[2])])
	for i, offset := range []int{4668, 4692} {
		at := int(m.chosenPE-0x21000000) + offset
		isPointer := false
		for _, slot := range m.slots {
			isPointer = isPointer || at == slot.pe
		}
		if !isPointer {
			native := m.chosenNative + int(m.off[17+i])
			copy(pe[at:at+4], m.buf[native:native+4])
		}
	}
	return fmt.Sprintf("%x", sha256.Sum256(pe))
}

func (m *generatorCreditMemory4D61B0) domainRead(key string) (out *uint32) {
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(interface{ RuntimeError() }); !ok {
				m.t.Fatalf("independent getter failed without a runtime memory fault: %v", failure)
			}
			out = nil
		}
	}()
	u := m.argument
	if u == nil {
		return nil
	}
	var v uint32
	switch key {
	case "flags":
		v = uint32(uint8(u.ObjFlags))
	case "update":
		v = m.toPE(uintptr(u.UpdateData))
	case "player":
		v = m.toPE(uintptr(unsafe.Pointer((*server.PlayerUpdateData)(u.UpdateData).Player)))
	case "counter", "player_flags":
		p := unsafe.Pointer((*server.PlayerUpdateData)(u.UpdateData).Player)
		v = *(*uint32)(generatorCreditOriginalField4D61B0(p, key == "player_flags"))
		// Four declared PE32 aliases overlap a pointer representation. Its
		// native low DWORD is address-dependent, so compare its relocated
		// pointer meaning, as in the whole-heap projection.
		at := int(m.chosenPE-0x21000000) + 4668
		if key == "player_flags" {
			at += 24
		}
		for _, slot := range m.slots {
			if at == slot.pe {
				v = m.toPE(m.pointerWord(slot.native))
			}
		}
	default:
		m.t.Fatal("undeclared independent getter", key)
	}
	return &v
}

func TestGeneratorCreditInitialMemory4D61B0(t *testing.T) {
	TestSpecialDeathRawNativeLayout54E460(t)
	inputs, initial := generatorCreditInputs4D61B0(t)
	for i, input := range inputs {
		t.Run(input.Name, func(t *testing.T) {
			m := newGeneratorCreditMemory4D61B0(t, input, initial[i])
			for _, key := range []string{"flags", "update", "player", "counter", "player_flags"} {
				got, want := m.domainRead(key), initial[i].Reads[key]
				if (got == nil) != (want == nil) || (got != nil && *got != *want) {
					t.Fatal("independent initial native read/fault correspondence", key, got, want)
				}
			}
			if !bytes.Equal(m.buf, m.before) || m.project() != initial[i].InitialHeapSHA256 {
				t.Fatal("whole initial heap changed during independent getters")
			}
		})
	}
}
