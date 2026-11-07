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

type seWrite40AA70 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type seRegion40AA70 struct {
	Name       string `json:"name"`
	Base       uint32 `json:"base"`
	Accessible int    `json:"accessible_bytes"`
	Initial    string `json:"initial_hex"`
}
type sePacket40AA70 struct {
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []seWrite40AA70   `json:"writes"`
}
type seSchedule40AA70 struct {
	Target    uint32           `json:"target"`
	Argc      int              `json:"argc32"`
	Repeat    bool             `json:"repeat_last"`
	Responses []sePacket40AA70 `json:"responses"`
}
type seStimulus40AA70 struct {
	PC         uint32          `json:"pc"`
	Site       string          `json:"site"`
	Address    uint32          `json:"address"`
	Width      int             `json:"width"`
	Occurrence int             `json:"occurrence"`
	Writes     []seWrite40AA70 `json:"writes"`
}
type seInput40AA70 struct {
	Args          []uint32                    `json:"args_u32"`
	InitialWords  map[string]uint32           `json:"initial_words"`
	Name          string                      `json:"name"`
	Entry         uint32                      `json:"entry"`
	Roles         map[string]uint32           `json:"roles"`
	Regions       []seRegion40AA70            `json:"regions"`
	Services      map[string]seSchedule40AA70 `json:"services"`
	AfterReads    []seStimulus40AA70          `json:"after_reads"`
	AfterServices []seStimulus40AA70          `json:"after_services"`
}
type seField40AA70 struct {
	Region                                         int
	PE                                             uint32
	PEWidth, NativeOffset, NativeWidth, Accessible int
	Pointer                                        bool
}
type seBank40AA70 struct {
	Name                 string
	PE                   uint32
	Prefix, AnchorOffset int
	Native               uint64
}
type seMeta40AA70 struct {
	Page       int
	Banks      []seBank40AA70
	Fields     []seField40AA70
	Identities map[string]uint64
	CLayout    [3]settingEligibilityFixtureField40AA70
	GoOffsets  [3]uintptr
	Opaque     []uint32
	GlobalCell uint64
}
type seGraph40AA70 struct {
	Input                   *seInput40AA70
	File                    *os.File
	Header                  unsafe.Pointer
	Pointers, OpaqueMaps    []unsafe.Pointer
	Tokens                  map[uint32]unsafe.Pointer
	Meta                    seMeta40AA70
	Reads                   map[string]int
	Services                map[uint32]int
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
}

func seHex40AA70(s string) []byte {
	p, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return p
}
func seWord40AA70(p []byte) uint32 {
	if len(p) > 4 {
		panic("PE scalar width")
	}
	var out uint32
	for i, b := range p {
		out |= uint32(b) << uint(8*i)
	}
	return out
}
func seBytes40AA70(v uint64, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte(v >> uint(8*i))
	}
	return out
}
func sePointerSite40AA70(site uint32) bool {
	switch site {
	case 0x40aa72, 0x40aaed, 0x40ab94, 0x40aba8:
		return true
	}
	return false
}
func seRegionAt40AA70(in *seInput40AA70, pe uint32) int {
	for i, r := range in.Regions {
		if uint64(pe) >= uint64(r.Base) && uint64(pe) < uint64(r.Base)+uint64(r.Accessible) {
			return i
		}
	}
	return -1
}
func seAvailable40AA70(in *seInput40AA70, pe uint32, width int) int {
	i := seRegionAt40AA70(in, pe)
	if i < 0 {
		return 0
	}
	n := int(uint64(in.Regions[i].Base) + uint64(in.Regions[i].Accessible) - uint64(pe))
	if n > width {
		n = width
	}
	return n
}

