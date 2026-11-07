//go:build darwin && arm64

package legacy

// Input-only native graph. It never opens an original row or computes a
// floating result. Public-state and status field bytes are projected by the
// parent from authentic observations after the real native child terminates.
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

type sdWrite40A970 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type sdRegion40A970 struct {
	Name       string `json:"name"`
	Base       uint32 `json:"base"`
	Accessible int    `json:"accessible_bytes"`
	Initial    string `json:"initial_hex"`
}
type sdFPInput40A970 struct {
	Format string `json:"format"`
	Hex    string `json:"hex"`
}
type sdPacket40A970 struct {
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []sdWrite40A970   `json:"writes"`
	X87       *sdFPInput40A970  `json:"x87,omitempty"`
}
type sdSchedule40A970 struct {
	Target    uint32           `json:"target"`
	Argc      int              `json:"argc32"`
	Repeat    bool             `json:"repeat_last"`
	Responses []sdPacket40A970 `json:"responses"`
}
type sdStimulus40A970 struct {
	PC         uint32          `json:"pc"`
	Site       string          `json:"site"`
	Address    uint32          `json:"address"`
	Width      int             `json:"width"`
	Occurrence int             `json:"occurrence"`
	Writes     []sdWrite40A970 `json:"writes"`
}
type sdInput40A970 struct {
	Name          string                      `json:"name"`
	Entry         uint32                      `json:"entry"`
	Roles         map[string]uint32           `json:"roles"`
	Regions       []sdRegion40A970            `json:"regions"`
	InitialWords  map[string]uint32           `json:"initial_words"`
	Services      map[string]sdSchedule40A970 `json:"services"`
	AfterReads    []sdStimulus40A970          `json:"after_reads"`
	AfterWrites   []sdStimulus40A970          `json:"after_writes"`
	AfterServices []sdStimulus40A970          `json:"after_services"`
}
type sdField40A970 struct {
	Region                                         int
	PE                                             uint32
	PEWidth, NativeOffset, NativeWidth, Accessible int
}
type sdBank40A970 struct {
	Name                 string
	PE                   uint32
	Prefix, AnchorOffset int
	Native               uint64
}
type sdMeta40A970 struct {
	Page                        int
	Banks                       []sdBank40A970
	Fields                      []sdField40A970
	Identities                  map[string]uint64
	Opaque                      []uint32
	Cells                       [5]uint64
	GoPlayerOffset              uintptr
	CPlayerOffset, CPlayerWidth uintptr
}
type sdGraph40A970 struct {
	Input                               *sdInput40A970
	File                                *os.File
	Header                              unsafe.Pointer
	Pointers, OpaqueMaps                []unsafe.Pointer
	Tokens                              map[uint32]unsafe.Pointer
	Meta                                sdMeta40A970
	Reads, Writes                       map[string]int
	Services                            map[uint32]int
	UsedReads, UsedWrites, UsedServices []int
	UsedResponses                       map[string][]int
}

// Index0 is the independent frame input. Index1..4 are the four independent
// public-state DWORDs. Addresses identify original input locations only;
// diagnostic native cells are not aliases of production memmap storage.
var sdCellPE40A970 = [5]uint32{0x84ea04, 0x5d5354, 0x5d5364, 0x5d5328, 0x5d4b04}

func sdHex40A970(s string) []byte {
	p, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return p
}
func sdWord40A970(p []byte) uint32 {
	if len(p) > 4 {
		panic("DWORD transport width")
	}
	var out uint32
	for i, b := range p {
		out |= uint32(b) << uint(8*i)
	}
	return out
}
func sdBytes40A970(v uint32) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
}
func sdRegionAt40A970(in *sdInput40A970, address uint32) int {
	for i, r := range in.Regions {
		if uint64(address) >= uint64(r.Base) && uint64(address) < uint64(r.Base)+uint64(r.Accessible) {
			return i
		}
	}
	return -1
}
func sdAvailable40A970(in *sdInput40A970, address uint32, width int) int {
	i := sdRegionAt40A970(in, address)
	if i < 0 {
		return 0
	}
	n := int(uint64(in.Regions[i].Base) + uint64(in.Regions[i].Accessible) - uint64(address))
	if n > width {
		n = width
	}
	return n
}

