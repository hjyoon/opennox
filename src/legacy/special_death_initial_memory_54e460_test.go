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
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type specialDeathInput54E460 struct {
	Name, Entry  string
	Unit         int
	Index        uint32
	Divisor      int32
	FPS, Frame   uint32
	Parts        int32
	Seconds      int32
	FloatBits    uint64 `json:"float_bits"`
	Quest        int32
	OwnerReturn  int    `json:"owner_return"`
	SourceOwner  int    `json:"source_owner"`
	ParentUpdate int    `json:"parent_update"`
	SourceUpdate int    `json:"source_update"`
	ParentClass  uint32 `json:"parent_class"`
	Slots        [4]int
	New          map[string]int
	Script       map[string]json.RawMessage
}

type specialDeathInitialWorld54E460 struct {
	Index, Divisor, FPS, Frame uint32
	Quest                      int32
	OwnerReturn                int `json:"owner_return"`
}

type specialDeathInitial54E460 struct {
	Name                    string
	Entry                   string
	Argument                uint32
	InitialHeapSHA256       string                         `json:"initial_heap_sha256"`
	SourceReads             map[string]*uint32             `json:"source_reads"`
	ParentReads             map[string]*uint32             `json:"parent_reads"`
	ObjectRoleReads         []map[string]*uint32           `json:"object_role_reads"`
	MarkerSlotDomains       map[string][]*uint32           `json:"marker_slot_domains"`
	ReturnedParentSlotReads []*uint32                      `json:"returned_parent_slot_reads"`
	InitialWorld            specialDeathInitialWorld54E460 `json:"initial_world"`
}

type specialDeathMemorySlot54E460 struct{ pe, native, width int }

// The observed fields have explicit PE32/native correspondences. The native
// allocation's remaining bytes are guarded, not certified as arbitrary aliases
// between differently widened object/update-data structures.
type specialDeathMemory54E460 struct {
	t                  *testing.T
	heap               unsafe.Pointer
	guard              uintptr
	buf, before, pe    []byte
	off                [29]uintptr
	slots              []specialDeathMemorySlot54E460
	peToNative         map[uint32]unsafe.Pointer
	nativeToPE         map[uintptr]uint32
	argument, parent   *server.Object
	world              specialDeathInitialWorld54E460
	actualDomainFaults int
}

var specialDeathRoleOffsets54E460 = [8]int{0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x9000, 0xa000}

func specialDeathInputs54E460(t *testing.T) ([]specialDeathInput54E460, []specialDeathInitial54E460) {
	t.Helper()
	var inputs []specialDeathInput54E460
	var initial []specialDeathInitial54E460
	trollDeathFixture54A270(t, "marker_boulder_generator_independent_inputs_54e460.json.gz", "33513f4d4be79a95526c408703bbe11ecb71bd7dfa69283c3d29f84d3dbf2cf8", "f9ea40aefdd2215f5b680df1ca90f6d5dcc170ffd84bd72350ad096e9444d0c3", &inputs)
	trollDeathFixture54A270(t, "marker_boulder_generator_initial_memory_54e460.json.gz", "6cb297188ca30a429f2414b80ba455710e3b078bca25a8dd10325a1108dd9570", "a2a4eec4e3b90b0dfd4b75606836bc8945c082151bbdf624084408bbfdc0b503", &initial)
	if len(inputs) != 200 || len(initial) != len(inputs) {
		t.Fatal("all independent inputs and whole initial-memory references required")
	}
	return inputs, initial
}

func specialDeathRolePE54E460(role int) uint32 {
	if role == -1 {
		return 0
	} else if role == -2 {
		return 0x41000000
	}
	return 0x21000000 + uint32(specialDeathRoleOffsets54E460[role])
}

