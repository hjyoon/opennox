//go:build darwin && arm64

package legacy

// This file transports independent input bytes and identities only. It never
// opens an original row or computes a caller branch, return or selected service.
import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type swWrite410550 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type swRegion410550 struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Page   uint32 `json:"page"`
	Guard  uint32 `json:"guard"`
	Prefix int    `json:"prefix_bytes"`
	Seed   byte   `json:"byte_seed"`
}
type swPacket410550 struct {
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []swWrite410550   `json:"writes"`
}
type swSchedule410550 struct {
	Target    uint32           `json:"target"`
	Argc      int              `json:"argc32"`
	Repeat    bool             `json:"repeat_last"`
	Responses []swPacket410550 `json:"responses"`
}
type swStimulus410550 struct {
	PC         uint32          `json:"pc"`
	Site       string          `json:"site"`
	Address    uint32          `json:"address"`
	Width      int             `json:"width"`
	Occurrence int             `json:"occurrence"`
	Writes     []swWrite410550 `json:"writes"`
}
type swInput410550 struct {
	Name             string                      `json:"name"`
	Family           string                      `json:"family"`
	Entry            uint32                      `json:"entry"`
	Args             []uint32                    `json:"args_u32"`
	Roles            map[string]uint32           `json:"roles"`
	Regions          []swRegion410550            `json:"regions"`
	InitialWords     map[string]uint32           `json:"initial_words"`
	InitialWrites    []swWrite410550             `json:"initial_writes"`
	Services         map[string]swSchedule410550 `json:"services"`
	AfterReads       []swStimulus410550          `json:"after_reads"`
	AfterServices    []swStimulus410550          `json:"after_services"`
	InitialRegisters map[string]uint32           `json:"initial_registers_u32"`
	InitialEflags    uint32                      `json:"initial_eflags"`
	StackSeed        byte                        `json:"stack_byte_seed"`
}
type swField410550 struct {
	Region                                         int
	PE                                             uint32
	PEWidth, NativeOffset, NativeWidth, Accessible int
	Pointer, Signed                                bool
}
type swBank410550 struct {
	Name                 string
	PE                   uint32
	Prefix, AnchorOffset int
	Native               uint64
}
type swMeta410550 struct {
	Page                            int
	Banks                           []swBank410550
	Fields                          []swField410550
	Identities                      map[string]uint64
	CLayout                         [4]uintptr
	GoOffsets                       [3]uintptr
	CScalarOffsets, GoScalarOffsets [6]uintptr
	Opaque                          []uint32
}
type swGraph410550 struct {
	Input                   *swInput410550
	File                    *os.File
	Header, Head            unsafe.Pointer
	Pointers, OpaqueMaps    []unsafe.Pointer
	Tokens                  map[uint32]unsafe.Pointer
	Meta                    swMeta410550
	Reads                   map[string]int
	Services                map[uint32]int
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
}

func swHex410550(s string) []byte {
	p, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return p
}
func swWord410550(p []byte) uint32 {
	if len(p) > 4 {
		panic("PE scalar width")
	}
	var out uint32
	for i, b := range p {
		out |= uint32(b) << uint(8*i)
	}
	return out
}
func swBytes410550(v uint64, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte(v >> uint(8*i))
	}
	return out
}

