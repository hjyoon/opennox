//go:build darwin && arm64

package legacy

// Only the independently chosen input is visible to a native worker. The
// parent opens original observations after Wait. Mapping and effects here
// transport fields and identities; they do not choose a branch or a result.
import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

type waWrite410390 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type waRegion410390 struct {
	Name       string `json:"name"`
	Base       uint32 `json:"base"`
	Accessible int    `json:"accessible_bytes"`
	Initial    string `json:"initial_hex"`
}
type waPacket410390 struct {
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []waWrite410390   `json:"writes"`
}
type waSchedule410390 struct {
	Target    uint32           `json:"target"`
	Argc      int              `json:"argc32"`
	Repeat    bool             `json:"repeat_last"`
	Responses []waPacket410390 `json:"responses"`
}
type waStimulus410390 struct {
	PC         uint32          `json:"pc"`
	Site       string          `json:"site"`
	Address    uint32          `json:"address"`
	Width      int             `json:"width"`
	Occurrence int             `json:"occurrence"`
	Writes     []waWrite410390 `json:"writes"`
}
type waInput410390 struct {
	Args          []uint32                    `json:"args_u32"`
	InitialWords  map[string]uint32           `json:"initial_words"`
	Name          string                      `json:"name"`
	Entry         uint32                      `json:"entry"`
	Roles         map[string]uint32           `json:"roles"`
	Regions       []waRegion410390            `json:"regions"`
	Services      map[string]waSchedule410390 `json:"services"`
	AfterReads    []waStimulus410390          `json:"after_reads"`
	AfterServices []waStimulus410390          `json:"after_services"`
}
type waField410390 struct {
	Region                                         int
	PE                                             uint32
	PEWidth, NativeOffset, NativeWidth, Accessible int
	Pointer, Signed                                bool
}
type waBank410390 struct {
	Name                 string
	PE                   uint32
	Prefix, AnchorOffset int
	Native               uint64
}
type waMeta410390 struct {
	Page       int
	Banks      []waBank410390
	Fields     []waField410390
	Identities map[string]uint64
	CLayout    [5]wallAttachFixtureField410390
	GoOffsets  [5]uintptr
	Opaque     []uint32
}
type waGraph410390 struct {
	Input                   *waInput410390
	File                    *os.File
	Header, Local           unsafe.Pointer
	Pointers, OpaqueMaps    []unsafe.Pointer
	Tokens                  map[uint32]unsafe.Pointer
	Meta                    waMeta410390
	Reads                   map[string]int
	Services                map[uint32]int
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
}

func waHex410390(s string) []byte {
	p, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return p
}
func waWord410390(p []byte) uint32 {
	if len(p) > 4 {
		panic("PE scalar width")
	}
	var out uint32
	for i, b := range p {
		out |= uint32(b) << uint(8*i)
	}
	return out
}
func waBytes410390(v uint64, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte(v >> uint(8*i))
	}
	return out
}
func waPointerSite410390(site uint32) bool {
	switch site {
	case 0x4103a0, 0x4103b0, 0x4103e3, 0x4103fa:
		return true
	}
	return false
}
func waRegionAt410390(in *waInput410390, pe uint32) int {
	for i, r := range in.Regions {
		if uint64(pe) >= uint64(r.Base) && uint64(pe) < uint64(r.Base)+uint64(r.Accessible) {
			return i
		}
	}
	return -1
}
func waAvailable410390(in *waInput410390, pe uint32, width int) int {
	i := waRegionAt410390(in, pe)
	if i < 0 {
		return 0
	}
	n := int(uint64(in.Regions[i].Base) + uint64(in.Regions[i].Accessible) - uint64(pe))
	if n > width {
		n = width
	}
	return n
}