func sdBuildGraph40A970(in *sdInput40A970, path string) (g *sdGraph40A970, err error) {
	g = &sdGraph40A970{Input: in, Tokens: map[uint32]unsafe.Pointer{}, Reads: map[string]int{}, Writes: map[string]int{}, Services: map[uint32]int{}, UsedResponses: map[string][]int{}}
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
	offset, width := suddenDeathFixtureCPlayerLayout40A970()
	named := unsafe.Offsetof(server.Player{}.Field3680)
	if offset != named || width != 4 {
		return nil, fmt.Errorf("actual compiled C/Go status DWORD layout")
	}
	g.Meta = sdMeta40A970{Page: page, Identities: map[string]uint64{}, GoPlayerOffset: named, CPlayerOffset: offset, CPlayerWidth: width}
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
	for i, r := range in.Regions {
		initial := sdHex40A970(r.Initial)
		if r.Accessible != len(initial) || r.Accessible < 3680 || r.Accessible > 3684 {
			return nil, fmt.Errorf("independent exact status prefix")
		}
		for j, old := range in.Regions {
			if j < i && uint64(r.Base) < uint64(old.Base)+uint64(old.Accessible) && uint64(old.Base) < uint64(r.Base)+uint64(r.Accessible) {
				return nil, fmt.Errorf("alias requires one physical region, not duplicate mappings")
			}
		}
		available := sdAvailable40A970(in, r.Base+3680, 4)
		cutoff := int(named) + available
		if cutoff > page {
			return nil, fmt.Errorf("native status cutoff")
		}
		p, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((i+1)*page), page)
		if e != nil {
			return nil, e
		}
		g.Pointers = append(g.Pointers, p)
		ddiFill4E0B30(p, page, byte(0xb0+i))
		if uintptr(p) <= 0x100000000 {
			return nil, fmt.Errorf("actual foreign high bank")
		}
		bank := sdBank40A970{r.Name, r.Base, r.Accessible, page - cutoff, uint64(uintptr(p))}
		g.Meta.Banks = append(g.Meta.Banks, bank)
		g.Tokens[r.Base] = unsafe.Add(p, bank.AnchorOffset)
		field := sdField40A970{i, r.Base + 3680, 4, bank.AnchorOffset + int(named), 4, available}
		g.Meta.Fields = append(g.Meta.Fields, field)
		if available > 0 {
			ddiCopyIn4E0B30(unsafe.Add(p, field.NativeOffset), initial[3680:3680+available])
		}
	}
	// Pointer words appear only at the declared external boundary. Each opaque
	// token gets a distinct protected native identity with no invented body.
	all := map[uint32]bool{}
	for _, v := range in.Roles {
		all[v] = true
	}
	for _, site := range []string{"0040a9ba", "0040a9df"} {
		s, ok := in.Services[site]
		if !ok {
			return nil, fmt.Errorf("declared pointer-return schedule")
		}
		for _, p := range s.Responses {
			all[p.Registers["eax"]] = true
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
		if sdRegionAt40A970(in, t) >= 0 {
			return nil, fmt.Errorf("unsupported readable interior pointer identity")
		}
		p := ddiMap4E0B30(page)
		if p == nil {
			return nil, fmt.Errorf("opaque foreign guard allocation")
		}
		g.OpaqueMaps = append(g.OpaqueMaps, p)
		g.Tokens[t] = unsafe.Add(p, page)
		g.Meta.Opaque = append(g.Meta.Opaque, t)
	}
	seen := map[unsafe.Pointer]bool{}
	for t, p := range g.Tokens {
		if p == nil || uintptr(p) <= 0x100000000 || seen[p] {
			return nil, fmt.Errorf("full native unique player identity")
		}
		seen[p] = true
		g.Meta.Identities[strconv.FormatUint(uint64(t), 10)] = uint64(uintptr(p))
	}
	// One shared physical bank holds five separated actual DWORD cells. Each
	// cell has adjacent canaries, and all unused bytes remain saved BEFORE.
	ri := len(in.Regions)
	p, e := rawDefaultDamageSharedMap4E0B30(int(g.File.Fd()), uint64((ri+1)*page), page)
	if e != nil {
		return nil, e
	}
	g.Pointers = append(g.Pointers, p)
	ddiFill4E0B30(p, page, 0xd7)
	if uintptr(p) <= 0x100000000 {
		return nil, fmt.Errorf("actual foreign state bank")
	}
	g.Meta.Banks = append(g.Meta.Banks, sdBank40A970{Name: "diagnostic-independent-five-DWORDs", Native: uint64(uintptr(p))})
	if len(in.InitialWords) != 5 {
		return nil, fmt.Errorf("exact five independently initialized DWORDs")
	}
	for field, pe := range sdCellPE40A970 {
		v, ok := in.InitialWords[fmt.Sprintf("%08x", pe)]
		if !ok {
			return nil, fmt.Errorf("missing independent DWORD input")
		}
		native := 128 + field*16
		address := unsafe.Add(p, native)
		ddiCopyIn4E0B30(address, sdBytes40A970(v))
		g.Meta.Cells[field] = uint64(uintptr(address))
		g.Meta.Fields = append(g.Meta.Fields, sdField40A970{ri, pe, 4, native, 4, 4})
	}
	for _, p := range append([]unsafe.Pointer{g.Header}, g.Pointers...) {
		if e = rawDefaultDamageSharedSync4E0B30(p, page); e != nil {
			return nil, e
		}
	}
	return g, nil
}

