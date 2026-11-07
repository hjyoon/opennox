//go:build darwin && arm64

package legacy

// The graph is built from independently chosen PE32 input bytes. Original
// observations are consumed only by the parent after the real child exits.
// It reuses guarded foreign allocation/signal primitives, not the DefaultDamage
// object graph, field map, reference fixture or native algorithm.
import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type rskWrite4E23C0 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type rskInitializer4E23C0 struct {
	Offset int    `json:"offset"`
	Hex    string `json:"hex"`
}
type rskRegion4E23C0 struct {
	Name         string                 `json:"name"`
	Address      uint32                 `json:"address"`
	Size         int                    `json:"size"`
	Fill         byte                   `json:"fill_byte"`
	Initializers []rskInitializer4E23C0 `json:"initializers"`
}
type rskPacket4E23C0 struct {
	Target    uint32            `json:"target"`
	Argc      int               `json:"argc32"`
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []rskWrite4E23C0  `json:"writes"`
}
type rskStimulus4E23C0 struct {
	PC         uint32           `json:"pc"`
	Site       string           `json:"site"`
	Address    uint32           `json:"address"`
	Width      int              `json:"width"`
	Occurrence int              `json:"occurrence"`
	Writes     []rskWrite4E23C0 `json:"writes"`
}
type rskInput4E23C0 struct {
	Name          string                     `json:"name"`
	Family        string                     `json:"family"`
	Args          []uint32                   `json:"args_u32"`
	Roles         map[string]uint32          `json:"roles"`
	Regions       []rskRegion4E23C0          `json:"regions"`
	Services      map[string]rskPacket4E23C0 `json:"services"`
	AfterReads    []rskStimulus4E23C0        `json:"after_reads"`
	AfterServices []rskStimulus4E23C0        `json:"after_services"`
	Prefixes      map[string]int             `json:"declared_readable_prefixes"`
}
type rskField4E23C0 struct {
	Region       int    `json:"region"`
	PE           uint32 `json:"pe"`
	PEWidth      int    `json:"pe_width"`
	NativeOffset int    `json:"native_offset"`
	NativeWidth  int    `json:"native_width"`
	Accessible   int    `json:"accessible"`
	Pointer      bool   `json:"pointer"`
}
type rskBank4E23C0 struct {
	Name         string `json:"name"`
	PE           uint32 `json:"pe"`
	Size         int    `json:"pe_size"`
	Native       uint64 `json:"native"`
	AnchorPE     uint32 `json:"anchor_pe"`
	AnchorOffset int    `json:"anchor_offset"`
}
type rskGraphMetadata4E23C0 struct {
	Page       int               `json:"page"`
	Banks      []rskBank4E23C0   `json:"banks"`
	Fields     []rskField4E23C0  `json:"fields"`
	Identities map[string]uint64 `json:"identities"`
	Offsets    [7]uintptr        `json:"named_Go_offsets"`
	COffsets   [5]uintptr        `json:"actual_C_object_offsets"`
	Opaque     []uint32          `json:"opaque_guard_tokens"`
	Scope      string            `json:"scope"`
}
type rskGraph4E23C0 struct {
	Input        *rskInput4E23C0
	File         *os.File
	Header       unsafe.Pointer
	Pointers     []unsafe.Pointer
	OpaqueMaps   []unsafe.Pointer
	Tokens       map[uint32]unsafe.Pointer
	PEBytes      [][]byte
	Meta         rskGraphMetadata4E23C0
	Reads        map[uint32]int
	Services     map[uint32]int
	UsedReads    []int
	UsedServices []int
	UsedPackets  []uint32
}