func waBuildGraph410390(in *waInput410390, path string) (g *waGraph410390, err error) {
	g = &waGraph410390{Input: in, Tokens: map[uint32]unsafe.Pointer{}, Reads: map[string]int{}, Services: map[uint32]int{}, UsedResponses: map[string][]int{}}
	owner := g
	defer func() {
		if err != nil {
			owner.close()
		}
	}()
	page := os.Getpagesize()
	if page < 8192 || page&(page-1) != 0 {
		return nil, fmt.Errorf("native guard page geometry")
	}
	g.Meta = waMeta410390{Page: page, Identities: map[string]uint64{}, CLayout: wallAttachFixtureCLayout410390(), GoOffsets: wallAttachNativeOffsets410390()}
	for i, w := range []uintptr{unsafe.Sizeof(int(0)), unsafe.Sizeof(int(0)), 1, unsafe.Sizeof(unsafe.Pointer(nil)), 1} {
		if g.Meta.CLayout[i].Offset != g.Meta.GoOffsets[i] || g.Meta.CLayout[i].Width != w {
			return nil, fmt.Errorf("actual C/Go named layout %d", i)
		}
	}
	if wallAttachFixturePointerWidth410390() != unsafe.Sizeof(unsafe.Pointer(nil)) {
		return nil, fmt.Errorf("C/Go pointer width")
	}
	g.File, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	if err = g.File.Truncate(int64((len(in.Regions) + 1) * page)); err != nil {
		return nil, err
	}
	g.Header, err = rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), 0, page)
	if err != nil {
		return nil, err
	}
	ddiFill4E0B30(g.Header, page, 0)
	*(*uint64)(g.Header) = rawDefaultDamageSharedMagic4E0B30
	*(*uint64)(unsafe.Add(g.Header, 8)) = rawDefaultDamageSharedVersion4E0B30
	type descriptor struct {
		token             uint32
		pe, native, width int
		pointer, signed   bool
	}
	var ds []descriptor
	add := func(role string, pe int, native uintptr, width int, pointer, signed bool) {
		token := in.Roles[role]
		if token == 0 || waRegionAt410390(in, token) < 0 {
			return
		}
		d := descriptor{token, pe, int(native), width, pointer, signed}
		for _, old := range ds {
			if old.token == d.token && old.pe == d.pe {
				if old != d {
					panic("conflicting alias descriptor")
				}
				return
			}
		}
		ds = append(ds, d)
	}
	for role := range in.Roles {
		switch {
		case role == "drawable":
			add(role, 12, g.Meta.GoOffsets[0], 4, false, true)
			add(role, 16, g.Meta.GoOffsets[1], 4, false, true)
		case strings.HasPrefix(role, "wall"):
			add(role, 4, g.Meta.GoOffsets[2], 1, false, false)
			add(role, 8, g.Meta.GoOffsets[4], 1, false, false)
			add(role, 32, g.Meta.GoOffsets[3], 4, true, false)
		case strings.HasPrefix(role, "polygon"):
			add(role, 130, 130, 1, false, false)
		}
	}
	sort.Slice(ds, func(i, j int) bool {
		if ds[i].token == ds[j].token {
			return ds[i].pe < ds[j].pe
		}
		return ds[i].token < ds[j].token
	})
	for i, r := range in.Regions {
		initial := waHex410390(r.Initial)
		if r.Accessible != len(initial) || r.Accessible < 1 || r.Accessible > 4096 {
			return nil, fmt.Errorf("exact independent region prefix")
		}
		lower, upper := 0, page
		for _, d := range ds {
			if d.token != r.Base {
				continue
			}
			available := waAvailable410390(in, d.token+uint32(d.pe), d.width)
			nw := d.width
			if (d.pointer || d.signed) && available == d.width {
				nw = int(unsafe.Sizeof(unsafe.Pointer(nil)))
			}
			if available == d.width {
				if d.native+nw > lower {
					lower = d.native + nw
				}
			} else if available == 0 {
				if d.native < upper {
					upper = d.native
				}
			} else {
				point := d.native + available
				if point > lower {
					lower = point
				}
				if point < upper {
					upper = point
				}
			}
		}
		cut := upper
		if upper == page {
			cut = lower
		}
		if cut == 0 {
			cut = 1
		}
		if lower > upper || cut > page {
			return nil, fmt.Errorf("native prefix constraints")
		}
		p, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((i+1)*page), page)
		if e != nil {
			return nil, e
		}
		g.Pointers = append(g.Pointers, p)
		if uintptr(p) <= 0xffffffff {
			return nil, fmt.Errorf("high native foreign bank")
		}
		ddiFill4E0B30(p, page, byte(0xb0+i))
		b := waBank410390{r.Name, r.Base, r.Accessible, page - cut, uint64(uintptr(p))}
		g.Meta.Banks = append(g.Meta.Banks, b)
		g.Tokens[r.Base] = unsafe.Add(p, b.AnchorOffset)
	}
	// Opaque PE tokens get distinct protected native identities, never numeric
	// pointer reconstruction. No accessible body is invented for such tokens.
	all := map[uint32]bool{}
	for _, v := range in.Roles {
		all[v] = true
	}
	for _, d := range ds {
		if !d.pointer || waAvailable410390(in, d.token+uint32(d.pe), 4) != 4 {
			continue
		}
		ri := waRegionAt410390(in, d.token)
		r := in.Regions[ri]
		all[waWord410390(waHex410390(r.Initial)[int(d.token-r.Base)+d.pe:int(d.token-r.Base)+d.pe+4])] = true
	}
	for site, s := range in.Services {
		n, e := strconv.ParseUint(site, 16, 32)
		if e != nil {
			return nil, e
		}
		if waPointerSite410390(uint32(n)) {
			for _, r := range s.Responses {
				all[r.Registers["eax"]] = true
			}
		}
	}
	allWrites := []waWrite410390{}
	for _, s := range in.Services {
		for _, r := range s.Responses {
			allWrites = append(allWrites, r.Writes...)
		}
	}
	for _, s := range append(append([]waStimulus410390{}, in.AfterReads...), in.AfterServices...) {
		allWrites = append(allWrites, s.Writes...)
	}
	for _, w := range allWrites {
		for _, d := range ds {
			if d.pointer && w.Address == d.token+uint32(d.pe) && len(waHex410390(w.Hex)) == 4 {
				all[waWord410390(waHex410390(w.Hex))] = true
			}
		}
	}
	var tokens []uint32
	for t := range all {
		if t != 0 {
			tokens = append(tokens, t)
		}
	}
	sort.Slice(tokens, func(i, j int) bool { return tokens[i] < tokens[j] })
	for _, t := range tokens {
		if _, ok := g.Tokens[t]; ok {
			continue
		}
		if waRegionAt410390(in, t) >= 0 {
			return nil, fmt.Errorf("undeclared readable interior pointer identity")
		}
		p := ddiMap4E0B30(page)
		if p == nil {
			return nil, fmt.Errorf("opaque guard allocation")
		}
		g.OpaqueMaps = append(g.OpaqueMaps, p)
		g.Tokens[t] = unsafe.Add(p, page)
		g.Meta.Opaque = append(g.Meta.Opaque, t)
	}
	seen := map[unsafe.Pointer]bool{}
	for t, p := range g.Tokens {
		if p == nil || uintptr(p) <= 0xffffffff || seen[p] {
			return nil, fmt.Errorf("full native unique canonical identity")
		}
		w := twoSettingsPointerFixturePointerWitness40A740(p)
		if w.Native != uintptr(p) || w.Low32 != uint32(uintptr(p)) || !w.Above4GiB || !w.Low32LosesBits {
			return nil, fmt.Errorf("actual C full-pointer and numeric low32-loss witness")
		}
		seen[p] = true
		g.Meta.Identities[strconv.FormatUint(uint64(t), 10)] = uint64(uintptr(p))
	}
	for _, d := range ds {
		ri := waRegionAt410390(in, d.token)
		b := g.Meta.Banks[ri]
		pe := d.token + uint32(d.pe)
		available := waAvailable410390(in, pe, d.width)
		nw := d.width
		if d.pointer || d.signed {
			nw = int(unsafe.Sizeof(unsafe.Pointer(nil)))
		}
		f := waField410390{ri, pe, d.width, b.AnchorOffset + d.native, nw, available, d.pointer, d.signed}
		g.Meta.Fields = append(g.Meta.Fields, f)
		if available == 0 {
			continue
		}
		r := in.Regions[ri]
		data := append([]byte(nil), waHex410390(r.Initial)[int(pe-r.Base):int(pe-r.Base)+available]...)
		if d.pointer && available == 4 {
			p, e := g.pointer(waWord410390(data))
			if e != nil {
				return nil, e
			}
			if f.NativeOffset < 0 || f.NativeOffset+nw > page {
				return nil, fmt.Errorf("complete native pointer slot extent")
			}
			dst := unsafe.Add(g.Pointers[ri], f.NativeOffset)
			twoSettingsPointerFixtureStorePointer40A740(dst, p)
			if twoSettingsPointerFixtureLoadPointer40A740(dst) != p {
				return nil, fmt.Errorf("C memcpy full-pointer transport")
			}
			continue
		}
		if d.signed && available == 4 {
			data = waBytes410390(uint64(int64(int32(waWord410390(data)))), 8)
		}
		if f.NativeOffset < 0 || f.NativeOffset+len(data) > page {
			return nil, fmt.Errorf("partial native field transport bounds")
		}
		ddiCopyIn4E0B30(unsafe.Add(g.Pointers[ri], f.NativeOffset), data)
	}
	for _, p := range append([]unsafe.Pointer{g.Header}, g.Pointers...) {
		if e := rawDefaultDamageSharedSync4E0B30(p, page); e != nil {
			return nil, e
		}
	}
	return g, nil
}
func (g *waGraph410390) pointer(t uint32) (unsafe.Pointer, error) {
	if t == 0 {
		return nil, nil
	}
	p, ok := g.Tokens[t]
	if !ok {
		return nil, fmt.Errorf("undeclared PE token %08x", t)
	}
	return p, nil
}
func (g *waGraph410390) address(pe uint32, width int) (unsafe.Pointer, *waField410390, error) {
	for _, f := range g.Meta.Fields {
		if pe < f.PE {
			continue
		}
		delta := uint64(pe) - uint64(f.PE)
		if width > 0 && delta+uint64(width) <= uint64(f.PEWidth) && (!(f.Pointer || f.Signed) || (delta == 0 && width == 4)) {
			if delta+uint64(width) > uint64(f.Accessible) {
				return nil, nil, fmt.Errorf("input effect must be fully mapped")
			}
			f.NativeOffset += int(delta)
			f.PE = pe
			f.PEWidth = width
			f.Accessible = width
			if !f.Pointer && !f.Signed {
				f.NativeWidth = width
			}
			return unsafe.Add(g.Pointers[f.Region], f.NativeOffset), &f, nil
		}
	}
	return nil, nil, fmt.Errorf("unprojected input effect %08x/%d", pe, width)
}
func (g *waGraph410390) apply(writes []waWrite410390) error {
	for _, w := range writes {
		data := waHex410390(w.Hex)
		if w.Address >= 0x220007f8 && uint64(w.Address)+uint64(len(data)) <= 0x22000800 {
			if g.Local == nil {
				return fmt.Errorf("local effect before actual native local exists")
			}
			ddiCopyIn4E0B30(unsafe.Add(g.Local, uintptr(w.Address-0x220007f8)), data)
			continue
		}
		p, f, e := g.address(w.Address, len(data))
		if e != nil {
			return e
		}
		if f.Pointer {
			v, e := g.pointer(waWord410390(data))
			if e != nil {
				return e
			}
			twoSettingsPointerFixtureStorePointer40A740(p, v)
		} else {
			if f.Signed && len(data) == 4 {
				data = waBytes410390(uint64(int64(int32(waWord410390(data)))), 8)
			}
			ddiCopyIn4E0B30(p, data)
		}
	}
	return nil
}
func (g *waGraph410390) afterRead(site, pe uint32, width int) error {
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
func (g *waGraph410390) afterService(site uint32) error {
	g.Services[site]++
	for i, s := range g.Input.AfterServices {
		if s.Site == fmt.Sprintf("%08x", site) && s.Occurrence == g.Services[site] {
			if e := g.apply(s.Writes); e != nil {
				return e
			}
			g.UsedServices = append(g.UsedServices, i)
		}
	}
	return nil
}
func (g *waGraph410390) snapshot() []byte {
	var out []byte
	for _, p := range g.Pointers {
		out = append(out, ddiSnapshot4E0B30(p, g.Meta.Page)...)
	}
	return out
}
func (g *waGraph410390) close() {
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

// Pull original bytes from actual full-page AFTER records, including accesses
// whose semantic operand ends at a guard. The caller retains authentic raw
// fault addresses separately; this projection never synthesizes a read.
func waOriginalBytes410390(pages map[uint32][]byte, address uint32, n int) ([]byte, error) {
	out := make([]byte, n)
	for i := range out {
		a := address + uint32(i)
		page := a &^ uint32(4095)
		p, ok := pages[page]
		if !ok || len(p) != 4096 {
			return nil, fmt.Errorf("original AFTER page %08x", page)
		}
		out[i] = p[a-page]
	}
	return out, nil
}
func waExpectedAfter410390(meta *waMeta410390, before []byte, fields map[uint32][]byte) ([]byte, error) {
	if len(before) != len(meta.Banks)*meta.Page {
		return nil, fmt.Errorf("whole native BEFORE extent")
	}
	out := append([]byte(nil), before...)
	for _, f := range meta.Fields {
		if f.Accessible == 0 {
			continue
		}
		data, ok := fields[f.PE]
		if !ok || len(data) != f.Accessible {
			return nil, fmt.Errorf("authentic original AFTER field span")
		}
		data = append([]byte(nil), data...)
		if f.Pointer && f.Accessible == 4 {
			t := waWord410390(data)
			var p uint64
			if t != 0 {
				var ok bool
				p, ok = meta.Identities[strconv.FormatUint(uint64(t), 10)]
				if !ok {
					return nil, fmt.Errorf("original pointer AFTER identity")
				}
			}
			data = waBytes410390(p, f.NativeWidth)
		}
		if f.Signed && f.Accessible == 4 {
			data = waBytes410390(uint64(int64(int32(waWord410390(data)))), 8)
		}
		offset := f.Region*meta.Page + f.NativeOffset
		if f.NativeOffset < 0 || f.NativeOffset+len(data) > meta.Page {
			return nil, fmt.Errorf("native AFTER field span")
		}
		copy(out[offset:], data)
	}
	return out, nil
}

// All observer identities and EA/width checks derive from the actual native
// operands and fixed original instruction descriptors, never the next oracle row.
func waOperand410390(g *waGraph410390, e wallAttachNativeOperand410390) (uint32, int, error) {
	type desc struct {
		field  int
		offset uint32
		width  int
		write  bool
	}
	ds := map[uint32]desc{0x4103c4: {0, 12, 4, false}, 0x4103c7: {1, 16, 4, false}, 0x4103ca: {2, 0, 4, true}, 0x4103ce: {3, 4, 1, false}, 0x4103d1: {4, 32, 4, true}, 0x4103dc: {5, 4, 4, true}, 0x4103e0: {3, 4, 1, true}, 0x410406: {6, 130, 1, false}, 0x41040d: {7, 8, 1, true}, 0x410416: {7, 8, 1, true}}
	d, ok := ds[e.Site]
	if !ok || e.Field != d.field || e.Write != d.write || e.IsPointer != (d.field == 4) {
		return 0, 0, fmt.Errorf("static native operand site/field/kind")
	}
	offsets := []uintptr{g.Meta.GoOffsets[0], g.Meta.GoOffsets[1], 0, g.Meta.GoOffsets[2], g.Meta.GoOffsets[3], 4, 130, g.Meta.GoOffsets[4]}
	nw := d.width
	if d.field == 0 || d.field == 1 || d.field == 4 {
		nw = int(unsafe.Sizeof(uintptr(0)))
	}
	if e.Offset != offsets[d.field] || int(e.Width) != nw || (!e.Write && e.ValueValid != e.Complete) || (e.Write && !e.ValueValid) {
		return 0, 0, fmt.Errorf("actual width/offset/value-validity")
	}
	if d.field == 2 || d.field == 5 {
		if e.Base == nil || !e.AddressValid || e.Address != unsafe.Add(e.Base, e.Offset) {
			return 0, 0, fmt.Errorf("actual private local operand")
		}
		if g.Local == nil {
			g.Local = e.Base
		}
		if g.Local != e.Base {
			return 0, 0, fmt.Errorf("one native local8 identity")
		}
		return 0x220007f8 + d.offset, d.width, nil
	}
	token := uint32(0)
	found := e.Base == nil
	for t, p := range g.Tokens {
		if p == e.Base {
			if found {
				return 0, 0, fmt.Errorf("pointer identity collision")
			}
			token = t
			found = true
		}
	}
	if !found {
		return 0, 0, fmt.Errorf("undeclared actual native base")
	}
	if e.Base == nil {
		if !e.AddressValid || e.Address != nil {
			return 0, 0, fmt.Errorf("known native nil-load0; logical original EA remains separate")
		}
	} else if !e.AddressValid || e.Address != unsafe.Add(e.Base, e.Offset) {
		return 0, 0, fmt.Errorf("actual native effective operand")
	}
	return token + d.offset, d.width, nil
}
