//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"runtime/debug"
	"sort"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

const ddiFixtureSHA4E0B30 = "af42e3ad8f63511cf562470cbf08c3814163a22ae533dfa36bd3048e2fbf651b"

var ddiRoles4E0B30 = []string{"target", "source", "weapon", "parent", "owner", "owner_alt", "target_update", "source_update", "alt_update", "weapon_use", "alt_use", "soundset", "sound_resource", "health"}

type ddiField4E0B30 struct {
	ID       int    `json:"id"`
	View     string `json:"view"`
	Offset   int    `json:"PE_offset"`
	Width    int    `json:"PE_storage_width"`
	Probes   []int  `json:"PE_probe_widths"`
	Semantic string `json:"semantic"`
}

func (f ddiField4E0B30) pointer() bool { return f.Semantic == "pointer" || f.Semantic == "callback" }

type ddiShape4E0B30 struct {
	Readable   *int `json:"readable"`
	Unaligned  int  `json:"unaligned"`
	Unreadable bool `json:"unreadable"`
}
type ddiRegion4E0B30 struct {
	Name         string `json:"name"`
	Address      uint32 `json:"address"`
	Size         int    `json:"size"`
	Fill         byte   `json:"fill"`
	Permissions  string `json:"permissions"`
	Initializers []struct {
		Offset int    `json:"offset"`
		Hex    string `json:"hex"`
	} `json:"initializers"`
}
type ddiInput4E0B30 struct {
	Name    string            `json:"name"`
	Args    [5]uint32         `json:"args"`
	Roles   map[string]uint32 `json:"roles"`
	Regions []ddiRegion4E0B30 `json:"regions"`
	Domain  struct {
		Shapes map[string]ddiShape4E0B30 `json:"shapes"`
	} `json:"domain"`
}
type ddiExpectedField4E0B30 struct {
	Role  string `json:"role"`
	Field int    `json:"field"`
	Hex   string `json:"hex"`
}
type ddiCase4E0B30 struct {
	Input    ddiInput4E0B30           `json:"input"`
	Regions  map[string]string        `json:"regions"`
	Fields   []ddiExpectedField4E0B30 `json:"fields"`
	InputSHA string                   `json:"reference_input_sha256"`
}
type ddiFixture4E0B30 struct {
	Schema string           `json:"schema"`
	Count  int              `json:"case_count"`
	Probes int              `json:"declared_probes"`
	Fields []ddiField4E0B30 `json:"fields"`
	Cases  []ddiCase4E0B30  `json:"cases"`
}

