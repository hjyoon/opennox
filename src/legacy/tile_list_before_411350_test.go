//go:build darwin && arm64

package legacy

// An independent input/original-BEFORE validator. It never calls the graph
// builder, reads its actual bytes as an expected baseline, or predicts results.
import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"unsafe"
)

func tlIndependentInputBefore411350(in *tlInput411350) (map[uint32][]byte, error) {
	if len(in.Regions) != 4 || len(in.EdgePages) != 2 || len(in.InitialWords) != 0 || len(in.Roles) != 8 {
		return nil, fmt.Errorf("independent input geometry")
	}
	argc := 2
	if in.Entry == 0x411350 {
		argc = 3
	} else if in.Entry != 0x4113a0 && in.Entry != 0x411490 {
		return nil, fmt.Errorf("input root")
	}
	if len(in.Args) != argc {
		return nil, fmt.Errorf("input arity")
	}
	pages := map[uint32][]byte{}
	physical := map[string]uint32{}
	nodeBases := map[uint32]bool{}
	roleNames := map[string]bool{"node0": true, "node1": true, "node2": true, "xy": true, "nil": true, "unmapped": true, "high_unmapped": true, "max_unmapped": true}
	for name := range in.Roles {
		if !roleNames[name] {
			return nil, fmt.Errorf("unknown pointer role")
		}
	}
	seedPage := func(seed byte) []byte {
		out := make([]byte, 4096)
		for i := range out {
			out[i] = byte(int(seed) + 29*i + 7*(i/16))
		}
		return out
	}
	locations := map[string]uint32{"node0": 0x21000000, "node1": 0x21004000, "node2": 0x21008000, "xy": 0x2100c000}
	for _, r := range in.Regions {
		page, ok := locations[r.Name]
		limit := 20
		kind := "node"
		if r.Name == "xy" {
			limit = 8
			kind = "xy"
		}
		if !ok || r.Page != page || r.Kind != kind || r.Guard != page+4096 || r.Prefix < 0 || r.Prefix > limit || pages[page] != nil {
			return nil, fmt.Errorf("independent region/name/prefix")
		}
		pages[page] = seedPage(r.Seed)
		base := r.Guard - uint32(r.Prefix)
		physical[r.Name] = base
		if kind == "node" {
			nodeBases[base] = true
		}
	}
	if in.Roles["nil"] != 0 || in.Roles["unmapped"] != 0x21f02000 || in.Roles["high_unmapped"] != 0x81002000 || in.Roles["max_unmapped"] != 0xffffffff || in.Roles["xy"] != physical["xy"] {
		return nil, fmt.Errorf("declared scalar/pointer roles")
	}
	for _, name := range []string{"node0", "node1", "node2"} {
		if !nodeBases[in.Roles[name]] {
			return nil, fmt.Errorf("only same-base node aliases")
		}
	}
	seen := map[uint32]bool{}
	for _, p := range in.EdgePages {
		if (p.Page != 0x862000 && p.Page != 0x863000) || seen[p.Page] {
			return nil, fmt.Errorf("independent EDGE pages")
		}
		seen[p.Page] = true
		if p.Mapped {
			pages[p.Page] = seedPage(p.Seed)
		}
	}
	allowed := func(address uint32, width int) bool {
		end := uint64(address) + uint64(width)
		for _, r := range in.Regions {
			if uint64(address) >= uint64(r.Guard)-uint64(r.Prefix) && end <= uint64(r.Guard) {
				return true
			}
		}
		for _, p := range in.EdgePages {
			if p.Mapped && uint64(address) >= uint64(p.Page) && end <= uint64(p.Page)+4096 {
				return true
			}
		}
		return false
	}
	for _, w := range in.InitialWrites {
		b, e := hex.DecodeString(w.Hex)
		if e != nil || len(b) < 1 || len(b) > 4 || !allowed(w.Address, len(b)) {
			return nil, fmt.Errorf("independent initializer span")
		}
		page := w.Address &^ uint32(4095)
		off := int(w.Address - page)
		if pages[page] == nil || off+len(b) > 4096 {
			return nil, fmt.Errorf("mapped initializer")
		}
		copy(pages[page][off:], b)
	}
	// Original BEFORE is captured after literal root frame setup, before CPU.
	stack := seedPage(in.StackSeed)
	binary.LittleEndian.PutUint32(stack[0x800:0x804], 0x24000000)
	for i, arg := range in.Args {
		binary.LittleEndian.PutUint32(stack[0x804+4*i:0x808+4*i], arg)
	}
	pages[0x22000000] = stack
	return pages, nil
}