func rskHex4E23C0(value string) []byte {
	p, e := hex.DecodeString(value)
	if e != nil {
		panic(e)
	}
	return p
}
func rskWord4E23C0(p []byte) uint32 {
	if len(p) > 4 {
		panic("scalar PE width")
	}
	var v uint32
	for i, b := range p {
		v |= uint32(b) << uint(8*i)
	}
	return v
}
func rskPutWord4E23C0(value uint64, width int) []byte {
	p := make([]byte, width)
	for i := range p {
		p[i] = byte(value >> uint(8*i))
	}
	return p
}
func rskHash4E23C0(p []byte) string { x := sha256.Sum256(p); return hex.EncodeToString(x[:]) }
func rskWriteJSON4E23C0(path string, value any) error {
	p, e := json.Marshal(value)
	if e != nil {
		return e
	}
	p = append(p, '\n')
	return os.WriteFile(path, p, 0600)
}

func rskInputPages4E23C0(in *rskInput4E23C0) (out [][]byte) {
	if len(in.Args) != 5 || len(in.Regions) != 6 {
		panic("Skeleton exact input geometry")
	}
	for _, r := range in.Regions {
		if r.Size != 4096 {
			panic("PE page size")
		}
		p := bytes.Repeat([]byte{r.Fill}, r.Size)
		for _, v := range r.Initializers {
			b := rskHex4E23C0(v.Hex)
			if v.Offset < 0 || v.Offset+len(b) > len(p) {
				panic("initializer bounds")
			}
			copy(p[v.Offset:], b)
		}
		out = append(out, p)
	}
	return
}
func rskRegionIndex4E23C0(in *rskInput4E23C0, address uint32) int {
	for i, r := range in.Regions {
		if uint64(address) >= uint64(r.Address) && uint64(address) < uint64(r.Address)+uint64(r.Size) {
			return i
		}
	}
	return -1
}
func rskAvailable4E23C0(in *rskInput4E23C0, address uint32, width int) int {
	i := rskRegionIndex4E23C0(in, address)
	if i < 0 {
		return 0
	}
	n := int(uint64(in.Regions[i].Address) + uint64(in.Regions[i].Size) - uint64(address))
	if n > width {
		n = width
	}
	return n
}
func rskPERead4E23C0(in *rskInput4E23C0, pages [][]byte, address uint32, width int) []byte {
	i := rskRegionIndex4E23C0(in, address)
	if i < 0 || rskAvailable4E23C0(in, address, width) != width {
		panic("PE declared readable span")
	}
	offset := int(address - in.Regions[i].Address)
	return append([]byte(nil), pages[i][offset:offset+width]...)
}