func (g *sdGraph40A970) pointer(token uint32) (unsafe.Pointer, error) {
	if token == 0 {
		return nil, nil
	}
	p, ok := g.Tokens[token]
	if !ok {
		return nil, fmt.Errorf("undeclared original PE token %08x", token)
	}
	return p, nil
}
func (g *sdGraph40A970) cells() (out [5]unsafe.Pointer) {
	p := g.Pointers[len(g.Pointers)-1]
	for i := range out {
		out[i] = unsafe.Add(p, 128+i*16)
	}
	return out
}
func (g *sdGraph40A970) address(pe uint32, width int) (unsafe.Pointer, *sdField40A970, error) {
	for _, field := range g.Meta.Fields {
		if pe < field.PE {
			continue
		}
		delta := uint64(pe) - uint64(field.PE)
		if width > 0 && delta+uint64(width) <= uint64(field.PEWidth) {
			if delta+uint64(width) > uint64(field.Accessible) {
				return nil, nil, fmt.Errorf("input effect must be fully mapped")
			}
			field.NativeOffset += int(delta)
			field.PE = pe
			field.PEWidth = width
			field.NativeWidth = width
			field.Accessible = width
			return unsafe.Add(g.Pointers[field.Region], field.NativeOffset), &field, nil
		}
	}
	return nil, nil, fmt.Errorf("unprojected independent write %08x/%d", pe, width)
}
func (g *sdGraph40A970) apply(writes []sdWrite40A970) error {
	for _, w := range writes {
		data := sdHex40A970(w.Hex)
		p, _, e := g.address(w.Address, len(data))
		if e != nil {
			return e
		}
		ddiCopyIn4E0B30(p, data)
	}
	return nil
}
func (g *sdGraph40A970) afterOperation(kind string, site, pe uint32, width int) error {
	var counts map[string]int
	var stimuli []sdStimulus40A970
	var used *[]int
	switch kind {
	case "after_reads":
		counts, stimuli, used = g.Reads, g.Input.AfterReads, &g.UsedReads
	case "after_writes":
		counts, stimuli, used = g.Writes, g.Input.AfterWrites, &g.UsedWrites
	default:
		return fmt.Errorf("declared completed-operation effect kind")
	}
	key := fmt.Sprintf("%08x/%08x/%d", site, pe, width)
	counts[key]++
	for i, s := range stimuli {
		if s.PC == site && s.Address == pe && s.Width == width && s.Occurrence == counts[key] {
			for _, j := range *used {
				if j == i {
					return fmt.Errorf("independent stimulus used twice")
				}
			}
			if e := g.apply(s.Writes); e != nil {
				return e
			}
			*used = append(*used, i)
		}
	}
	return nil
}
func (g *sdGraph40A970) afterRead(site, pe uint32, width int) error {
	return g.afterOperation("after_reads", site, pe, width)
}
func (g *sdGraph40A970) afterWrite(site, pe uint32, width int) error {
	return g.afterOperation("after_writes", site, pe, width)
}
func (g *sdGraph40A970) afterService(site uint32) error {
	g.Services[site]++
	for i, s := range g.Input.AfterServices {
		if s.Site == fmt.Sprintf("%08x", site) && s.Occurrence == g.Services[site] {
			for _, j := range g.UsedServices {
				if j == i {
					return fmt.Errorf("independent service stimulus used twice")
				}
			}
			if e := g.apply(s.Writes); e != nil {
				return e
			}
			g.UsedServices = append(g.UsedServices, i)
		}
	}
	return nil
}
func (g *sdGraph40A970) snapshot() []byte {
	var out []byte
	for _, p := range g.Pointers {
		out = append(out, ddiSnapshot4E0B30(p, g.Meta.Page)...)
	}
	return out
}
func (g *sdGraph40A970) close() {
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

// Authentic selected AFTER field bytes are read by the parent from actual
// original full-page observations. Native unselected bytes use saved BEFORE.
func sdExpectedAfter40A970(meta *sdMeta40A970, before []byte, fields map[uint32][]byte) ([]byte, error) {
	if len(before) != len(meta.Banks)*meta.Page {
		return nil, fmt.Errorf("whole native BEFORE extent")
	}
	out := append([]byte(nil), before...)
	for _, f := range meta.Fields {
		if f.Accessible == 0 {
			continue
		}
		p, ok := fields[f.PE]
		if !ok || len(p) != f.Accessible {
			return nil, fmt.Errorf("authentic original AFTER field span")
		}
		offset := f.Region*meta.Page + f.NativeOffset
		if f.NativeOffset < 0 || f.NativeOffset+len(p) > meta.Page {
			return nil, fmt.Errorf("native field AFTER projection")
		}
		copy(out[offset:], p)
	}
	return out, nil
}

// This validates actual public-state/status events only. Numeric store32 and
// convert32 private diagnostics have no native memory identity in this graph.
func sdOperand40A970(g *sdGraph40A970, e suddenDeathNativeOperand40A970) (uint32, int, error) {
	type descriptor struct {
		field int
		write bool
	}
	sites := map[uint32]descriptor{0x40a970: {0, false}, 0x40a97b: {1, true}, 0x40a980: {2, true}, 0x40a99c: {3, true}, 0x40a9b5: {4, true}, 0x40a9c5: {5, false}}
	d, ok := sites[e.Site]
	if !ok || e.Field != d.field || e.Write != d.write || e.Width != 4 {
		return 0, 0, fmt.Errorf("static DWORD public/status site/type")
	}
	if e.Complete && !e.ValueValid {
		return 0, 0, fmt.Errorf("completed actual value must be valid")
	}
	if !e.Complete && !e.Write && e.ValueValid {
		return 0, 0, fmt.Errorf("read-before cannot invent a value")
	}
	if e.Field < 5 {
		address := g.cells()[e.Field]
		if !e.AddressValid || e.Address != address {
			return 0, 0, fmt.Errorf("actual independent diagnostic DWORD owner")
		}
		return sdCellPE40A970[e.Field], 4, nil
	}
	if e.Offset != g.Meta.GoPlayerOffset {
		return 0, 0, fmt.Errorf("actual named player status offset")
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
		return 0, 0, fmt.Errorf("actual player base identity")
	}
	if e.Base == nil {
		if e.Address != nil {
			return 0, 0, fmt.Errorf("no low unsafe.Pointer exposure for nil")
		}
	} else if !e.AddressValid || uintptr(e.Address) != uintptr(e.Base)+e.Offset {
		return 0, 0, fmt.Errorf("actual full foreign status address")
	}
	return token + 3680, 4, nil
}

// Packet payloads are raw inputs. They are not converted to a host float,
// normalized, rounded, or replaced with values from an original observation.
func sdRawFPInput40A970(p *sdFPInput40A970) (string, [10]byte, error) {
	var out [10]byte
	if p == nil {
		return "", out, fmt.Errorf("missing declared raw FP input")
	}
	bits := sdHex40A970(p.Hex)
	if (p.Format != "f64" || len(bits) != 8) && (p.Format != "ext80" || len(bits) != 10) {
		return "", out, fmt.Errorf("literal raw floating input format/width")
	}
	copy(out[:], bits)
	return p.Format, out, nil
}
