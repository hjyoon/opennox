//go:build darwin && arm64

package legacy

// Independent-input transport only: no reference rows or branch/return model.
import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"unsafe"
)

type tlWrite411350 = swWrite410550
type tlRegion411350 = swRegion410550
type tlStimulus411350 = swStimulus410550
type tlEdgePage411350 struct {
	Page   uint32 `json:"page"`
	Mapped bool   `json:"mapped"`
	Seed   byte   `json:"byte_seed"`
}
type tlInput411350 struct {
	Name             string             `json:"name"`
	Family           string             `json:"family"`
	Entry            uint32             `json:"entry"`
	Args             []uint32           `json:"args_u32"`
	Roles            map[string]uint32  `json:"roles"`
	Regions          []tlRegion411350   `json:"regions"`
	EdgePages        []tlEdgePage411350 `json:"edge_pages"`
	InitialWords     map[string]uint32  `json:"initial_words"`
	InitialWrites    []tlWrite411350    `json:"initial_writes"`
	AfterReads       []tlStimulus411350 `json:"after_reads"`
	AfterReturns     []tlStimulus411350 `json:"after_internal_returns"`
	InitialRegisters map[string]uint32  `json:"initial_registers_u32"`
	InitialEflags    uint32             `json:"initial_eflags"`
	StackSeed        byte               `json:"stack_byte_seed"`
}
type tlField411350 struct {
	Bank                                           int
	PE                                             uint32
	PEWidth, NativeOffset, NativeWidth, Accessible int
	Pointer                                        bool
}
type tlBank411350 struct {
	Name, Kind                   string
	PE                           uint32
	Prefix, NativePrefix, Anchor int
	Native                       uint64
	Mapped                       bool
}
type tlMeta411350 struct {
	Page                 int
	Banks                []tlBank411350
	Fields               []tlField411350
	Identities           map[string]uint64
	CLayout              [7]uintptr
	GoOffsets            [5]uintptr
	Opaque               []uint32
	NativeOnlyNextPrefix int
	EdgeFault            uint64
}
type tlGraph411350 struct {
	Input                  *tlInput411350
	File                   *os.File
	Header                 unsafe.Pointer
	Pointers, OpaqueMaps   []unsafe.Pointer
	Tokens                 map[uint32]unsafe.Pointer
	EdgeFault              unsafe.Pointer
	Meta                   tlMeta411350
	Reads                  map[string]int
	Returns                map[uint32]int
	UsedReads, UsedReturns []int
}