func newSpecialDeathMemory54E460(t *testing.T, c specialDeathInput54E460, want specialDeathInitial54E460) *specialDeathMemory54E460 {
	t.Helper()
	heap, guard, free := generatorCreditOriginalHeap4D61B0()
	t.Cleanup(free)
	off := specialDeathOriginalLayout54E460()
	if heap == nil || guard < off[0]+512 || off[0]+2*off[28] >= 0x1000 || off[10] >= 0x1000 || off[13] >= 0x1000 || (off[28] == 8 && uintptr(heap) <= math.MaxUint32) {
		t.Fatal("whole C-owned high heap, complete native records and protected predecessor required")
	}
	m := &specialDeathMemory54E460{t: t, heap: heap, guard: guard, off: off, buf: unsafe.Slice((*byte)(heap), 0x10000), pe: make([]byte, 0x10000), peToNative: make(map[uint32]unsafe.Pointer), nativeToPE: make(map[uintptr]uint32)}
	for i := range m.pe {
		m.pe[i] = byte(i*17 + (i>>7)*29)
	}
	for _, offset := range specialDeathRoleOffsets54E460 {
		p := unsafe.Add(heap, offset)
		m.peToNative[0x21000000+uint32(offset)], m.nativeToPE[uintptr(p)] = p, 0x21000000+uint32(offset)
	}
	m.peToNative[0], m.nativeToPE[0] = nil, 0
	for role, offset := range specialDeathRoleOffsets54E460[:6] {
		for _, field := range [][2]uint32{{8, 2}, {16, 0x11223344 + uint32(role)}, {56, 0x40e00000 + uint32(role)}, {60, 0x41000000 + uint32(role)}, {520, specialDeathRolePE54E460(1)}, {748, specialDeathRolePE54E460(7)}} {
			binary.LittleEndian.PutUint32(m.pe[offset+int(field[0]):], field[1])
		}
	}
	for _, field := range [][2]uint32{{0x1000 + 748, specialDeathRolePE54E460(c.SourceUpdate)}, {0x1000 + 520, specialDeathRolePE54E460(c.SourceOwner)}, {0x2000 + 748, specialDeathRolePE54E460(c.ParentUpdate)}, {0x2000 + 8, c.ParentClass}} {
		binary.LittleEndian.PutUint32(m.pe[field[0]:], field[1])
	}
	for i, role := range c.Slots {
		binary.LittleEndian.PutUint32(m.pe[0x9000+116+4*i:], specialDeathRolePE54E460(role))
	}
	copy(m.buf, m.pe)
	for _, offset := range specialDeathRoleOffsets54E460[:6] {
		for i, peOff := range []int{8, 16, 56, 60, 108, 116} {
			m.scalar(offset+peOff, offset+int(off[1+i]))
		}
		m.pointer(offset+520, offset+int(off[8]))
		m.pointer(offset+748, offset+int(off[9]))
	}
	for _, offset := range specialDeathRoleOffsets54E460[6:] {
		// Generator header and callback scalars, plus the separately observed
		// PlayerUpdate marker slots. Unobserved collision fields are not a
		// general cross-structure alias correspondence.
		for i := 0; i < 12; i++ {
			m.pointer(offset+4*i, offset+int(off[28])*i)
		}
		for i := 0; i < 6; i++ {
			m.scalar(offset+48+4*i, offset+int(off[21])+4*i)
		}
		for i := 0; i < 4; i++ {
			m.pointer(offset+116+4*i, offset+int(off[11])+int(off[28])*i)
		}
	}
	m.argument, m.parent = (*server.Object)(m.fromPE(specialDeathRolePE54E460(c.Unit))), (*server.Object)(m.fromPE(specialDeathRolePE54E460(c.OwnerReturn)))
	m.world = specialDeathInitialWorld54E460{c.Index, uint32(c.Divisor), c.FPS, c.Frame, c.Quest, c.OwnerReturn}
	m.before = bytes.Clone(m.buf)
	if c.Name != want.Name || c.Entry != want.Entry || specialDeathRolePE54E460(c.Unit) != want.Argument || m.world != want.InitialWorld || fmt.Sprintf("%x", sha256.Sum256(m.pe)) != want.InitialHeapSHA256 || m.project() != want.InitialHeapSHA256 {
		t.Fatal("complete independent PE32 initial memory and native transport must match", c.Name)
	}
	return m
}

