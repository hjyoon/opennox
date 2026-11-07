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

	"github.com/opennox/opennox/v1/server"
)

type tspWrite40A740 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type tspRegion40A740 struct {
	Name       string `json:"name"`
	Base       uint32 `json:"base"`
	Accessible int    `json:"accessible_bytes"`
	Initial    string `json:"initial_hex"`
}
type tspPacket40A740 struct {
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []tspWrite40A740  `json:"writes"`
}
type tspSchedule40A740 struct {
	Target    uint32            `json:"target"`
	Argc      int               `json:"argc32"`
	Repeat    bool              `json:"repeat_last"`
	Responses []tspPacket40A740 `json:"responses"`
}
type tspStimulus40A740 struct {
	PC         uint32           `json:"pc"`
	Site       string           `json:"site"`
	Address    uint32           `json:"address"`
	Width      int              `json:"width"`
	Occurrence int              `json:"occurrence"`
	Writes     []tspWrite40A740 `json:"writes"`
}
type tspInput40A740 struct {
	Name    string            `json:"name"`
	Entry   uint32            `json:"entry"`
	Roles   map[string]uint32 `json:"roles"`
	Regions []tspRegion40A740 `json:"regions"`
	Engine  struct {
		Mapped bool   `json:"mapped"`
		Value  uint32 `json:"value"`
	} `json:"engine_flags"`
	Services      map[string]tspSchedule40A740 `json:"services"`
	AfterReads    []tspStimulus40A740          `json:"after_reads"`
	AfterServices []tspStimulus40A740          `json:"after_services"`
}
type tspField40A740 struct {
	Region                                         int
	PE                                             uint32
	PEWidth, NativeOffset, NativeWidth, Accessible int
	Pointer                                        bool
}
type tspBank40A740 struct {
	Name                 string
	PE                   uint32
	Prefix, AnchorOffset int
	Native               uint64
}
type tspMeta40A740 struct {
	Page       int
	Banks      []tspBank40A740
	Fields     []tspField40A740
	Identities map[string]uint64
	CLayout    [4]twoSettingsPointerFixtureField40A740
	GoOffsets  [4]uintptr
	Opaque     []uint32
	EngineCell uint64
}
type tspGraph40A740 struct {
	Input                   *tspInput40A740
	File                    *os.File
	Header                  unsafe.Pointer
	Pointers, OpaqueMaps    []unsafe.Pointer
	Tokens                  map[uint32]unsafe.Pointer
	Meta                    tspMeta40A740
	Reads                   map[string]int
	Services                map[uint32]int
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
}

func tspHex40A740(s string) []byte {
	p, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return p
}
func tspWord40A740(p []byte) uint32 {
	if len(p) > 4 {
		panic("PE scalar width")
	}
	var out uint32
	for i, b := range p {
		out |= uint32(b) << uint(8*i)
	}
	return out
}
func tspBytes40A740(v uint64, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte(v >> uint(8*i))
	}
	return out
}
func tspPointerSite40A740(site uint32) bool {
	switch site {
	case 0x40a741, 0x40a784, 0x40a79a, 0x40a7c9, 0x40a7d9, 0x40a7ee, 0x40a819:
		return true
	}
	return false
}
func tspRegionAt40A740(in *tspInput40A740, pe uint32) int {
	for i, r := range in.Regions {
		if uint64(pe) >= uint64(r.Base) && uint64(pe) < uint64(r.Base)+uint64(r.Accessible) {
			return i
		}
	}
	return -1
}
func tspAvailable40A740(in *tspInput40A740, pe uint32, width int) int {
	i := tspRegionAt40A740(in, pe)
	if i < 0 {
		return 0
	}
	n := int(uint64(in.Regions[i].Base) + uint64(in.Regions[i].Accessible) - uint64(pe))
	if n > width {
		n = width
	}
	return n
}