func tlPattern411350(seed byte) []byte {
	b := make([]byte, 4096)
	for i := range b {
		b[i] = byte(int(seed) + 29*i + 7*(i>>4))
	}
	return b
}
func tlInitialPages411350(in *tlInput411350) (map[uint32][]byte, error) {
	if len(in.Regions) != 4 || len(in.EdgePages) != 2 || len(in.InitialWords) != 0 {
		return nil, fmt.Errorf("literal graph geometry")
	}
	pages := map[uint32][]byte{}
	for _, r := range in.Regions {
		limit := 20
		if r.Kind == "xy" {
			limit = 8
		} else if r.Kind != "node" {
			return nil, fmt.Errorf("input kind")
		}
		if r.Page%4096 != 0 || r.Guard != r.Page+4096 || r.Prefix < 0 || r.Prefix > limit || pages[r.Page] != nil {
			return nil, fmt.Errorf("input prefix/page")
		}
		pages[r.Page] = tlPattern411350(r.Seed)
	}
	seen := map[uint32]bool{}
	for _, e := range in.EdgePages {
		if (e.Page != 0x862000 && e.Page != 0x863000) || seen[e.Page] {
			return nil, fmt.Errorf("EDGE pages")
		}
		seen[e.Page] = true
		if e.Mapped {
			pages[e.Page] = tlPattern411350(e.Seed)
		}
	}
	for _, w := range in.InitialWrites {
		b := swHex410550(w.Hex)
		page := w.Address &^ uint32(4095)
		off := int(w.Address - page)
		if len(b) < 1 || len(b) > 4 || pages[page] == nil || off+len(b) > 4096 {
			return nil, fmt.Errorf("literal initializer extent")
		}
		copy(pages[page][off:], b)
	}
	return pages, nil
}
func tlBuildGraph411350(in *tlInput411350, path string) (g *tlGraph411350, err error) {
	g = &tlGraph411350{Input: in, Tokens: map[uint32]unsafe.Pointer{}, Reads: map[string]int{}, Returns: map[uint32]int{}}
	owner := g
	defer func() {
		if err != nil {
			owner.close()
		}
	}()
	page := os.Getpagesize()
	if page < 8192 || page&(page-1) != 0 {
		return nil, fmt.Errorf("native guard page")
	}
	initial, e := tlInitialPages411350(in)
	if e != nil {
		return nil, e
	}
	var n tileListNativeNode411350
	g.Meta = tlMeta411350{Page: page, Identities: map[string]uint64{}, CLayout: tileListFixtureLayout411350(), GoOffsets: [5]uintptr{unsafe.Offsetof(n.Value), unsafe.Offsetof(n.Preserved), unsafe.Offsetof(n.Edge), unsafe.Offsetof(n.Variation), unsafe.Offsetof(n.Next)}}
	for i, v := range g.Meta.GoOffsets {
		if g.Meta.CLayout[i] != v {
			return nil, fmt.Errorf("actual node offset%d", i)
		}
	}
	pw := int(unsafe.Sizeof(unsafe.Pointer(nil)))
	if pw != 8 || g.Meta.CLayout[5] != unsafe.Sizeof(n) || g.Meta.CLayout[5] != 24 || g.Meta.CLayout[6] != uintptr(pw) {
		return nil, fmt.Errorf("actual native node/PTR width")
	}
	g.File, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	if err = g.File.Truncate(int64(7 * page)); err != nil {
		return nil, err
	}
	g.Header, err = rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), 0, page)
	if err != nil {
		return nil, err
	}
	ddiFill4E0B30(g.Header, page, 0)
	*(*uint64)(g.Header) = rawDefaultDamageSharedMagic4E0B30
	*(*uint64)(unsafe.Add(g.Header, 8)) = rawDefaultDamageSharedVersion4E0B30
	addBank := func(name, kind string, pe uint32, prefix, cut int, mapped bool) error {
		bank, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((len(g.Pointers)+1)*page), page)
		if e != nil {
			return e
		}
		g.Pointers = append(g.Pointers, bank)
		ddiFill4E0B30(bank, page, byte(0xa0+len(g.Pointers)))
		if uintptr(bank) <= 0xffffffff {
			return fmt.Errorf("actual high native C bank")
		}
		g.Meta.Banks = append(g.Meta.Banks, tlBank411350{name, kind, pe, prefix, cut, page - cut, uint64(uintptr(bank)), mapped})
		return nil
	}
	for _, r := range in.Regions {
		cut := r.Prefix
		if r.Kind == "node" && cut == 20 {
			cut = 24
		}
		if err = addBank(r.Name, r.Kind, r.Guard-uint32(r.Prefix), r.Prefix, cut, true); err != nil {
			return nil, err
		}
		b := g.Meta.Banks[len(g.Meta.Banks)-1]
		if _, exists := g.Tokens[b.PE]; exists {
			return nil, fmt.Errorf("physical base identity")
		}
		g.Tokens[b.PE] = unsafe.Add(g.Pointers[len(g.Pointers)-1], b.Anchor)
	}
	for _, ep := range in.EdgePages {
		if err = addBank(fmt.Sprintf("edge%08x", ep.Page), "edge", ep.Page, 4096, 4096, ep.Mapped); err != nil {
			return nil, err
		}
		i := len(g.Pointers) - 1
		b := g.Meta.Banks[i]
		if ep.Mapped {
			ddiCopyIn4E0B30(unsafe.Add(g.Pointers[i], b.Anchor), initial[ep.Page])
			g.Meta.Fields = append(g.Meta.Fields, tlField411350{i, ep.Page, 4096, b.Anchor, 4096, 4096, false})
		}
	}
	for i, r := range in.Regions {
		b := g.Meta.Banks[i]
		widths := []int{4, 4}
		offsets := []int{0, 4}
		if r.Kind == "node" {
			widths = []int{4, 4, 4, 4, 4}
			offsets = []int{0, 4, 8, 12, 16}
		}
		for j, off := range offsets {
			accessible := r.Prefix - off
			if accessible < 0 {
				accessible = 0
			}
			if accessible > widths[j] {
				accessible = widths[j]
			}
			pointer := r.Kind == "node" && off == 16
			nw := widths[j]
			if pointer {
				nw = pw
			}
			g.Meta.Fields = append(g.Meta.Fields, tlField411350{i, b.PE + uint32(off), widths[j], b.Anchor + off, nw, accessible, pointer})
		}
	}
	all := map[uint32]bool{}
	for _, token := range in.Roles {
		all[token] = true
	}
	for _, f := range g.Meta.Fields {
		if f.Pointer && f.Accessible == 4 {
			data, e := swOriginalBytes410550(initial, f.PE, 4)
			if e != nil {
				return nil, e
			}
			all[swWord410550(data)] = true
		}
	}
	for _, s := range append(append([]tlStimulus411350{}, in.AfterReads...), in.AfterReturns...) {
		for _, w := range s.Writes {
			for _, f := range g.Meta.Fields {
				if f.Pointer && f.PE == w.Address && len(swHex410550(w.Hex)) == 4 {
					all[swWord410550(swHex410550(w.Hex))] = true
				}
			}
		}
	}
	for _, token := range in.Args {
		if token == in.Roles["nil"] || token == in.Roles["unmapped"] || token == in.Roles["high_unmapped"] || token == in.Roles["max_unmapped"] {
			all[token] = true
		}
	}
	var tokens []uint32
	for token := range all {
		if token != 0 {
			tokens = append(tokens, token)
		}
	}
	sort.Slice(tokens, func(i, j int) bool { return tokens[i] < tokens[j] })
	for _, token := range tokens {
		if _, ok := g.Tokens[token]; !ok {
			p := ddiMap4E0B30(page)
			if p == nil {
				return nil, fmt.Errorf("opaque guard")
			}
			g.OpaqueMaps = append(g.OpaqueMaps, p)
			g.Tokens[token] = unsafe.Add(p, page)
			g.Meta.Opaque = append(g.Meta.Opaque, token)
		}
	}
	for token, p := range g.Tokens {
		w := twoSettingsPointerFixturePointerWitness40A740(p)
		if uintptr(p) <= 0xffffffff || w.Native != uintptr(p) || !w.Above4GiB || !w.Low32LosesBits || w.Low32 != uint32(uintptr(p)) {
			return nil, fmt.Errorf("C high pointer witness")
		}
		g.Meta.Identities[strconv.FormatUint(uint64(token), 10)] = uint64(uintptr(p))
	}
	p := ddiMap4E0B30(page)
	if p == nil {
		return nil, fmt.Errorf("EDGE fault carrier")
	}
	g.OpaqueMaps = append(g.OpaqueMaps, p)
	g.EdgeFault = unsafe.Add(p, page)
	g.Meta.EdgeFault = uint64(uintptr(g.EdgeFault))
	for _, f := range g.Meta.Fields {
		if f.Accessible == 0 || g.Meta.Banks[f.Bank].Kind == "edge" {
			continue
		}
		data, e := swOriginalBytes410550(initial, f.PE, f.Accessible)
		if e != nil {
			return nil, e
		}
		slot := unsafe.Add(g.Pointers[f.Bank], f.NativeOffset)
		if f.Pointer && f.Accessible == 4 {
			v, e := g.pointer(swWord410550(data))
			if e != nil {
				return nil, e
			}
			tileListFixtureStorePointer411350(slot, v)
			if tileListFixtureLoadPointer411350(slot) != v {
				return nil, fmt.Errorf("full PTR initialization")
			}
		} else {
			ddiCopyIn4E0B30(slot, data)
		}
	}
	for _, p := range append([]unsafe.Pointer{g.Header}, g.Pointers...) {
		if err = rawDefaultDamageSharedSync4E0B30(p, page); err != nil {
			return nil, err
		}
	}
	return g, nil
}
func (g *tlGraph411350) pointer(token uint32) (unsafe.Pointer, error) {
	if token == 0 {
		return nil, nil
	}
	p, ok := g.Tokens[token]
	if !ok {
		return nil, fmt.Errorf("undeclared pointer identity%08x", token)
	}
	return p, nil
}