func (m *specialDeathMemory54E460) addSlot(pe, native, width int) {
	for _, slot := range m.slots {
		if native < slot.native+slot.width && slot.native < native+width {
			m.t.Fatal("declared native initial-memory fields overlap", pe, native, slot)
		}
	}
	m.slots = append(m.slots, specialDeathMemorySlot54E460{pe, native, width})
}

func (m *specialDeathMemory54E460) scalar(pe, native int) {
	m.addSlot(pe, native, 4)
	copy(m.buf[native:native+4], m.pe[pe:pe+4])
}

func (m *specialDeathMemory54E460) pointer(pe, native int) {
	m.addSlot(pe, native, int(m.off[28]))
	p := uintptr(m.fromPE(binary.LittleEndian.Uint32(m.pe[pe:])))
	for i := 0; i < int(m.off[28]); i++ {
		m.buf[native+i] = byte(p >> (8 * i))
	}
}

func (m *specialDeathMemory54E460) fromPE(pe uint32) unsafe.Pointer {
	if p, ok := m.peToNative[pe]; ok {
		return p
	}
	p := generatorCreditOriginalBad4D61B0(m.heap, m.guard, len(m.peToNative))
	if uintptr(p)+m.off[0]+2*m.off[28]+128 >= uintptr(m.heap) {
		m.t.Fatal("invalid pointer's complete native object must remain protected")
	}
	m.peToNative[pe], m.nativeToPE[uintptr(p)] = p, pe
	return p
}

func (m *specialDeathMemory54E460) toPE(p uintptr) uint32 {
	if pe, ok := m.nativeToPE[p]; ok {
		return pe
	}
	m.t.Fatalf("undeclared native pointer %#x", p)
	return 0
}

func (m *specialDeathMemory54E460) project() string {
	pe := bytes.Clone(m.pe)
	for _, slot := range m.slots {
		if slot.width == 4 && m.off[28] != 4 {
			copy(pe[slot.pe:slot.pe+4], m.buf[slot.native:slot.native+4])
			continue
		}
		// Pointer/scalar distinction must not be inferred from width on386.
		pointer := false
		for _, base := range specialDeathRoleOffsets54E460[:6] {
			pointer = pointer || slot.pe == base+520 || slot.pe == base+748
		}
		for _, base := range specialDeathRoleOffsets54E460[6:] {
			pointer = pointer || (slot.pe >= base && slot.pe < base+48) || (slot.pe >= base+116 && slot.pe < base+132)
		}
		if !pointer {
			copy(pe[slot.pe:slot.pe+4], m.buf[slot.native:slot.native+4])
			continue
		}
		var value uintptr
		for i := 0; i < int(m.off[28]); i++ {
			value |= uintptr(m.buf[slot.native+i]) << (8 * i)
		}
		binary.LittleEndian.PutUint32(pe[slot.pe:], m.toPE(value))
	}
	return fmt.Sprintf("%x", sha256.Sum256(pe))
}

func (m *specialDeathMemory54E460) readObject(unit *server.Object, key string) (out *uint32) {
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if value := recover(); value != nil {
			if _, ok := value.(interface{ RuntimeError() }); !ok {
				m.t.Fatalf("independent native read failed without a memory fault: %v", value)
			}
			m.actualDomainFaults++
			out = nil
		}
	}()
	var v uint32
	switch key {
	case "class_byte":
		v = uint32((*[unsafe.Sizeof(server.Object{})]byte)(unsafe.Pointer(unit))[m.off[1]])
	case "class_word":
		v = uint32(unit.ObjClass)
	case "flags_byte":
		v = uint32((*[unsafe.Sizeof(server.Object{})]byte)(unsafe.Pointer(unit))[m.off[2]])
	case "flags_word":
		v = uint32(unit.ObjFlags)
	case "x_bits":
		v = math.Float32bits(unit.PosVec.X)
	case "y_bits":
		v = math.Float32bits(unit.PosVec.Y)
	case "vertical_bits":
		v = math.Float32bits(unit.Field27)
	case "mass_bits":
		v = unit.Field29
	case "owner":
		v = m.toPE(uintptr(unsafe.Pointer(unit.Obj130)))
	case "update":
		v = m.toPE(uintptr(unit.UpdateData))
	default:
		m.t.Fatal("undeclared object read", key)
	}
	return &v
}