func rskBuildGraph4E23C0(in *rskInput4E23C0, path string) (g *rskGraph4E23C0, err error) {
	g = &rskGraph4E23C0{Input: in, Tokens: make(map[uint32]unsafe.Pointer), PEBytes: rskInputPages4E23C0(in), Reads: make(map[uint32]int), Services: make(map[uint32]int)}
	owner := g
	defer func() {
		if err != nil {
			owner.close()
		}
	}()
	page := os.Getpagesize()
	if page < 4096 || page&(page-1) != 0 {
		return nil, fmt.Errorf("native page size")
	}
	g.Meta = rskGraphMetadata4E23C0{Page: page, Offsets: rawSkeletonOffsets4E23C0(), COffsets: rawSkeletonFixtureCObjectOffsets4E23C0(), Identities: make(map[string]uint64), Scope: "foreign C-owned selected-field projection; nil physical EA and opaque accessible-body equivalence are not claimed"}
	for i := 0; i < 5; i++ {
		if g.Meta.Offsets[i] != g.Meta.COffsets[i] {
			return nil, fmt.Errorf("actual C/Go named object offset %d", i)
		}
	}
	g.File, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	if err = g.File.Truncate(int64((1 + len(in.Regions)) * page)); err != nil {
		return nil, err
	}
	g.Header, err = rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), 0, page)
	if err != nil {
		return nil, err
	}
	ddiFill4E0B30(g.Header, page, 0)
	*(*uint64)(g.Header) = rawDefaultDamageSharedMagic4E0B30
	*(*uint64)(unsafe.Add(g.Header, 8)) = rawDefaultDamageSharedVersion4E0B30
	// A descriptor is source-layout input transport, never selected by an
	// original branch/result. Duplicate alias descriptors must agree exactly.
	type desc struct {
		token   uint32
		pe      int
		native  int
		width   int
		pointer bool
	}
	var ds []desc
	add := func(role string, pe int, native uintptr, width int, pointer bool) {
		token := in.Roles[role]
		if token == 0 || rskRegionIndex4E23C0(in, token) < 0 {
			return
		}
		d := desc{token, pe, int(native), width, pointer}
		for _, old := range ds {
			if old.token == d.token && old.pe == d.pe {
				if old != d {
					panic("alias descriptor conflict")
				}
				return
			}
		}
		ds = append(ds, d)
	}
	add("target", 8, unsafe.Offsetof(server.Object{}.ObjClass), 4, false)
	add("target", 56, g.Meta.Offsets[0], 8, false)
	for _, role := range []string{"target", "source", "weapon"} {
		add(role, 72, g.Meta.Offsets[1], 4, false)
		add(role, 76, g.Meta.Offsets[2], 4, false)
	}
	add("target", 124, g.Meta.Offsets[3], 2, false)
	add("target", 748, g.Meta.Offsets[4], 4, true)
	for _, role := range []string{"update", "alternate_update"} {
		add(role, 480, g.Meta.Offsets[5], 1, false)
		add(role, 481, g.Meta.Offsets[6], 1, false)
	}
	// One anchor per actual PE allocation. Every alias of an anchor preserves
	// one full native identity. Do not choose an update view from the class.
	roles := []string{"target", "source", "weapon", "update", "alternate_update"}
	for i, r := range in.Regions {
		anchor := uint32(0)
		anchorRole := ""
		for _, role := range roles {
			t := in.Roles[role]
			if rskRegionIndex4E23C0(in, t) == i {
				if anchor != 0 && anchor != t {
					return nil, fmt.Errorf("multiple distinct anchors in PE region")
				}
				anchor = t
				anchorRole = role
			}
		}
		bank := rskBank4E23C0{Name: r.Name, PE: r.Address, Size: r.Size, AnchorPE: anchor}
		if anchor != 0 {
			bank.AnchorOffset = int(anchor - r.Address)
			if prefix, ok := in.Prefixes[anchorRole]; ok {
				lower, upper := 0, page
				for _, d := range ds {
					if d.token != anchor {
						continue
					}
					available := rskAvailable4E23C0(in, d.token+uint32(d.pe), d.width)
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
				cutoff := upper
				if upper == page {
					cutoff = lower
				}
				if prefix == 0 {
					cutoff = 0
				}
				if lower > upper || cutoff < 0 || cutoff > page {
					return nil, fmt.Errorf("native prefix constraints")
				}
				bank.AnchorOffset = page - cutoff
			}
		}
		p, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((i+1)*page), page)
		if e != nil {
			return nil, e
		}
		if uintptr(p) <= 0xffffffff {
			return nil, fmt.Errorf("actual foreign mapping must exceed 4GiB")
		}
		g.Pointers = append(g.Pointers, p)
		ddiFill4E0B30(p, page, byte(0xb0+i))
		bank.Native = uint64(uintptr(p))
		g.Meta.Banks = append(g.Meta.Banks, bank)
		if anchor != 0 {
			g.Tokens[anchor] = unsafe.Add(p, bank.AnchorOffset)
		}
	}
	// Opaque words are carried by distinct native guard identities. They are
	// not decoded as raw native addresses and expose no fabricated object body.
	tokens := make([]uint32, 0, len(in.Roles))
	for _, v := range in.Roles {
		tokens = append(tokens, v)
	}
	sort.Slice(tokens, func(i, j int) bool { return tokens[i] < tokens[j] })
	for _, v := range tokens {
		if v == 0 {
			continue
		}
		if _, ok := g.Tokens[v]; ok {
			continue
		}
		p := ddiMap4E0B30(page)
		if p == nil {
			return nil, fmt.Errorf("opaque foreign guard map")
		}
		g.OpaqueMaps = append(g.OpaqueMaps, p)
		g.Tokens[v] = unsafe.Add(p, page)
		g.Meta.Opaque = append(g.Meta.Opaque, v)
	}
	for k, v := range g.Tokens {
		if uintptr(v) <= 0xffffffff {
			return nil, fmt.Errorf("native full identity")
		}
		g.Meta.Identities[strconv.FormatUint(uint64(k), 10)] = uint64(uintptr(v))
	}
	for _, d := range ds {
		ri := rskRegionIndex4E23C0(in, d.token)
		b := g.Meta.Banks[ri]
		pe := d.token + uint32(d.pe)
		available := rskAvailable4E23C0(in, pe, d.width)
		nw := d.width
		if d.pointer {
			nw = int(unsafe.Sizeof(unsafe.Pointer(nil)))
		}
		f := rskField4E23C0{Region: ri, PE: pe, PEWidth: d.width, NativeOffset: b.AnchorOffset + d.native, NativeWidth: nw, Accessible: available, Pointer: d.pointer}
		g.Meta.Fields = append(g.Meta.Fields, f)
		if available == 0 {
			continue
		}
		data := rskPERead4E23C0(in, g.PEBytes, pe, available)
		if d.pointer && available == d.width {
			pointer, e := g.pointer(rskWord4E23C0(data))
			if e != nil {
				return nil, e
			}
			data = rskPutWord4E23C0(uint64(uintptr(pointer)), nw)
		}
		if f.NativeOffset < 0 || f.NativeOffset+len(data) > page {
			return nil, fmt.Errorf("native partial span placement")
		}
		ddiCopyIn4E0B30(unsafe.Add(g.Pointers[ri], f.NativeOffset), data)
	}
	// The runtime clock is input memory, not the product frame service and not
	// an expected byte callback. Its four independently selected bytes survive.
	ri := rskRegionIndex4E23C0(in, 0x84ea04)
	if ri < 0 {
		return nil, fmt.Errorf("clock input region")
	}
	offset := int(uint32(0x84ea04) - in.Regions[ri].Address)
	g.Meta.Fields = append(g.Meta.Fields, rskField4E23C0{Region: ri, PE: 0x84ea04, PEWidth: 4, NativeOffset: offset, NativeWidth: 4, Accessible: 4})
	ddiCopyIn4E0B30(unsafe.Add(g.Pointers[ri], offset), rskPERead4E23C0(in, g.PEBytes, 0x84ea04, 4))
	for _, p := range append([]unsafe.Pointer{g.Header}, g.Pointers...) {
		if e := rawDefaultDamageSharedSync4E0B30(p, page); e != nil {
			return nil, e
		}
	}
	return g, nil
}