func seBuildGraph40AA70(in *seInput40AA70, path string) (g *seGraph40AA70, err error) {
	g = &seGraph40AA70{Input: in, Tokens: map[uint32]unsafe.Pointer{}, Reads: map[string]int{}, Services: map[uint32]int{}, UsedResponses: map[string][]int{}}
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
	g.Meta = seMeta40AA70{Page: page, Identities: map[string]uint64{}, CLayout: settingEligibilityFixtureCLayout40AA70(), GoOffsets: [3]uintptr{unsafe.Offsetof(server.Player{}.Field3680), unsafe.Offsetof(server.Player{}.Field2068), unsafe.Offsetof(server.Player{}.Field2140)}}
	for i, w := range []uintptr{4, 4, 4} {
		if g.Meta.CLayout[i].Offset != g.Meta.GoOffsets[i] || g.Meta.CLayout[i].Width != w {
			return nil, fmt.Errorf("actual named C/Go field layout %d", i)
		}
	}
	if settingEligibilityFixtureCPointerWidth40AA70() != unsafe.Sizeof(unsafe.Pointer(nil)) {
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
		if token == 0 || seRegionAt40AA70(in, token) < 0 {
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
			add(role, 52, 52, 1, false)
		case role == "argument":
			add(role, 3680, g.Meta.GoOffsets[0], 4, false)
			add(role, 2068, g.Meta.GoOffsets[1], 4, false)
		case strings.HasPrefix(role, "iterator"):
			add(role, 2140, g.Meta.GoOffsets[2], 4, false)
		}
	}
	sort.Slice(ds, func(i, j int) bool {
		if ds[i].token == ds[j].token {
			return ds[i].pe < ds[j].pe
		}
		return ds[i].token < ds[j].token
	})
	for i, r := range in.Regions {
		initial := seHex40AA70(r.Initial)
		if r.Accessible != len(initial) || r.Accessible < 1 || r.Accessible > 4096 {
			return nil, fmt.Errorf("exact independent region prefix")
		}
		lower, upper := 0, page
		for _, d := range ds {
			if d.token != r.Base {
				continue
			}
			available := seAvailable40AA70(in, d.token+uint32(d.pe), d.width)
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
		b := seBank40AA70{r.Name, r.Base, r.Accessible, page - cut, uint64(uintptr(p))}
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
		if !d.pointer || seAvailable40AA70(in, d.token+uint32(d.pe), 4) != 4 {
			continue
		}
		ri := seRegionAt40AA70(in, d.token)
		r := in.Regions[ri]
		all[seWord40AA70(seHex40AA70(r.Initial)[int(d.token-r.Base)+d.pe:int(d.token-r.Base)+d.pe+4])] = true
	}
	for site, s := range in.Services {
		n, e := strconv.ParseUint(site, 16, 32)
		if e != nil {
			return nil, e
		}
		if sePointerSite40AA70(uint32(n)) {
			for _, r := range s.Responses {
				all[r.Registers["eax"]] = true
			}
		}
	}
	allWrites := []seWrite40AA70{}
	for _, s := range in.Services {
		for _, r := range s.Responses {
			allWrites = append(allWrites, r.Writes...)
		}
	}
	for _, s := range append(append([]seStimulus40AA70{}, in.AfterReads...), in.AfterServices...) {
		allWrites = append(allWrites, s.Writes...)
	}
	for _, w := range allWrites {
		for _, d := range ds {
			if d.pointer && w.Address == d.token+uint32(d.pe) && len(seHex40AA70(w.Hex)) == 4 {
				all[seWord40AA70(seHex40AA70(w.Hex))] = true
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
		if seRegionAt40AA70(in, t) >= 0 {
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
		ri := seRegionAt40AA70(in, d.token)
		b := g.Meta.Banks[ri]
		pe := d.token + uint32(d.pe)
		available := seAvailable40AA70(in, pe, d.width)
		nw := d.width
		if d.pointer {
			nw = int(unsafe.Sizeof(unsafe.Pointer(nil)))
		}
		f := seField40AA70{ri, pe, d.width, b.AnchorOffset + d.native, nw, available, d.pointer}
		g.Meta.Fields = append(g.Meta.Fields, f)
		if available == 0 {
			continue
		}
		r := in.Regions[ri]
		data := append([]byte(nil), seHex40AA70(r.Initial)[int(pe-r.Base):int(pe-r.Base)+available]...)
		if d.pointer && available == 4 {
			p, e := g.pointer(seWord40AA70(data))
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
	// Diagnostic shared DWORD is physically loaded by this observer. The nil
	// observer separately uses real production memmap 5D4594+3508.
	ri := len(in.Regions)
	p, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((ri+1)*page), page)
	if e != nil {
		return nil, e
	}
	g.Pointers = append(g.Pointers, p)
	ddiFill4E0B30(p, page, 0xda)
	value, ok := in.InitialWords["005d5348"]
	if !ok || len(in.InitialWords) != 1 {
		return nil, fmt.Errorf("one independent global DWORD")
	}
	offset := 128
	ddiCopyIn4E0B30(unsafe.Add(p, offset), seBytes40AA70(uint64(value), 4))
	g.Meta.Banks = append(g.Meta.Banks, seBank40AA70{"diagnostic-global3508", 0x5d5348, 4, offset, uint64(uintptr(p))})
	g.Meta.Fields = append(g.Meta.Fields, seField40AA70{ri, 0x5d5348, 4, offset, 4, 4, false})
	g.Meta.GlobalCell = uint64(uintptr(unsafe.Add(p, offset)))
	for _, p := range append([]unsafe.Pointer{g.Header}, g.Pointers...) {
		if e = rawDefaultDamageSharedSync4E0B30(p, page); e != nil {
			return nil, e
		}
	}
	return g, nil
}
func (g *seGraph40AA70) pointer(t uint32) (unsafe.Pointer, error) {
	if t == 0 {
		return nil, nil
	}
	p, ok := g.Tokens[t]
	if !ok {
		return nil, fmt.Errorf("undeclared PE token %08x", t)
	}
	return p, nil
}
func (g *seGraph40AA70) address(pe uint32, width int) (unsafe.Pointer, *seField40AA70, error) {
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
func (g *seGraph40AA70) apply(writes []seWrite40AA70) error {
	for _, w := range writes {
		data := seHex40AA70(w.Hex)
		p, f, e := g.address(w.Address, len(data))
		if e != nil {
			return e
		}
		if f.Pointer {
			v, e := g.pointer(seWord40AA70(data))
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
func (g *seGraph40AA70) afterRead(site, pe uint32, width int) error {
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
func (g *seGraph40AA70) afterService(site uint32) error {
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
func (g *seGraph40AA70) snapshot() []byte {
	var out []byte
	for _, p := range g.Pointers {
		out = append(out, ddiSnapshot4E0B30(p, g.Meta.Page)...)
	}
	return out
}
func (g *seGraph40AA70) globalCell() unsafe.Pointer {
	b := g.Meta.Banks[len(g.Meta.Banks)-1]
	return unsafe.Add(g.Pointers[len(g.Pointers)-1], b.AnchorOffset)
}
func (g *seGraph40AA70) close() {
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
func seOriginalBytes40AA70(pages map[uint32][]byte, address uint32, n int) ([]byte, error) {
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
func seExpectedAfter40AA70(meta *seMeta40AA70, before []byte, fields map[uint32][]byte) ([]byte, error) {
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
			t := seWord40AA70(data)
			var p uint64
			if t != 0 {
				var ok bool
				p, ok = meta.Identities[strconv.FormatUint(uint64(t), 10)]
				if !ok {
					return nil, fmt.Errorf("original pointer AFTER identity")
				}
			}
			data = seBytes40AA70(p, f.NativeWidth)
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
func seOperand40AA70(g *seGraph40AA70, e settingEligibilityNativeOperand40AA70) (uint32, int, error) {
	type descriptor struct {
		field int
		pe    uint32
		width int
	}
	ds := map[uint32]descriptor{0x40aafd: {0, 52, 1}, 0x40aaa8: {1, 3680, 4}, 0x40aadf: {2, 2068, 4}, 0x40ab9d: {3, 2140, 4}, 0x40aab3: {4, 0x5d5348, 4}}
	d, ok := ds[e.Site]
	if !ok || e.Field != d.field || int(e.Width) != d.width || e.Complete != e.ValueValid {
		return 0, 0, fmt.Errorf("static BYTE/DWORD native site/field/width/validity")
	}
	if d.field == 4 {
		if !e.AddressValid || uint64(uintptr(e.Address)) != g.Meta.GlobalCell {
			return 0, 0, fmt.Errorf("actual diagnostic global owner")
		}
		return d.pe, d.width, nil
	}
	offsets := []uintptr{52, g.Meta.GoOffsets[0], g.Meta.GoOffsets[1], g.Meta.GoOffsets[2]}
	if e.Offset != offsets[d.field] {
		return 0, 0, fmt.Errorf("named native offset")
	}
	token := uint32(0)
	found := e.Base == nil
	for t, p := range g.Tokens {
		if p == e.Base {
			if found {
				return 0, 0, fmt.Errorf("native identity collision")
			}
			token = t
			found = true
		}
	}
	if !found {
		return 0, 0, fmt.Errorf("actual base identity")
	}
	if e.Base == nil {
		if e.Address != nil {
			return 0, 0, fmt.Errorf("no exposed low unsafe pointer")
		}
	} else if !e.AddressValid || uintptr(e.Address) != uintptr(e.Base)+e.Offset {
		return 0, 0, fmt.Errorf("actual full native EA")
	}
	return token + d.pe, d.width, nil
}