func swPointerSite410550(site uint32) bool {
	switch site {
	case 0x410551, 0x410569, 0x4107b8, 0x4107d2, 0x4107eb:
		return true
	}
	return false
}
func swRegionAt410550(in *swInput410550, pe uint32) int {
	for i, r := range in.Regions {
		base := uint64(r.Guard) - uint64(r.Prefix)
		if uint64(pe) >= base && uint64(pe) < uint64(r.Guard) {
			return i
		}
	}
	return -1
}
func swAvailable410550(in *swInput410550, pe uint32, width int) int {
	if pe >= 0x611c3c && uint64(pe)+uint64(width) <= 0x611c40 {
		return width
	}
	i := swRegionAt410550(in, pe)
	if i < 0 {
		return 0
	}
	n := int(uint64(in.Regions[i].Guard) - uint64(pe))
	if n > width {
		n = width
	}
	return n
}
func swInitialPages410550(in *swInput410550) (map[uint32][]byte, error) {
	pages := map[uint32][]byte{}
	for _, r := range in.Regions {
		if r.Page%4096 != 0 || r.Guard != r.Page+4096 || r.Prefix < 0 || r.Prefix > 36 || pages[r.Page] != nil {
			return nil, fmt.Errorf("independent input page geometry")
		}
		b := make([]byte, 4096)
		for i := range b {
			b[i] = byte(int(r.Seed) + 29*i + 7*(i>>4))
		}
		pages[r.Page] = b
	}
	b := make([]byte, 4096)
	for i := range b {
		b[i] = byte(183 + 29*i + 7*(i>>4))
	}
	pages[0x611000] = b
	head, ok := in.InitialWords["00611c3c"]
	if !ok || len(in.InitialWords) != 1 {
		return nil, fmt.Errorf("literal original head input")
	}
	copy(b[0xc3c:], swBytes410550(uint64(head), 4))
	for _, w := range in.InitialWrites {
		data := swHex410550(w.Hex)
		if len(data) < 1 || len(data) > 4 {
			return nil, fmt.Errorf("literal initializer width")
		}
		if swAvailable410550(in, w.Address, len(data)) != len(data) {
			return nil, fmt.Errorf("literal initializer logical prefix")
		}
		page := w.Address &^ uint32(4095)
		off := int(w.Address - page)
		if pages[page] == nil || off+len(data) > 4096 {
			return nil, fmt.Errorf("literal initializer physical span")
		}
		copy(pages[page][off:], data)
	}
	return pages, nil
}
func swBuildGraph410550(in *swInput410550, path string) (g *swGraph410550, err error) {
	g = &swGraph410550{Input: in, Tokens: map[uint32]unsafe.Pointer{}, Reads: map[string]int{}, Services: map[uint32]int{}, UsedResponses: map[string][]int{}}
	owner := g
	defer func() {
		if err != nil {
			owner.close()
		}
	}()
	page := os.Getpagesize()
	if page < 8192 || page&(page-1) != 0 {
		return nil, fmt.Errorf("native guard-page geometry")
	}
	g.Meta = swMeta410550{Page: page, Identities: map[string]uint64{}, CLayout: secretWallFixtureLayout410550(), GoOffsets: secretWallNativeOffsets410550()}
	var named server.SecretWall
	g.Meta.CScalarOffsets = secretWallFixtureScalarOffsets410550()
	g.Meta.GoScalarOffsets = [6]uintptr{unsafe.Offsetof(named.X), unsafe.Offsetof(named.Y), unsafe.Offsetof(named.OpenWait), unsafe.Offsetof(named.Flags), unsafe.Offsetof(named.LastOpen), unsafe.Offsetof(named.PlayerBits)}
	for i, v := range g.Meta.GoScalarOffsets {
		if g.Meta.CScalarOffsets[i] != v {
			return nil, fmt.Errorf("actual named C/Go scalar offset %d", i)
		}
	}
	if unsafe.Offsetof(named.LastOpen)-unsafe.Offsetof(named.Flags) != 4 {
		return nil, fmt.Errorf("four native adjacent BYTEs")
	}
	pw := int(unsafe.Sizeof(unsafe.Pointer(nil)))
	if secretWallFixturePointerWidth410550() != uintptr(pw) || g.Meta.CLayout[3] != unsafe.Sizeof(server.SecretWall{}) || g.Meta.CLayout[3] != uintptr(32+2*(pw-4)) {
		return nil, fmt.Errorf("actual C record/pointer width")
	}
	for i, v := range g.Meta.GoOffsets {
		if g.Meta.CLayout[i] != v {
			return nil, fmt.Errorf("actual named C/Go offset %d", i)
		}
	}
	if g.Meta.GoOffsets[0] != 0 || g.Meta.GoOffsets[2] != 10 {
		return nil, fmt.Errorf("named Next/ID offsets")
	}
	if len(in.Regions) != 5 {
		return nil, fmt.Errorf("five independent physical input regions")
	}
	initial, e := swInitialPages410550(in)
	if e != nil {
		return nil, e
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
	type descriptor struct {
		token             uint32
		pe, native, width int
		pointer           bool
	}
	var ds []descriptor
	add := func(token uint32, pe int, native uintptr, width int, pointer bool) error {
		d := descriptor{token, pe, int(native), width, pointer}
		for _, old := range ds {
			if old.token == token && old.pe == pe {
				if old != d {
					return fmt.Errorf("conflicting same-base alias descriptor")
				}
				return nil
			}
		}
		ds = append(ds, d)
		return nil
	}
	for _, name := range []string{"node0", "node1", "node2", "wall0", "wall1"} {
		token, ok := in.Roles[name]
		if !ok {
			return nil, fmt.Errorf("missing input role")
		}
		kind := "node"
		if name[:4] == "wall" {
			kind = "wall"
		}
		found := false
		for _, r := range in.Regions {
			if r.Kind == kind && token == r.Guard-uint32(r.Prefix) {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("only same-kind same-base input aliases")
		}
		if kind == "node" {
			if e = add(token, 0, g.Meta.GoOffsets[0], 4, true); e != nil {
				return nil, e
			}
			if e = add(token, 12, g.Meta.GoOffsets[1], 4, true); e != nil {
				return nil, e
			}
			for i, pe := range []int{4, 8, 16, 20, 24, 28} {
				if e = add(token, pe, g.Meta.GoScalarOffsets[i], 4, false); e != nil {
					return nil, e
				}
			}
		} else {
			if e = add(token, 10, g.Meta.GoOffsets[2], 2, false); e != nil {
				return nil, e
			}
		}
	}
	sort.Slice(ds, func(i, j int) bool {
		if ds[i].token == ds[j].token {
			return ds[i].pe < ds[j].pe
		}
		return ds[i].token < ds[j].token
	})
	for i, r := range in.Regions {
		cut := r.Prefix
		if r.Kind == "node" {
			if r.Prefix > 32 {
				return nil, fmt.Errorf("node input prefix")
			}
			if r.Prefix >= 4 {
				cut += pw - 4
			}
			if r.Prefix >= 16 {
				cut += pw - 4
			}
		} else if r.Kind != "wall" {
			return nil, fmt.Errorf("input region kind")
		}
		if cut < 0 || cut > page {
			return nil, fmt.Errorf("native lifted prefix")
		}
		bank, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((i+1)*page), page)
		if e != nil {
			return nil, e
		}
		g.Pointers = append(g.Pointers, bank)
		if uintptr(bank) <= 0xffffffff {
			return nil, fmt.Errorf("high native C-owned bank")
		}
		ddiFill4E0B30(bank, page, byte(0xb0+i))
		base := r.Guard - uint32(r.Prefix)
		anchor := page - cut
		g.Meta.Banks = append(g.Meta.Banks, swBank410550{r.Name, base, r.Prefix, anchor, uint64(uintptr(bank))})
		if _, exists := g.Tokens[base]; exists {
			return nil, fmt.Errorf("independent physical base identity")
		}
		g.Tokens[base] = unsafe.Add(bank, anchor)
	}
	bank, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64(6*page), page)
	if e != nil {
		return nil, e
	}
	g.Pointers = append(g.Pointers, bank)
	ddiFill4E0B30(bank, page, 0xb5)
	g.Meta.Banks = append(g.Meta.Banks, swBank410550{"head", 0x611c3c, 4, page - pw, uint64(uintptr(bank))})
	g.Head = unsafe.Add(bank, page-pw)
	headWitness := twoSettingsPointerFixturePointerWitness40A740(g.Head)
	if uintptr(g.Head) <= 0xffffffff || headWitness.Native != uintptr(g.Head) || headWitness.Low32 != uint32(uintptr(g.Head)) || !headWitness.Above4GiB || !headWitness.Low32LosesBits {
		return nil, fmt.Errorf("actual shared head C pointer witness")
	}
	all := map[uint32]bool{}
	for _, v := range in.Roles {
		all[v] = true
	}
	all[in.InitialWords["00611c3c"]] = true
	if in.Entry == 0x4107a0 {
		if len(in.Args) != 1 {
			return nil, fmt.Errorf("remove input pointer argument")
		}
		all[in.Args[0]] = true
	}
	for _, d := range ds {
		if d.pointer && swAvailable410550(in, d.token+uint32(d.pe), 4) == 4 {
			data, e := swOriginalBytes410550(initial, d.token+uint32(d.pe), 4)
			if e != nil {
				return nil, e
			}
			all[swWord410550(data)] = true
		}
	}
	writes := []swWrite410550{}
	for site, s := range in.Services {
		n, e := strconv.ParseUint(site, 16, 32)
		if e != nil {
			return nil, e
		}
		for _, packet := range s.Responses {
			if swPointerSite410550(uint32(n)) {
				all[packet.Registers["eax"]] = true
			}
			writes = append(writes, packet.Writes...)
		}
	}
	for _, s := range in.AfterReads {
		writes = append(writes, s.Writes...)
	}
	for _, s := range in.AfterServices {
		writes = append(writes, s.Writes...)
	}
	for _, w := range writes {
		data := swHex410550(w.Hex)
		if w.Address == 0x611c3c && len(data) == 4 {
			all[swWord410550(data)] = true
		}
		for _, d := range ds {
			if d.pointer && w.Address == d.token+uint32(d.pe) && len(data) == 4 {
				all[swWord410550(data)] = true
			}
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
		if _, ok := g.Tokens[token]; ok {
			continue
		}
		if swRegionAt410550(in, token) >= 0 {
			return nil, fmt.Errorf("undeclared interior pointer token")
		}
		p := ddiMap4E0B30(page)
		if p == nil {
			return nil, fmt.Errorf("opaque identity guard allocation")
		}
		g.OpaqueMaps = append(g.OpaqueMaps, p)
		g.Tokens[token] = unsafe.Add(p, page)
		g.Meta.Opaque = append(g.Meta.Opaque, token)
	}
	seen := map[unsafe.Pointer]bool{}
	for token, p := range g.Tokens {
		if p == nil || uintptr(p) <= 0xffffffff || seen[p] {
			return nil, fmt.Errorf("high distinct canonical native identity")
		}
		w := twoSettingsPointerFixturePointerWitness40A740(p)
		if w.Native != uintptr(p) || w.Low32 != uint32(uintptr(p)) || !w.Above4GiB || !w.Low32LosesBits {
			return nil, fmt.Errorf("actual C full-pointer numeric witness")
		}
		seen[p] = true
		g.Meta.Identities[strconv.FormatUint(uint64(token), 10)] = uint64(uintptr(p))
	}
	for _, d := range ds {
		ri := -1
		for i, b := range g.Meta.Banks[:5] {
			if b.PE == d.token {
				ri = i
				break
			}
		}
		if ri < 0 {
			return nil, fmt.Errorf("input base physical bank")
		}
		b := g.Meta.Banks[ri]
		pe := d.token + uint32(d.pe)
		available := swAvailable410550(in, pe, d.width)
		nw := d.width
		if d.pointer {
			nw = pw
		}
		f := swField410550{ri, pe, d.width, b.AnchorOffset + d.native, nw, available, d.pointer, false}
		g.Meta.Fields = append(g.Meta.Fields, f)
		if available == 0 {
			continue
		}
		data, e := swOriginalBytes410550(initial, pe, available)
		if e != nil {
			return nil, e
		}
		if f.Pointer && available == 4 {
			v, e := g.pointer(swWord410550(data))
			if e != nil {
				return nil, e
			}
			if f.NativeOffset < 0 || f.NativeOffset+pw > page {
				return nil, fmt.Errorf("complete native pointer field")
			}
			slot := unsafe.Add(g.Pointers[ri], f.NativeOffset)
			secretWallFixtureStorePointer410550(slot, v)
			if secretWallFixtureLoadPointer410550(slot) != v {
				return nil, fmt.Errorf("C full-pointer memcpy roundtrip")
			}
		} else {
			if f.NativeOffset < 0 || f.NativeOffset+len(data) > page {
				return nil, fmt.Errorf("raw partial pointer/scalar field")
			}
			ddiCopyIn4E0B30(unsafe.Add(g.Pointers[ri], f.NativeOffset), data)
		}
	}
	g.Meta.Fields = append(g.Meta.Fields, swField410550{5, 0x611c3c, 4, page - pw, pw, 4, true, false})
	head, e := g.pointer(in.InitialWords["00611c3c"])
	if e != nil {
		return nil, e
	}
	secretWallFixtureStorePointer410550(g.Head, head)
	for _, p := range append([]unsafe.Pointer{g.Header}, g.Pointers...) {
		if e = rawDefaultDamageSharedSync4E0B30(p, page); e != nil {
			return nil, e
		}
	}
	return g, nil
}
func (g *swGraph410550) pointer(token uint32) (unsafe.Pointer, error) {
	if token == 0 {
		return nil, nil
	}
	p, ok := g.Tokens[token]
	if !ok {
		return nil, fmt.Errorf("undeclared original pointer token %08x", token)
	}
	return p, nil
}
func (g *swGraph410550) address(pe uint32, width int) (unsafe.Pointer, *swField410550, error) {
	for _, field := range g.Meta.Fields {
		f := field
		if pe < f.PE {
			continue
		}
		delta := uint64(pe) - uint64(f.PE)
		if width > 0 && delta+uint64(width) <= uint64(f.PEWidth) && (!f.Pointer || (delta == 0 && width == 4)) {
			if delta+uint64(width) > uint64(f.Accessible) {
				return nil, nil, fmt.Errorf("effect requires complete input field")
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
func (g *swGraph410550) apply(writes []swWrite410550) error {
	for _, w := range writes {
		data := swHex410550(w.Hex)
		p, f, e := g.address(w.Address, len(data))
		if e != nil {
			return e
		}
		if f.Pointer {
			v, e := g.pointer(swWord410550(data))
			if e != nil {
				return e
			}
			secretWallFixtureStorePointer410550(p, v)
		} else {
			ddiCopyIn4E0B30(p, data)
		}
	}
	return nil
}
func (g *swGraph410550) afterRead(site, pe uint32, width int) error {
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
func (g *swGraph410550) afterService(site uint32) error {
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
func (g *swGraph410550) snapshot() []byte {
	var out []byte
	for _, p := range g.Pointers {
		out = append(out, ddiSnapshot4E0B30(p, g.Meta.Page)...)
	}
	return out
}
func (g *swGraph410550) close() {
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
func swOriginalBytes410550(pages map[uint32][]byte, address uint32, n int) ([]byte, error) {
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
func swExpectedAfter410550(meta *swMeta410550, before []byte, fields map[uint32][]byte) ([]byte, error) {
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
			t := swWord410550(data)
			var p uint64
			if t != 0 {
				var ok bool
				p, ok = meta.Identities[strconv.FormatUint(uint64(t), 10)]
				if !ok {
					return nil, fmt.Errorf("original pointer AFTER identity")
				}
			}
			data = swBytes410550(p, f.NativeWidth)
		}
		if f.Signed && f.Accessible == 4 {
			data = swBytes410550(uint64(int64(int32(swWord410550(data)))), 8)
		}
		offset := f.Region*meta.Page + f.NativeOffset
		if f.NativeOffset < 0 || f.NativeOffset+len(data) > meta.Page {
			return nil, fmt.Errorf("native AFTER field span")
		}
		copy(out[offset:], data)
	}
	return out, nil
}

// Reached-site argument cells are real private Go storage, not the original
// caller stack. Their logical PE argument slot is compared as metadata only.
func swOperand410550(g *swGraph410550, e secretWallNativeOperand410550) (uint32, int, error) {
	type descriptor struct {
		field  int
		offset uint32
		width  int
		write  bool
	}
	table := map[uint32]descriptor{0x41055a: {4, 0, 2, false}, 0x41055f: {2, 12, 4, false}, 0x410562: {3, 10, 2, false}, 0x410579: {2, 12, 4, false}, 0x410730: {0, 0, 4, false}, 0x41073a: {1, 0, 4, false}, 0x41074b: {0, 0, 4, true}, 0x410753: {0, 0, 4, true}, 0x4107a0: {0, 0, 4, false}, 0x4107ad: {5, 0, 4, false}, 0x4107c7: {0, 0, 4, false}, 0x4107da: {0, 0, 4, true}, 0x4107f3: {1, 0, 4, true}}
	d, ok := table[e.Site]
	if !ok || e.Field != d.field || e.Write != d.write {
		return 0, 0, fmt.Errorf("sealed native site/field/access")
	}
	pointer := d.field == 0 || d.field == 1 || d.field == 2 || d.field == 5
	nw := d.width
	if pointer {
		nw = int(unsafe.Sizeof(unsafe.Pointer(nil)))
	}
	offsets := []uintptr{0, g.Meta.GoOffsets[0], g.Meta.GoOffsets[1], g.Meta.GoOffsets[2], 0, 0}
	if e.IsPointer != pointer || int(e.Width) != nw || e.Offset != offsets[d.field] || !e.AddressValid || (!e.Write && e.ValueValid != e.Complete) || (e.Write && !e.ValueValid) {
		return 0, 0, fmt.Errorf("actual native width/offset/validity")
	}
	if d.field == 4 || d.field == 5 {
		if e.Base == nil || e.Address != e.Base {
			return 0, 0, fmt.Errorf("actual reached private argument cell")
		}
		return 0x22000804, d.width, nil
	}
	if d.field == 0 {
		if e.Base != g.Head || e.Address != g.Head {
			return 0, 0, fmt.Errorf("selected actual shared head cell")
		}
		return 0x611c3c, 4, nil
	}
	token := uint32(0)
	found := e.Base == nil
	for t, p := range g.Tokens {
		if e.Base == p {
			if found {
				return 0, 0, fmt.Errorf("canonical identity collision")
			}
			token = t
			found = true
		}
	}
	if !found {
		return 0, 0, fmt.Errorf("undeclared native operand base")
	}
	if e.Base == nil {
		if e.Address != nil {
			return 0, 0, fmt.Errorf("known actual nil trap, original EA remains separate")
		}
	} else if e.Address != unsafe.Add(e.Base, e.Offset) {
		return 0, 0, fmt.Errorf("actual native effective operand")
	}
	return token + d.offset, d.width, nil
}