func tspBuildGraph40A740(in *tspInput40A740, path string) (g *tspGraph40A740, err error) {
	g = &tspGraph40A740{Input: in, Tokens: map[uint32]unsafe.Pointer{}, Reads: map[string]int{}, Services: map[uint32]int{}, UsedResponses: map[string][]int{}}
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
	g.Meta = tspMeta40A740{Page: page, Identities: map[string]uint64{}, CLayout: twoSettingsPointerFixtureCLayout40A740(), GoOffsets: [4]uintptr{unsafe.Offsetof(server.Object{}.UpdateData), unsafe.Offsetof(server.PlayerUpdateData{}.Player), unsafe.Offsetof(server.Player{}.Field3680), unsafe.Offsetof(server.Player{}.PlayerInd)}}
	for i, w := range []uintptr{unsafe.Sizeof(unsafe.Pointer(nil)), unsafe.Sizeof(unsafe.Pointer(nil)), 4, 1} {
		if g.Meta.CLayout[i].Offset != g.Meta.GoOffsets[i] || g.Meta.CLayout[i].Width != w {
			return nil, fmt.Errorf("actual named C/Go field layout %d", i)
		}
	}
	if twoSettingsPointerFixtureCPointerWidth40A740() != unsafe.Sizeof(unsafe.Pointer(nil)) {
		return nil, fmt.Errorf("C/Go pointer width")
	}
	g.File, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	if err = g.File.Truncate(int64((len(in.Regions) + 2) * page)); err != nil {
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
		pointer           bool
	}
	var ds []descriptor
	add := func(role string, pe int, native uintptr, width int, pointer bool) {
		token := in.Roles[role]
		if token == 0 || tspRegionAt40A740(in, token) < 0 {
			return
		}
		d := descriptor{token, pe, int(native), width, pointer}
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
		case strings.HasPrefix(role, "info"):
			add(role, 53, 53, 1, false)
		case strings.HasPrefix(role, "unit"):
			add(role, 748, g.Meta.GoOffsets[0], 4, true)
		case strings.HasPrefix(role, "update"):
			add(role, 276, g.Meta.GoOffsets[1], 4, true)
		case strings.HasPrefix(role, "player"):
			add(role, 3680, g.Meta.GoOffsets[2], 4, false)
			add(role, 2064, g.Meta.GoOffsets[3], 1, false)
		}
	}
	sort.Slice(ds, func(i, j int) bool {
		if ds[i].token == ds[j].token {
			return ds[i].pe < ds[j].pe
		}
		return ds[i].token < ds[j].token
	})
	for i, r := range in.Regions {
		initial := tspHex40A740(r.Initial)
		if r.Accessible != len(initial) || r.Accessible < 1 || r.Accessible > 4096 {
			return nil, fmt.Errorf("exact independent region prefix")
		}
		lower, upper := 0, page
		for _, d := range ds {
			if d.token != r.Base {
				continue
			}
			available := tspAvailable40A740(in, d.token+uint32(d.pe), d.width)
			nw := d.width
			if d.pointer && available == d.width {
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
		b := tspBank40A740{r.Name, r.Base, r.Accessible, page - cut, uint64(uintptr(p))}
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
		if !d.pointer || tspAvailable40A740(in, d.token+uint32(d.pe), 4) != 4 {
			continue
		}
		ri := tspRegionAt40A740(in, d.token)
		r := in.Regions[ri]
		all[tspWord40A740(tspHex40A740(r.Initial)[int(d.token-r.Base)+d.pe:int(d.token-r.Base)+d.pe+4])] = true
	}
	for site, s := range in.Services {
		n, e := strconv.ParseUint(site, 16, 32)
		if e != nil {
			return nil, e
		}
		if tspPointerSite40A740(uint32(n)) {
			for _, r := range s.Responses {
				all[r.Registers["eax"]] = true
			}
		}
	}
	allWrites := []tspWrite40A740{}
	for _, s := range in.Services {
		for _, r := range s.Responses {
			allWrites = append(allWrites, r.Writes...)
		}
	}
	for _, s := range append(append([]tspStimulus40A740{}, in.AfterReads...), in.AfterServices...) {
		allWrites = append(allWrites, s.Writes...)
	}
	for _, w := range allWrites {
		for _, d := range ds {
			if d.pointer && w.Address == d.token+uint32(d.pe) && len(tspHex40A740(w.Hex)) == 4 {
				all[tspWord40A740(tspHex40A740(w.Hex))] = true
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
		if tspRegionAt40A740(in, t) >= 0 {
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
		ri := tspRegionAt40A740(in, d.token)
		b := g.Meta.Banks[ri]
		pe := d.token + uint32(d.pe)
		available := tspAvailable40A740(in, pe, d.width)
		nw := d.width
		if d.pointer {
			nw = int(unsafe.Sizeof(unsafe.Pointer(nil)))
		}
		f := tspField40A740{ri, pe, d.width, b.AnchorOffset + d.native, nw, available, d.pointer}
		g.Meta.Fields = append(g.Meta.Fields, f)
		if available == 0 {
			continue
		}
		r := in.Regions[ri]
		data := append([]byte(nil), tspHex40A740(r.Initial)[int(pe-r.Base):int(pe-r.Base)+available]...)
		if d.pointer && available == 4 {
			p, e := g.pointer(tspWord40A740(data))
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
		if f.NativeOffset < 0 || f.NativeOffset+len(data) > page {
			return nil, fmt.Errorf("partial native field transport bounds")
		}
		ddiCopyIn4E0B30(unsafe.Add(g.Pointers[ri], f.NativeOffset), data)
	}
	// This is a diagnostic C-owned DWORD, not a PE-blob alias of production Go
	// flags. With mapped=false the engine really loads the protected boundary.
	ri := len(in.Regions)
	p, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((ri+1)*page), page)
	if e != nil {
		return nil, e
	}
	g.Pointers = append(g.Pointers, p)
	ddiFill4E0B30(p, page, 0xda)
	offset, available := page, 0
	if in.Engine.Mapped {
		offset, available = 128, 4
		ddiCopyIn4E0B30(unsafe.Add(p, offset), tspBytes40A740(uint64(in.Engine.Value), 4))
	}
	g.Meta.Banks = append(g.Meta.Banks, tspBank40A740{"diagnostic-engine-flags", 0x85b7a0, available, offset, uint64(uintptr(p))})
	g.Meta.Fields = append(g.Meta.Fields, tspField40A740{ri, 0x85b7a0, 4, offset, 4, available, false})
	g.Meta.EngineCell = uint64(uintptr(unsafe.Add(p, offset)))
	for _, p := range append([]unsafe.Pointer{g.Header}, g.Pointers...) {
		if e = rawDefaultDamageSharedSync4E0B30(p, page); e != nil {
			return nil, e
		}
	}
	return g, nil
}
func (g *tspGraph40A740) pointer(t uint32) (unsafe.Pointer, error) {
	if t == 0 {
		return nil, nil
	}
	p, ok := g.Tokens[t]
	if !ok {
		return nil, fmt.Errorf("undeclared PE token %08x", t)
	}
	return p, nil
}
func (g *tspGraph40A740) address(pe uint32, width int) (unsafe.Pointer, *tspField40A740, error) {
	for _, f := range g.Meta.Fields {
		if pe < f.PE {
			continue
		}
		delta := uint64(pe) - uint64(f.PE)
		if width > 0 && delta+uint64(width) <= uint64(f.PEWidth) && (!f.Pointer || (delta == 0 && width == 4)) {
			if delta+uint64(width) > uint64(f.Accessible) {
				return nil, nil, fmt.Errorf("input effect must be fully mapped")
			}
			f.NativeOffset += int(delta)
			f.PE = pe
			f.PEWidth = width
			f.Accessible = width
			if !f.Pointer {
				f.NativeWidth = width
			}
			return unsafe.Add(g.Pointers[f.Region], f.NativeOffset), &f, nil
		}
	}
	return nil, nil, fmt.Errorf("unprojected input effect %08x/%d", pe, width)
}
func (g *tspGraph40A740) apply(writes []tspWrite40A740) error {
	for _, w := range writes {
		data := tspHex40A740(w.Hex)
		p, f, e := g.address(w.Address, len(data))
		if e != nil {
			return e
		}
		if f.Pointer {
			v, e := g.pointer(tspWord40A740(data))
			if e != nil {
				return e
			}
			twoSettingsPointerFixtureStorePointer40A740(p, v)
		} else {
			ddiCopyIn4E0B30(p, data)
		}
	}
	return nil
}
func (g *tspGraph40A740) afterRead(site, pe uint32, width int) error {
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
func (g *tspGraph40A740) afterService(site uint32) error {
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
func (g *tspGraph40A740) snapshot() []byte {
	var out []byte
	for _, p := range g.Pointers {
		out = append(out, ddiSnapshot4E0B30(p, g.Meta.Page)...)
	}
	return out
}
func (g *tspGraph40A740) engineCell() unsafe.Pointer {
	b := g.Meta.Banks[len(g.Meta.Banks)-1]
	return unsafe.Add(g.Pointers[len(g.Pointers)-1], b.AnchorOffset)
}
func (g *tspGraph40A740) close() {
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
func tspOriginalBytes40A740(pages map[uint32][]byte, address uint32, n int) ([]byte, error) {
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
func tspExpectedAfter40A740(meta *tspMeta40A740, before []byte, fields map[uint32][]byte) ([]byte, error) {
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
			t := tspWord40A740(data)
			var p uint64
			if t != 0 {
				var ok bool
				p, ok = meta.Identities[strconv.FormatUint(uint64(t), 10)]
				if !ok {
					return nil, fmt.Errorf("original pointer AFTER identity")
				}
			}
			data = tspBytes40A740(p, f.NativeWidth)
		}
		offset := f.Region*meta.Page + f.NativeOffset
		if f.NativeOffset < 0 || f.NativeOffset+len(data) > meta.Page {
			return nil, fmt.Errorf("native AFTER field span")
		}
		copy(out[offset:], data)
	}
	return out, nil
}

// Fields identify independent static instructions; widths are sealed operand
// widths, never copied from a raw invalid callback's missing fragment width.
func tspOperand40A740(g *tspGraph40A740, e twoSettingsPointerNativeOperand40A740) (uint32, int, error) {
	type d struct {
		field   int
		pe      uint32
		width   int
		pointer bool
	}
	ds := map[uint32]d{0x40a759: {0, 53, 1, false}, 0x40a7a3: {1, 748, 4, true}, 0x40a7a9: {2, 276, 4, true}, 0x40a7af: {3, 3680, 1, false}, 0x40a7b7: {4, 2064, 1, false}, 0x40a7c0: {5, 0x85b7a0, 4, false}, 0x40a7fe: {3, 3680, 1, false}, 0x40a806: {4, 2064, 1, false}, 0x40a80f: {5, 0x85b7a0, 4, false}}
	x, ok := ds[e.Site]
	if !ok || e.Field != x.field || e.IsPointer != x.pointer {
		return 0, 0, fmt.Errorf("static native operand site/field/type")
	}
	width := x.width
	if x.pointer {
		width = int(unsafe.Sizeof(unsafe.Pointer(nil)))
	}
	if int(e.Width) != width || e.Complete != e.ValueValid {
		return 0, 0, fmt.Errorf("actual load completion/width")
	}
	if x.field == 5 {
		if !e.AddressValid || uint64(uintptr(e.Address)) != g.Meta.EngineCell {
			return 0, 0, fmt.Errorf("diagnostic actual DWORD owner")
		}
		return x.pe, x.width, nil
	}
	offsets := []uintptr{53, g.Meta.GoOffsets[0], g.Meta.GoOffsets[1], g.Meta.GoOffsets[2], g.Meta.GoOffsets[3]}
	if e.Offset != offsets[x.field] {
		return 0, 0, fmt.Errorf("named native operand offset")
	}
	token := uint32(0)
	found := e.Base == nil
	for t, p := range g.Tokens {
		if p == e.Base {
			if found {
				return 0, 0, fmt.Errorf("identity collision")
			}
			token = t
			found = true
		}
	}
	if !found {
		return 0, 0, fmt.Errorf("actual operand base identity")
	}
	if e.Base == nil {
		if e.Address != nil {
			return 0, 0, fmt.Errorf("low unsafe.Pointer must not be exposed")
		}
	} else if !e.AddressValid || uintptr(e.Address) != uintptr(e.Base)+e.Offset {
		return 0, 0, fmt.Errorf("actual full foreign EA")
	}
	return token + x.pe, x.width, nil
}