func (g *rskGraph4E23C0) pointer(token uint32) (unsafe.Pointer, error) {
	if token == 0 {
		return nil, nil
	}
	p, ok := g.Tokens[token]
	if !ok {
		return nil, fmt.Errorf("undeclared canonical token %08x", token)
	}
	return p, nil
}
func (g *rskGraph4E23C0) address(pe uint32, width int) (unsafe.Pointer, *rskField4E23C0, error) {
	for i := range g.Meta.Fields {
		f := &g.Meta.Fields[i]
		if pe < f.PE {
			continue
		}
		delta := uint64(pe) - uint64(f.PE)
		if width > 0 && delta+uint64(width) <= uint64(f.PEWidth) && (!f.Pointer || (delta == 0 && width == f.PEWidth)) {
			if delta+uint64(width) > uint64(f.Accessible) {
				return nil, nil, fmt.Errorf("declared scalar span is not fully mapped")
			}
			view := *f
			view.NativeOffset += int(delta)
			view.PE = pe
			view.PEWidth = width
			view.Accessible = width
			if !f.Pointer {
				view.NativeWidth = width
			}
			return unsafe.Add(g.Pointers[f.Region], view.NativeOffset), &view, nil
		}
	}
	return nil, nil, fmt.Errorf("unprojected declared write %08x/%d", pe, width)
}
func (g *rskGraph4E23C0) apply(writes []rskWrite4E23C0) error {
	for _, w := range writes {
		data := rskHex4E23C0(w.Hex)
		address, f, e := g.address(w.Address, len(data))
		if e != nil {
			return e
		}
		if f.Accessible != f.PEWidth {
			return fmt.Errorf("input write must be fully mapped")
		}
		if f.Pointer {
			p, e := g.pointer(rskWord4E23C0(data))
			if e != nil {
				return e
			}
			data = rskPutWord4E23C0(uint64(uintptr(p)), f.NativeWidth)
		}
		ddiCopyIn4E0B30(address, data)
	}
	return nil
}
func (g *rskGraph4E23C0) afterRead(site, pe uint32, width int) error {
	g.Reads[site]++
	for i, s := range g.Input.AfterReads {
		if s.PC == site && s.Address == pe && s.Width == width && s.Occurrence == g.Reads[site] {
			if e := g.apply(s.Writes); e != nil {
				return e
			}
			g.UsedReads = append(g.UsedReads, i)
		}
	}
	return nil
}
func (g *rskGraph4E23C0) afterService(site uint32) error {
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
func (g *rskGraph4E23C0) snapshot() []byte {
	out := make([]byte, 0, len(g.Pointers)*g.Meta.Page)
	for _, p := range g.Pointers {
		out = append(out, ddiSnapshot4E0B30(p, g.Meta.Page)...)
	}
	return out
}
func (g *rskGraph4E23C0) close() {
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

// Projection of an authentic original full AFTER page into native field
// locations. Native canary bytes come exclusively from the saved BEFORE.
func rskExpectedAfter4E23C0(meta *rskGraphMetadata4E23C0, before []byte, pages map[uint32][]byte) ([]byte, error) {
	if len(before) != len(meta.Banks)*meta.Page {
		return nil, fmt.Errorf("native BEFORE extent")
	}
	out := append([]byte(nil), before...)
	for _, f := range meta.Fields {
		if f.Accessible == 0 {
			continue
		}
		bank := meta.Banks[f.Region]
		peBytes, ok := pages[bank.PE]
		if !ok || len(peBytes) != bank.Size {
			return nil, fmt.Errorf("original full AFTER page")
		}
		offset := int(f.PE - bank.PE)
		if offset < 0 || offset+f.Accessible > len(peBytes) {
			return nil, fmt.Errorf("original partial projection")
		}
		data := append([]byte(nil), peBytes[offset:offset+f.Accessible]...)
		if f.Pointer && f.Accessible == f.PEWidth {
			token := rskWord4E23C0(data)
			native := uint64(0)
			if token != 0 {
				var found bool
				native, found = meta.Identities[strconv.FormatUint(uint64(token), 10)]
				if !found {
					return nil, fmt.Errorf("unbound original pointer AFTER token")
				}
			}
			data = rskPutWord4E23C0(native, f.NativeWidth)
		}
		destination := f.Region*meta.Page + f.NativeOffset
		if destination < 0 || destination+len(data) > len(out) || f.NativeOffset+len(data) > meta.Page {
			return nil, fmt.Errorf("native expected projection bounds")
		}
		copy(out[destination:], data)
	}
	return out, nil
}

func rskReadHeader4E23C0(p []byte) (out [41]uint64, err error) {
	if len(p) < 41*8 {
		return out, fmt.Errorf("shared signal header extent")
	}
	for i := range out {
		out[i] = binary.LittleEndian.Uint64(p[i*8:])
	}
	return out, nil
}