func tlIndependentNativeBefore411350(in *tlInput411350, pages map[uint32][]byte, meta *tlMeta411350, nativePrefix int) ([]byte, error) {
	page := meta.Page
	if page != os.Getpagesize() || page < 8192 || page > 1<<20 || page&(page-1) != 0 || len(meta.Banks) != 6 || meta.NativeOnlyNextPrefix != nativePrefix {
		return nil, fmt.Errorf("independent native page/bank/mode shape")
	}
	var named tileListNativeNode411350
	goOffsets := [5]uintptr{unsafe.Offsetof(named.Value), unsafe.Offsetof(named.Preserved), unsafe.Offsetof(named.Edge), unsafe.Offsetof(named.Variation), unsafe.Offsetof(named.Next)}
	if goOffsets != [5]uintptr{0, 4, 8, 12, 16} || unsafe.Sizeof(named) != 24 || unsafe.Sizeof(unsafe.Pointer(nil)) != 8 || meta.GoOffsets != goOffsets || meta.CLayout != [7]uintptr{0, 4, 8, 12, 16, 24, 8} || tileListFixtureLayout411350() != meta.CLayout {
		return nil, fmt.Errorf("independent named C/Go layout and widths")
	}
	read := func(address uint32, width int) ([]byte, error) {
		p := pages[address&^uint32(4095)]
		off := int(address & 4095)
		if len(p) != 4096 || width < 1 || off+width > 4096 {
			return nil, fmt.Errorf("original BEFORE field span")
		}
		return p[off : off+width], nil
	}
	physical := map[uint32]int{}
	target := -1
	for i, r := range in.Regions {
		physical[r.Guard-uint32(r.Prefix)] = i
		if r.Guard-uint32(r.Prefix) == in.Roles["node0"] {
			target = i
		}
	}
	if nativePrefix != 0 {
		if nativePrefix < 20 || nativePrefix > 24 || in.Entry != 0x411350 || len(in.AfterReads) != 0 || len(in.AfterReturns) != 0 || target < 0 || in.Args[0] != in.Roles["node0"] {
			return nil, fmt.Errorf("native-only prefix scope")
		}
		for _, r := range in.Regions {
			if r.Kind == "node" {
				if r.Prefix != 20 {
					return nil, fmt.Errorf("native-only complete PE input")
				}
				b, e := read(r.Guard-uint32(r.Prefix)+16, 4)
				if e != nil || binary.LittleEndian.Uint32(b) != 0 {
					return nil, fmt.Errorf("native-only literal nil Next")
				}
			}
		}
	}
	// Only recorded actual native allocation addresses are variable metadata.
	// Geometry/field descriptors are derived independently, never copied as truth.
	type interval struct{ start, end uint64 }
	var ranges []interval
	reserve := func(start, size uint64) error {
		if start < 1<<32 || size == 0 || start > ^uint64(0)-size {
			return fmt.Errorf("native address range")
		}
		end := start + size
		for _, r := range ranges {
			if start < r.end && r.start < end {
				return fmt.Errorf("overlapping native allocations")
			}
		}
		ranges = append(ranges, interval{start, end})
		return nil
	}
	expectedBanks := make([]tlBank411350, 0, 6)
	for i, r := range in.Regions {
		cut := r.Prefix
		if r.Kind == "node" && cut == 20 {
			cut = 24
		}
		if nativePrefix != 0 && i == target {
			cut = nativePrefix
		}
		base := meta.Banks[i].Native
		if base <= 0xffffffff || base%uint64(page) != 0 {
			return nil, fmt.Errorf("high page-aligned native bank")
		}
		if e := reserve(base, uint64(2*page)); e != nil {
			return nil, e
		}
		expectedBanks = append(expectedBanks, tlBank411350{r.Name, r.Kind, r.Guard - uint32(r.Prefix), r.Prefix, cut, page - cut, base, true})
	}
	for j, ep := range in.EdgePages {
		i := 4 + j
		base := meta.Banks[i].Native
		if base <= 0xffffffff || base%uint64(page) != 0 {
			return nil, fmt.Errorf("high EDGE bank")
		}
		if e := reserve(base, uint64(2*page)); e != nil {
			return nil, e
		}
		expectedBanks = append(expectedBanks, tlBank411350{fmt.Sprintf("edge%08x", ep.Page), "edge", ep.Page, 4096, 4096, page - 4096, base, ep.Mapped})
	}
	if !reflect.DeepEqual(meta.Banks, expectedBanks) {
		return nil, fmt.Errorf("independent six-bank geometry/anchors")
	}
	var fields []tlField411350
	for j, ep := range in.EdgePages {
		if ep.Mapped {
			fields = append(fields, tlField411350{4 + j, ep.Page, 4096, page - 4096, 4096, 4096, false})
		}
	}
	for i, r := range in.Regions {
		count := 5
		if r.Kind == "xy" {
			count = 2
		}
		for j := 0; j < count; j++ {
			off := 4 * j
			accessible := r.Prefix - off
			if accessible < 0 {
				accessible = 0
			}
			if accessible > 4 {
				accessible = 4
			}
			pointer := r.Kind == "node" && j == 4
			width := 4
			if pointer {
				width = 8
			}
			fields = append(fields, tlField411350{i, r.Guard - uint32(r.Prefix) + uint32(off), 4, expectedBanks[i].Anchor + off, width, accessible, pointer})
		}
	}
	if !reflect.DeepEqual(meta.Fields, fields) {
		return nil, fmt.Errorf("independent DWORD/PTR8/BYTE-page field descriptors")
	}
	needed := map[uint32]bool{}
	for token := range physical {
		needed[token] = true
	}
	for _, token := range in.Roles {
		if token != 0 {
			needed[token] = true
		}
	}
	for _, r := range in.Regions {
		if r.Kind == "node" && r.Prefix == 20 {
			b, e := read(r.Guard-20+16, 4)
			if e != nil {
				return nil, e
			}
			if token := binary.LittleEndian.Uint32(b); token != 0 {
				needed[token] = true
			}
		}
	}
	for _, schedule := range [][]tlStimulus411350{in.AfterReads, in.AfterReturns} {
		for _, s := range schedule {
			for _, w := range s.Writes {
				data, e := hex.DecodeString(w.Hex)
				if e != nil || (len(data) != 1 && len(data) != 4) {
					return nil, fmt.Errorf("independent effect width")
				}
				for _, r := range in.Regions {
					if r.Kind != "node" {
						continue
					}
					slot := r.Guard - uint32(r.Prefix) + 16
					if uint64(w.Address) < uint64(slot)+4 && uint64(slot) < uint64(w.Address)+uint64(len(data)) {
						if w.Address != slot || len(data) != 4 || r.Prefix != 20 {
							return nil, fmt.Errorf("no partial Next effect")
						}
						if token := binary.LittleEndian.Uint32(data); token != 0 {
							needed[token] = true
						}
					}
				}
			}
		}
	}
	if len(meta.Identities) != len(needed) {
		return nil, fmt.Errorf("exact input-derived pointer registry")
	}
	var opaque []uint32
	values := map[uint64]bool{}
	for token := range needed {
		key := strconv.FormatUint(uint64(token), 10)
		value, ok := meta.Identities[key]
		if !ok || value <= 0xffffffff || values[value] {
			return nil, fmt.Errorf("unique high full-width identity")
		}
		values[value] = true
		if i, ok := physical[token]; ok {
			if value != expectedBanks[i].Native+uint64(expectedBanks[i].Anchor) {
				return nil, fmt.Errorf("known physical identity/alias owner")
			}
		} else {
			opaque = append(opaque, token)
			if value%uint64(page) != 0 || value < uint64(page) {
				return nil, fmt.Errorf("opaque guard alignment")
			}
			if e := reserve(value-uint64(page), uint64(2*page)); e != nil {
				return nil, e
			}
		}
	}
	for key := range meta.Identities {
		v, e := strconv.ParseUint(key, 10, 32)
		if e != nil || v == 0 || strconv.FormatUint(v, 10) != key || !needed[uint32(v)] {
			return nil, fmt.Errorf("unknown/noncanonical identity key")
		}
	}
	sort.Slice(opaque, func(i, j int) bool { return opaque[i] < opaque[j] })
	if !reflect.DeepEqual(meta.Opaque, opaque) {
		return nil, fmt.Errorf("exact high opaque guard registry")
	}
	if meta.EdgeFault <= 0xffffffff || meta.EdgeFault%uint64(page) != 0 {
		return nil, fmt.Errorf("EDGE guard identity")
	}
	if e := reserve(meta.EdgeFault-uint64(page), uint64(2*page)); e != nil {
		return nil, e
	}
	// Every expected native byte starts from a fixed canary, not actual BEFORE.
	expected := make([]byte, 6*page)
	for i := 0; i < 6; i++ {
		for j := 0; j < page; j++ {
			expected[i*page+j] = byte(0xa1 + i)
		}
	}
	for _, f := range fields {
		if f.Accessible == 0 {
			continue
		}
		data, e := read(f.PE, f.Accessible)
		if e != nil {
			return nil, e
		}
		value := append([]byte(nil), data...)
		if f.Pointer && f.Accessible == 4 {
			token := binary.LittleEndian.Uint32(data)
			pointer := uint64(0)
			if token != 0 {
				pointer = meta.Identities[strconv.FormatUint(uint64(token), 10)]
			}
			value = make([]byte, 8)
			binary.LittleEndian.PutUint64(value, pointer)
		}
		available := page - f.NativeOffset
		if len(value) > available {
			if nativePrefix == 0 || f.Bank != target || !f.Pointer || available < 0 {
				return nil, fmt.Errorf("independent native field bound")
			}
			value = value[:available]
		}
		if f.NativeOffset < 0 || f.NativeOffset+len(value) > page {
			return nil, fmt.Errorf("independent native copy extent")
		}
		copy(expected[f.Bank*page+f.NativeOffset:], value)
	}
	return expected, nil
}