func (m *specialDeathMemory54E460) readParentSlot(index int) (out *uint32) {
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if value := recover(); value != nil {
			if _, ok := value.(interface{ RuntimeError() }); !ok {
				m.t.Fatalf("independent marker read failed without a memory fault: %v", value)
			}
			m.actualDomainFaults++
			out = nil
		}
	}()
	d := (*server.PlayerUpdateData)(m.parent.UpdateData)
	v := m.toPE(uintptr(unsafe.Pointer(d.Field29[index])))
	return &v
}

func TestSpecialDeathInitialBeforeNative54E460(t *testing.T) {
	if os.Getenv("OPENNOX_SPECIAL_DEATH_INITIAL_BEFORE") != "1" {
		return
	}
	for i := 0; i < 3; i++ {
		if specialDeathOriginalABI54E460(i) != 32 {
			t.Fatal("initial memory gate must precede all three raw native C entries", i)
		}
	}
}

func TestSpecialDeathInitialMemory54E460(t *testing.T) {
	TestSpecialDeathRawNativeLayout54E460(t)
	inputs, initial := specialDeathInputs54E460(t)
	for i, input := range inputs {
		t.Run(input.Name, func(t *testing.T) {
			m := newSpecialDeathMemory54E460(t, input, initial[i])
			check := func(got, want *uint32, key string) {
				if (got == nil) != (want == nil) || (got != nil && *got != *want) {
					t.Fatal("independent original/native initial field or fault domain", key, got, want)
				}
			}
			for _, view := range []struct {
				unit *server.Object
				want map[string]*uint32
			}{{m.argument, initial[i].SourceReads}, {m.parent, initial[i].ParentReads}} {
				for key, want := range view.want {
					check(m.readObject(view.unit, key), want, key)
				}
			}
			for role, offset := range specialDeathRoleOffsets54E460[:6] {
				u := (*server.Object)(unsafe.Add(m.heap, offset))
				for key, want := range initial[i].ObjectRoleReads[role] {
					check(m.readObject(u, key), want, key)
				}
				p, words := specialDeathOriginalObject54E460(u)
				if p[0] != uintptr(unsafe.Pointer(u)) || m.toPE(p[1]) != *initial[i].ObjectRoleReads[role]["update"] || m.toPE(p[3]) != *initial[i].ObjectRoleReads[role]["owner"] {
					t.Fatal("independent typed C object pointer getters", role)
				}
				for j, key := range []string{"class_word", "flags_word", "x_bits", "y_bits", "vertical_bits", "mass_bits"} {
					check(&words[j+1], initial[i].ObjectRoleReads[role][key], key)
				}
			}
			for role := 6; role < 8; role++ {
				offset := specialDeathRoleOffsets54E460[role]
				p := unsafe.Add(m.heap, offset)
				slots := specialDeathInitialSlotsC54E460(p)
				for j, v := range slots {
					got := m.toPE(v)
					check(&got, initial[i].MarkerSlotDomains[fmt.Sprint(role)][j], "typed C marker slot")
				}
				pointers, words := specialDeathInitialGeneratorC54E460(p)
				for j, v := range pointers {
					if m.toPE(v) != binary.LittleEndian.Uint32(m.pe[offset+4*j:]) {
						t.Fatal("typed C generator complete pointer header", role, j)
					}
				}
				for j, v := range words {
					if v != binary.LittleEndian.Uint32(m.pe[offset+48+4*j:]) {
						t.Fatal("typed C generator scalar/callback header", role, j)
					}
				}
			}
			for j, want := range initial[i].ReturnedParentSlotReads {
				check(m.readParentSlot(j), want, "returned parent marker slot")
			}
			if !bytes.Equal(m.buf, m.before) || m.project() != initial[i].InitialHeapSHA256 {
				t.Fatal("complete native guard and original whole-heap projection changed during reads")
			}
			if m.actualDomainFaults > 0 {
				t.Logf("INITIAL_NATIVE_MEMORY_FAULT_DOMAINS=%d", m.actualDomainFaults)
			}
		})
	}
}