func ddiSHA4E0B30(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
func ddiUnsigned4E0B30(data []byte) uint64 {
	var word [8]byte
	copy(word[:], data)
	return binary.LittleEndian.Uint64(word[:])
}
func ddiBytes4E0B30(value uint64, width int) []byte {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], value)
	return append([]byte(nil), data[:width]...)
}
func ddiLoadFixture4E0B30() (*ddiFixture4E0B30, error) {
	raw, err := os.ReadFile("testdata/default_damage_parent_full_initial_4e0b30.json.gz")
	if err != nil {
		return nil, err
	}
	if ddiSHA4E0B30(raw) != ddiFixtureSHA4E0B30 {
		return nil, fmt.Errorf("fixture SHA")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer z.Close()
	var f ddiFixture4E0B30
	if err = json.NewDecoder(z).Decode(&f); err != nil {
		return nil, err
	}
	if f.Schema != "opennox-default-damage-parent-full-initial-fixture-v1" || f.Count != 1038 || len(f.Cases) != 1038 || len(f.Fields) != 27 || f.Probes != 158814 {
		return nil, fmt.Errorf("fixture dimensions")
	}
	names := map[string]bool{}
	for _, c := range f.Cases {
		if names[c.Input.Name] {
			return nil, fmt.Errorf("duplicate case")
		}
		names[c.Input.Name] = true
	}
	return &f, nil
}

type ddiAllocation4E0B30 struct {
	Map      unsafe.Pointer
	Base     int
	Expected []byte
	Label    string
}
type ddiToken4E0B30 struct {
	Kind       uint32
	Allocation uint32
	Offset     uint64
	Pointer    unsafe.Pointer
	Bits       uintptr
}
type ddiFixup4E0B30 struct {
	Destination uint32
	Field       uint32
	GoOffset    uint64
	Token       ddiToken4E0B30
}
type ddiProbe4E0B30 struct {
	ID         uint32
	Allocation uint32
	Field      uint32
	Width      uint32
	GoOffset   uint64
	Readable   bool
	Canonical  uint64
	Pointer    bool
}
type ddiGraph4E0B30 struct {
	Page        int
	Allocations []ddiAllocation4E0B30
	Mirrors     map[string]ddiAllocation4E0B30
	Tokens      map[uint32]ddiToken4E0B30
	Fields      []ddiExpectedField4E0B30
	Definitions []ddiField4E0B30
	Offsets     [27]uintptr
	Case        ddiCase4E0B30
	Fixups      []ddiFixup4E0B30
	Probes      []ddiProbe4E0B30
}

func (g *ddiGraph4E0B30) close() {
	for _, a := range g.Allocations {
		if a.Map != nil {
			ddiUnmap4E0B30(a.Map, g.Page)
		}
	}
	for _, a := range g.Mirrors {
		if a.Map != nil {
			ddiUnmap4E0B30(a.Map, g.Page)
		}
	}
	g.Allocations = nil
	g.Mirrors = nil
}
func (g *ddiGraph4E0B30) alloc(label string, base int, fill byte) (int, error) {
	if base < 0 || base > g.Page {
		return 0, fmt.Errorf("%s native base outside page", label)
	}
	p := ddiMap4E0B30(g.Page)
	if p == nil {
		return 0, fmt.Errorf("%s mmap", label)
	}
	if uintptr(p) <= math.MaxUint32 {
		ddiUnmap4E0B30(p, g.Page)
		return 0, fmt.Errorf("%s actual high C mapping required", label)
	}
	ddiFill4E0B30(p, g.Page, fill)
	expected := bytes.Repeat([]byte{fill}, g.Page)
	g.Allocations = append(g.Allocations, ddiAllocation4E0B30{Map: p, Base: base, Expected: expected, Label: label})
	return len(g.Allocations) - 1, nil
}
func (g *ddiGraph4E0B30) base(index int) unsafe.Pointer {
	return unsafe.Add(g.Allocations[index].Map, g.Allocations[index].Base)
}
func (g *ddiGraph4E0B30) original(role string, offset, width int) []byte {
	a, ok := g.Mirrors[role]
	if !ok {
		return nil
	}
	var region *ddiRegion4E0B30
	for i := range g.Case.Input.Regions {
		if g.Case.Input.Regions[i].Name == role {
			region = &g.Case.Input.Regions[i]
			break
		}
	}
	if region == nil {
		return nil
	}
	start := int(g.Case.Input.Roles[role]-region.Address) + offset
	if start < 0 || start >= region.Size {
		return nil
	}
	n := min(width, region.Size-start)
	return ddiSnapshot4E0B30(unsafe.Add(a.Map, a.Base+start), n)
}
func ddiRoleIndex4E0B30(role string) int {
	for i, r := range ddiRoles4E0B30 {
		if role == r {
			return i
		}
	}
	return -1
}
func ddiRoleView4E0B30(index int) string {
	if index < 6 {
		return "object"
	}
	if index < 9 {
		return "update-composite"
	}
	if index < 11 {
		return "wand"
	}
	if index == 11 {
		return "soundset"
	}
	return "opaque"
}
func ddiFullSpan4E0B30(index int) int {
	if index < 6 {
		return int(unsafe.Sizeof(server.Object{}))
	}
	if index < 9 {
		return max(int(unsafe.Sizeof(server.MonsterUpdateData{})), int(unsafe.Sizeof(server.PlayerUpdateData{})))
	}
	if index < 11 {
		return int(unsafe.Sizeof(server.WandUseData{}))
	}
	return 4096 - 256
}
func ddiBuildGraph4E0B30(f *ddiFixture4E0B30, c ddiCase4E0B30) (g *ddiGraph4E0B30, err error) {
	g = &ddiGraph4E0B30{Page: os.Getpagesize(), Mirrors: map[string]ddiAllocation4E0B30{}, Tokens: map[uint32]ddiToken4E0B30{}, Definitions: f.Fields, Fields: c.Fields, Offsets: rawParentInitialGoOffsets4E0B30(), Case: c}
	defer func() {
		if err != nil {
			g.close()
		}
	}()
	if unsafe.Sizeof(uintptr(0)) != 8 || g.Page < 4096 {
		return g, fmt.Errorf("local native64 page profile required")
	}
	// Actual named C/Go offsets, including independently compiled fixture C views.
	for id, off := range g.Offsets {
		co := ddiCOffset4E0B30(id)
		if id >= 21 && id <= 24 {
			if co != ^uintptr(0) {
				return g, fmt.Errorf("invented monster C offsetof")
			}
		} else if co != off {
			return g, fmt.Errorf("C/Go offsetof field%d", id)
		}
	}
	// Distinct selected fields must never collide in either representation.
	for _, view := range []string{"object", "update-composite", "wand", "soundset"} {
		peEnd, ne := 0, uintptr(0)
		for _, d := range f.Fields {
			if d.View != view {
				continue
			}
			nw := d.Width
			if d.pointer() {
				nw = 8
			}
			if d.Offset < peEnd || g.Offsets[d.ID] < ne {
				return g, fmt.Errorf("selected projection collision %s field%d", view, d.ID)
			}
			peEnd = d.Offset + d.Width
			ne = g.Offsets[d.ID] + uintptr(nw)
		}
	}
	regions := map[string]ddiRegion4E0B30{}
	for _, r := range c.Input.Regions {
		if r.Size != 4096 || r.Permissions != "rw" {
			return g, fmt.Errorf("unsupported input region geometry %s", r.Name)
		}
		p := ddiMap4E0B30(g.Page)
		if p == nil {
			return g, fmt.Errorf("PE mirror mmap")
		}
		ddiFill4E0B30(p, g.Page, 0x6D)
		m := ddiAllocation4E0B30{Map: p, Base: g.Page - 4096, Label: r.Name}
		g.Mirrors[r.Name] = m
		ddiFill4E0B30(unsafe.Add(p, m.Base), 4096, r.Fill)
		for _, init := range r.Initializers {
			v, e := hex.DecodeString(init.Hex)
			if e != nil || init.Offset < 0 || init.Offset+len(v) > 4096 {
				return g, fmt.Errorf("PE initializer")
			}
			ddiCopyIn4E0B30(unsafe.Add(p, m.Base+init.Offset), v)
		}
		actual := ddiSnapshot4E0B30(unsafe.Add(p, m.Base), 4096)
		if ddiSHA4E0B30(actual) != c.Regions[r.Name] {
			return g, fmt.Errorf("independent complete PE initial SHA %s", r.Name)
		}
		m.Expected = ddiSnapshot4E0B30(p, g.Page)
		g.Mirrors[r.Name] = m
		regions[r.Name] = r
	}
	for index, role := range ddiRoles4E0B30 {
		shape := c.Input.Domain.Shapes[role]
		cutoff := 0
		for _, ef := range c.Fields {
			if ef.Role != role {
				continue
			}
			d := f.Fields[ef.Field]
			data, e := hex.DecodeString(ef.Hex)
			if e != nil {
				return g, e
			}
			actual := g.original(role, d.Offset, d.Width)
			if !bytes.Equal(actual, data) {
				return g, fmt.Errorf("independent field bytes %s/%d", role, d.ID)
			}
			count := len(actual)
			if d.pointer() && count == 4 {
				count = 8
			}
			if count > 0 {
				cutoff = max(cutoff, int(g.Offsets[d.ID])+count)
			}
		}
		base := 256 + shape.Unaligned
		fill := byte(0x6B)
		if r, ok := regions[role]; ok {
			fill = r.Fill
			if shape.Readable == nil {
				base = int(c.Input.Roles[role] - r.Address)
			}
		}
		if shape.Unreadable {
			base = g.Page
		} else if shape.Readable != nil {
			base = g.Page - cutoff
		} else if base+ddiFullSpan4E0B30(index) > g.Page {
			return g, fmt.Errorf("full native type span crosses guard %s", role)
		}
		if _, err = g.alloc(role, base, fill); err != nil {
			return g, err
		}
		a := &g.Allocations[index]
		if m, ok := g.Mirrors[role]; ok {
			raw := ddiSnapshot4E0B30(unsafe.Add(m.Map, m.Base), 4096)
			ddiCopyIn4E0B30(a.Map, raw)
			copy(a.Expected, raw)
		}
		word := c.Input.Roles[role]
		if _, exists := g.Tokens[word]; word == 0 || exists {
			return g, fmt.Errorf("nil/aliased physical role declaration %s; argument aliases are separate", role)
		}
		g.Tokens[word] = ddiToken4E0B30{Kind: 1, Allocation: uint32(index), Offset: uint64(base), Pointer: g.base(index), Bits: uintptr(g.base(index))}
	}
	g.Tokens[0] = ddiToken4E0B30{}
	for _, kind := range []uint32{2, 3} {
		p, b := ddiCallback4E0B30(kind)
		if b <= math.MaxUint32 {
			return g, fmt.Errorf("full native callback identity")
		}
		word := uint32(0x23001000)
		if kind == 3 {
			word = 0x23002000
		}
		g.Tokens[word] = ddiToken4E0B30{Kind: kind, Pointer: p, Bits: b}
	}
	needed := map[uint32]bool{}
	for _, word := range c.Input.Args[:3] {
		needed[word] = true
	}
	for _, ef := range c.Fields {
		d := f.Fields[ef.Field]
		data, _ := hex.DecodeString(ef.Hex)
		if d.pointer() && len(data) == 4 {
			needed[uint32(ddiUnsigned4E0B30(data))] = true
		}
	}
	var unknown []uint32
	for value := range needed {
		if _, ok := g.Tokens[value]; !ok {
			unknown = append(unknown, value)
		}
	}
	sort.Slice(unknown, func(i, j int) bool { return unknown[i] < unknown[j] })
	for _, value := range unknown {
		for _, r := range c.Input.Regions {
			if value >= r.Address && uint64(value) < uint64(r.Address)+uint64(r.Size) {
				return g, fmt.Errorf("unsupported readable PE interior pointer %#x; no invented guard relocation", value)
			}
		}
		idx, e := g.alloc(fmt.Sprintf("unmapped-%08x", value), g.Page, 0xA6)
		if e != nil {
			return g, e
		}
		g.Tokens[value] = ddiToken4E0B30{Kind: 1, Allocation: uint32(idx), Offset: uint64(g.Page), Pointer: g.base(idx), Bits: uintptr(g.base(idx))}
	}
	identities := map[uintptr]uint32{}
	for word, token := range g.Tokens {
		if previous, exists := identities[token.Bits]; exists && previous != word {
			return g, fmt.Errorf("different canonical pointers share a native identity")
		}
		if uintptr(token.Pointer) != token.Bits {
			return g, fmt.Errorf("pointer/uintptr transport disagrees")
		}
		identities[token.Bits] = word
	}
	for _, ef := range c.Fields {
		index := ddiRoleIndex4E0B30(ef.Role)
		d := f.Fields[ef.Field]
		data, _ := hex.DecodeString(ef.Hex)
		a := &g.Allocations[index]
		off := int(g.Offsets[d.ID])
		count := len(data)
		if ddiRoleView4E0B30(index) != d.View {
			return g, fmt.Errorf("role/field view")
		}
		if d.pointer() && count == 4 {
			token := g.Tokens[uint32(ddiUnsigned4E0B30(data))]
			if a.Base+off+8 > g.Page {
				return g, fmt.Errorf("full pointer cannot fit cutoff %s/%d", ef.Role, d.ID)
			}
			if !ddiPutPointer4E0B30(g.base(index), d.ID, g.Offsets[d.ID], token.Pointer) {
				return g, fmt.Errorf("C typed pointer initializer")
			}
			copy(a.Expected[a.Base+off:a.Base+off+8], ddiBytes4E0B30(uint64(token.Bits), 8))
			g.Fixups = append(g.Fixups, ddiFixup4E0B30{uint32(index), uint32(d.ID), uint64(g.Offsets[d.ID]), token})
		} else if count > 0 {
			if a.Base+off+count > g.Page {
				return g, fmt.Errorf("literal field cannot fit cutoff %s/%d", ef.Role, d.ID)
			}
			if !ddiPut4E0B30(g.base(index), d.ID, g.Offsets[d.ID], data) {
				return g, fmt.Errorf("C scalar/prefix initializer")
			}
			copy(a.Expected[a.Base+off:a.Base+off+count], data)
		}
		for _, peWidth := range d.Probes {
			width := peWidth
			if d.pointer() {
				width = 8
			}
			readable := len(data) >= peWidth
			nativeReadable := a.Base+off+width <= g.Page
			if readable != nativeReadable {
				return g, fmt.Errorf("dynamic cutoff constraint collision %s field%d PE%d native%d accessible%d; no padding/skip permitted", ef.Role, d.ID, peWidth, width, len(data))
			}
			canonical := uint64(0)
			if readable {
				canonical = ddiUnsigned4E0B30(data[:peWidth])
			}
			g.Probes = append(g.Probes, ddiProbe4E0B30{uint32(len(g.Probes)), uint32(index), uint32(d.ID), uint32(width), uint64(g.Offsets[d.ID]), readable, canonical, d.pointer()})
		}
	}
	if len(g.Probes) != 153 {
		return g, fmt.Errorf("153 actual probe schedule required")
	}
	for _, a := range g.Allocations {
		if !bytes.Equal(ddiSnapshot4E0B30(a.Map, g.Page), a.Expected) {
			return g, fmt.Errorf("complete native initial buffer mismatch %s", a.Label)
		}
	}
	return g, nil
}

type ddiRead4E0B30 struct {
	Fault        bool
	Value        uint64
	FaultAddress uintptr
	Panic        string
}

func ddiGoRead4E0B30(base unsafe.Pointer, offset uintptr, width int) (out ddiRead4E0B30) {
	defer func() {
		if r := recover(); r != nil {
			out.Fault = true
			out.Panic = fmt.Sprint(r)
			if a, ok := r.(interface{ Addr() uintptr }); ok {
				out.FaultAddress = a.Addr()
			}
		}
	}()
	loadWidth := width
	if width == 8 {
		loadWidth = 0
	}
	out.Value = uint64(rawParentInitialGoBits4E0B30(base, offset, loadWidth))
	return
}
func ddiFaultMode4E0B30() func() {
	previous := debug.SetPanicOnFault(true)
	return func() { debug.SetPanicOnFault(previous) }
}
func (g *ddiGraph4E0B30) expectedNative(p ddiProbe4E0B30) uint64 {
	if p.Pointer {
		return uint64(g.Tokens[uint32(p.Canonical)].Bits)
	}
	return p.Canonical
}
func (g *ddiGraph4E0B30) checkCanaries() error {
	for _, a := range g.Allocations {
		if !bytes.Equal(ddiSnapshot4E0B30(a.Map, g.Page), a.Expected) {
			return fmt.Errorf("whole native buffer changed %s", a.Label)
		}
	}
	for _, a := range g.Mirrors {
		if !bytes.Equal(ddiSnapshot4E0B30(a.Map, g.Page), a.Expected) {
			return fmt.Errorf("whole canonical C mirror changed %s", a.Label)
		}
	}
	return nil
}
func (g *ddiGraph4E0B30) roundtrip() (map[string]string, error) {
	out := map[string][]byte{}
	for name, a := range g.Mirrors {
		out[name] = ddiSnapshot4E0B30(unsafe.Add(a.Map, a.Base), 4096)
	}
	for _, ef := range g.Fields {
		data, _ := hex.DecodeString(ef.Hex)
		if len(data) == 0 {
			continue
		}
		d := g.Definitions[ef.Field]
		index := ddiRoleIndex4E0B30(ef.Role)
		a := g.Allocations[index]
		actual := ddiSnapshot4E0B30(unsafe.Add(a.Map, a.Base+int(g.Offsets[d.ID])), len(data))
		if d.pointer() && len(data) == 4 {
			bits := rawParentInitialCBits4E0B30(g.base(index), g.Offsets[d.ID], 0)
			found := false
			for word, token := range g.Tokens {
				if token.Bits == bits {
					actual = ddiBytes4E0B30(uint64(word), 4)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("roundtrip unknown full native pointer")
			}
		}
		var r ddiRegion4E0B30
		for _, candidate := range g.Case.Input.Regions {
			if candidate.Name == ef.Role {
				r = candidate
				break
			}
		}
		start := int(g.Case.Input.Roles[ef.Role]-r.Address) + d.Offset
		copy(out[ef.Role][start:start+len(actual)], actual)
	}
	hashes := map[string]string{}
	for name, data := range out {
		hashes[name] = ddiSHA4E0B30(data)
		if hashes[name] != g.Case.Regions[name] {
			return nil, fmt.Errorf("full canonical PE roundtrip SHA %s", name)
		}
	}
	return hashes, nil
}

func ddiWrite32_4E0B30(w io.Writer, value uint32) error {
	return binary.Write(w, binary.LittleEndian, value)
}
func ddiWrite64_4E0B30(w io.Writer, value uint64) error {
	return binary.Write(w, binary.LittleEndian, value)
}
func (g *ddiGraph4E0B30) writeWorkerCase(w io.Writer, index uint32) error {
	for _, v := range []uint32{index, uint32(len(g.Allocations)), uint32(len(g.Fixups)), uint32(len(g.Probes))} {
		if e := ddiWrite32_4E0B30(w, v); e != nil {
			return e
		}
	}
	for i, a := range g.Allocations {
		for _, v := range []uint32{uint32(i), uint32(a.Base), uint32(g.Page)} {
			if e := ddiWrite32_4E0B30(w, v); e != nil {
				return e
			}
		}
		if _, e := w.Write(ddiSnapshot4E0B30(a.Map, g.Page)); e != nil {
			return e
		}
	}
	for _, fix := range g.Fixups {
		for _, v := range []uint32{fix.Destination, fix.Field} {
			if e := ddiWrite32_4E0B30(w, v); e != nil {
				return e
			}
		}
		if e := ddiWrite64_4E0B30(w, fix.GoOffset); e != nil {
			return e
		}
		for _, v := range []uint32{fix.Token.Kind, fix.Token.Allocation} {
			if e := ddiWrite32_4E0B30(w, v); e != nil {
				return e
			}
		}
		if e := ddiWrite64_4E0B30(w, fix.Token.Offset); e != nil {
			return e
		}
	}
	for _, p := range g.Probes {
		for _, v := range []uint32{p.ID, p.Allocation, p.Field, p.Width} {
			if e := ddiWrite32_4E0B30(w, v); e != nil {
				return e
			}
		}
		if e := ddiWrite64_4E0B30(w, p.GoOffset); e != nil {
			return e
		}
	}
	return nil
}