// Worker gate uses input bytes only, before shared signal arming/root dispatch.
// No original observations or expected result are passed to the child.
func tlCheckInputNativeBefore411350(in *tlInput411350, meta *tlMeta411350, actual []byte, nativePrefix int) error {
	pages, e := tlIndependentInputBefore411350(in)
	if e != nil {
		return e
	}
	expected, e := tlIndependentNativeBefore411350(in, pages, meta, nativePrefix)
	if e != nil {
		return e
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("every native BEFORE byte vs independent input projection")
	}
	return nil
}

// Parent additionally requires all genuine original BEFORE pages, including
// stack canaries/root frame, to equal the frozen independent literal input.
func tlCheckOriginalNativeBefore411350(in *tlInput411350, row *tlReference411350, meta *tlMeta411350, actual []byte, nativePrefix int) ([]byte, error) {
	canonical, e := tlIndependentInputBefore411350(in)
	if e != nil {
		return nil, e
	}
	if len(row.BeforePages) != len(canonical) {
		return nil, fmt.Errorf("all original BEFORE pages retained parent-only")
	}
	for address, wanted := range canonical {
		if !bytes.Equal(row.BeforePages[address], wanted) {
			return nil, fmt.Errorf("frozen original BEFORE/input page%08x", address)
		}
	}
	expected, e := tlIndependentNativeBefore411350(in, row.BeforePages, meta, nativePrefix)
	if e != nil {
		return nil, e
	}
	if !bytes.Equal(actual, expected) {
		return nil, fmt.Errorf("every native BEFORE byte vs authentic original projection")
	}
	return expected, nil
}