// Separate native-only discriminator. Original completePE20 normally maps24.
// Copy only already-initialized ACTUAL native bytes, never an original result
// or a completed pointer synthesized from a fault. Selected controls have no
// mutations and a literal nilNext, so no relocated incoming pointer needs fixup.
func (g *tlGraph411350) nativeOnlyNextPrefix(prefix int) error {
	if prefix < 20 || prefix > 24 || g.Input.Entry != 0x411350 || len(g.Input.AfterReads) != 0 || len(g.Input.AfterReturns) != 0 {
		return fmt.Errorf("native-only prefix contract")
	}
	i := -1
	token := g.Input.Roles["node0"]
	for index, b := range g.Meta.Banks {
		if b.Kind == "node" && b.PE == token {
			i = index
			break
		}
	}
	if i < 0 || g.Meta.Banks[i].Prefix != 20 || g.Input.Args[0] != token {
		return fmt.Errorf("complete PE node0 input")
	}
	initial, e := tlInitialPages411350(g.Input)
	if e != nil {
		return e
	}
	for _, r := range g.Input.Regions {
		if r.Kind == "node" {
			base := r.Guard - uint32(r.Prefix)
			if r.Prefix != 20 {
				return fmt.Errorf("native-only complete input nodes")
			}
			b, e := swOriginalBytes410550(initial, base+16, 4)
			if e != nil || swWord410550(b) != 0 {
				return fmt.Errorf("native-only literal nil Next")
			}
		}
	}
	b := &g.Meta.Banks[i]
	raw := ddiSnapshot4E0B30(g.Pointers[i], g.Meta.Page)
	data := append([]byte(nil), raw[b.Anchor:b.Anchor+prefix]...)
	oldAnchor := b.Anchor
	b.NativePrefix = prefix
	b.Anchor = g.Meta.Page - prefix
	delta := b.Anchor - oldAnchor
	ddiFill4E0B30(g.Pointers[i], g.Meta.Page, byte(0xa1+i))
	ddiCopyIn4E0B30(unsafe.Add(g.Pointers[i], b.Anchor), data)
	for index := range g.Meta.Fields {
		if g.Meta.Fields[index].Bank == i {
			g.Meta.Fields[index].NativeOffset += delta
		}
	}
	g.Tokens[token] = unsafe.Add(g.Pointers[i], b.Anchor)
	g.Meta.Identities[strconv.FormatUint(uint64(token), 10)] = uint64(uintptr(g.Tokens[token]))
	g.Meta.NativeOnlyNextPrefix = prefix
	w := twoSettingsPointerFixturePointerWitness40A740(g.Tokens[token])
	if !w.Above4GiB || w.Native != uintptr(g.Tokens[token]) || !w.Low32LosesBits {
		return fmt.Errorf("native-only pointer witness")
	}
	return rawDefaultDamageSharedSync4E0B30(g.Pointers[i], g.Meta.Page)
}
func (g *tlGraph411350) edgeAddress(site, index, offset, logical uint32) unsafe.Pointer {
	if logical != uint32(0x85b3fc)+offset {
		panic("EDGE logical metadata")
	}
	for i, b := range g.Meta.Banks {
		if b.Kind == "edge" && logical >= b.PE && uint64(logical) < uint64(b.PE)+4096 {
			if b.Mapped {
				return unsafe.Add(g.Pointers[i], b.Anchor+int(logical-b.PE))
			}
			return unsafe.Add(g.Pointers[i], g.Meta.Page+int(logical-b.PE))
		}
	}
	return g.EdgeFault // actual guard carrier; physical originalEA equivalence false
}
func (g *tlGraph411350) apply(writes []tlWrite411350) error {
	for _, w := range writes {
		data := swHex410550(w.Hex)
		found := false
		for _, f := range g.Meta.Fields {
			if w.Address < f.PE {
				continue
			}
			delta := uint64(w.Address) - uint64(f.PE)
			if delta+uint64(len(data)) > uint64(f.Accessible) {
				continue
			}
			if f.Pointer && (delta != 0 || len(data) != 4) {
				return fmt.Errorf("partial PTR mutation forbidden")
			}
			slot := unsafe.Add(g.Pointers[f.Bank], f.NativeOffset+int(delta))
			if f.Pointer {
				p, e := g.pointer(swWord410550(data))
				if e != nil {
					return e
				}
				tileListFixtureStorePointer411350(slot, p)
			} else {
				ddiCopyIn4E0B30(slot, data)
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("unprojected memory-only effect%08x/%d", w.Address, len(data))
		}
	}
	return nil
}
func (g *tlGraph411350) afterRead(site, pe uint32, width int) error {
	key := fmt.Sprintf("%08x/%08x/%d", site, pe, width)
	g.Reads[key]++
	for i, s := range g.Input.AfterReads {
		if s.PC == site && s.Address == pe && s.Width == width && s.Occurrence == g.Reads[key] {
			if e := g.apply(s.Writes); e != nil {
				return e
			}
			g.UsedReads = append(g.UsedReads, i)
		}
	}
	return nil
}
func (g *tlGraph411350) afterReturn(site uint32) error {
	g.Returns[site]++
	for i, s := range g.Input.AfterReturns {
		if s.Site == fmt.Sprintf("%08x", site) && s.Occurrence == g.Returns[site] {
			if e := g.apply(s.Writes); e != nil {
				return e
			}
			g.UsedReturns = append(g.UsedReturns, i)
		}
	}
	return nil
}
func (g *tlGraph411350) snapshot() []byte {
	var out []byte
	for _, p := range g.Pointers {
		out = append(out, ddiSnapshot4E0B30(p, g.Meta.Page)...)
	}
	return out
}
func (g *tlGraph411350) close() {
	if g == nil {
		return
	}
	for _, p := range g.Pointers {
		rawDefaultDamageSharedUnmap4E0B30(p, g.Meta.Page)
	}
	if g.Header != nil {
		rawDefaultDamageSharedUnmap4E0B30(g.Header, g.Meta.Page)
	}
	for _, p := range g.OpaqueMaps {
		ddiUnmap4E0B30(p, g.Meta.Page)
	}
	if g.File != nil {
		g.File.Close()
	}
	runtime.KeepAlive(g)
}
