//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"unsafe"
)

// Mutable projection of explicit INPUT effects. No parent writes, branches or
// original output bytes are replayed into this graph. Native engine comparison
// must observe independently executed parent writes, never call apply as them.
type rddProjection4E0B30 struct {
	Graph      *ddiGraph4E0B30
	Canonical  map[string][]byte
	ClockIndex int
	Clock      [4]byte
}

func rddNewProjection4E0B30(g *ddiGraph4E0B30, in rddInput4E0B30) (*rddProjection4E0B30, error) {
	p := &rddProjection4E0B30{Graph: g, Canonical: map[string][]byte{}}
	for name, a := range g.Mirrors {
		p.Canonical[name] = ddiSnapshot4E0B30(unsafe.Add(a.Map, a.Base), 4096)
	}
	clockSet := false
	for _, init := range in.ImageInitializers {
		data, e := hex.DecodeString(init.Hex)
		if e != nil {
			return nil, e
		}
		if init.Address != 0x84ea04 || len(data) != 4 || clockSet {
			return nil, fmt.Errorf("declared frame-only mutable image input domain")
		}
		copy(p.Clock[:], data)
		clockSet = true
	}
	if !clockSet {
		return nil, fmt.Errorf("frame initializer absent")
	}
	index, e := g.alloc("consumer-input-clock", 256, 0x9a)
	if e != nil {
		return nil, e
	}
	p.ClockIndex = index
	a := &g.Allocations[index]
	ddiCopyIn4E0B30(g.base(index), p.Clock[:])
	copy(a.Expected[a.Base:], p.Clock[:])
	return p, nil
}
func (p *rddProjection4E0B30) resolve(effect rddEffect4E0B30, args []uint32) (uint64, []byte, error) {
	n := 0
	address := uint64(0)
	if effect.Address != nil {
		address = uint64(*effect.Address)
		n++
	}
	if effect.Arg != nil {
		if *effect.Arg < 0 || *effect.Arg >= len(args) {
			return 0, nil, fmt.Errorf("effect arg outside actual transport frame")
		}
		address = uint64(args[*effect.Arg])
		n++
	}
	if effect.Role != "" {
		word, ok := p.Graph.Case.Input.Roles[effect.Role]
		if !ok {
			return 0, nil, fmt.Errorf("unknown effect role")
		}
		address = uint64(word)
		n++
	}
	if n != 1 {
		return 0, nil, fmt.Errorf("effect selector count")
	}
	next := int64(address) + int64(effect.Offset)
	if next < 0 || next > 0xffffffff {
		return 0, nil, fmt.Errorf("effect uint32 address range")
	}
	data, e := hex.DecodeString(effect.Hex)
	if e != nil || len(data) == 0 {
		return 0, nil, fmt.Errorf("input effect bytes: %v", e)
	}
	return uint64(next), data, nil
}
func (p *rddProjection4E0B30) apply(effect rddEffect4E0B30, args []uint32) error {
	address, data, e := p.resolve(effect, args)
	if e != nil {
		return e
	}
	g := p.Graph
	if address >= 0x84ea04 && address+uint64(len(data)) <= 0x84ea08 {
		offset := int(address - 0x84ea04)
		copy(p.Clock[offset:], data)
		a := &g.Allocations[p.ClockIndex]
		ddiCopyIn4E0B30(unsafe.Add(g.base(p.ClockIndex), offset), data)
		copy(a.Expected[a.Base+offset:], data)
		return p.check()
	}
	// The frozen input domain targets named field storage. Reject rather than
	// silently copy unrelated gaps or widen a partial field's accessible extent.
	for _, ef := range g.Fields {
		d := g.Definitions[ef.Field]
		start := uint64(g.Case.Input.Roles[ef.Role]) + uint64(d.Offset)
		if address < start || address+uint64(len(data)) > start+uint64(d.Width) {
			continue
		}
		index := ddiRoleIndex4E0B30(ef.Role)
		a := &g.Allocations[index]
		offset := int(g.Offsets[d.ID]) + int(address-start)
		native := data
		if d.pointer() {
			if address != start || len(data) != 4 {
				return fmt.Errorf("partial pointer input mutation requires separate explicit domain")
			}
			word := uint32(ddiUnsigned4E0B30(data))
			token, ok := g.Tokens[word]
			if !ok {
				return fmt.Errorf("unmapped explicit pointer input %#x", word)
			}
			if a.Base+offset+8 > g.Page {
				return fmt.Errorf("pointer mutation crosses original native cutoff")
			}
			if !ddiPutPointer4E0B30(g.base(index), d.ID, g.Offsets[d.ID], token.Pointer) {
				return fmt.Errorf("actual typed C pointer input transport")
			}
			native = ddiBytes4E0B30(uint64(token.Bits), 8)
		} else {
			if a.Base+offset+len(data) > g.Page {
				return fmt.Errorf("scalar mutation crosses original native cutoff")
			}
			ddiCopyIn4E0B30(unsafe.Add(g.base(index), offset), data)
		}
		copy(a.Expected[a.Base+offset:], native)
		var region *ddiRegion4E0B30
		for n := range g.Case.Input.Regions {
			r := &g.Case.Input.Regions[n]
			if r.Name == ef.Role {
				region = r
				break
			}
		}
		if region == nil {
			return fmt.Errorf("effect role has no readable input region")
		}
		at := int(address - uint64(region.Address))
		mirror := g.Mirrors[ef.Role]
		if at < 0 || at+len(data) > 4096 {
			return fmt.Errorf("canonical input effect extent")
		}
		copy(p.Canonical[ef.Role][at:], data)
		ddiCopyIn4E0B30(unsafe.Add(mirror.Map, mirror.Base+at), data)
		copy(mirror.Expected[mirror.Base+at:], data)
		g.Mirrors[ef.Role] = mirror
		return p.check()
	}
	return fmt.Errorf("input effect lacks exact selected field projection %#x/%d", address, len(data))
}
func (p *rddProjection4E0B30) check() error {
	g := p.Graph
	if e := g.checkCanaries(); e != nil {
		return e
	}
	// Independent reverse map: only pointer-semantic fields are relocated.
	// sound_id remains scalar even when its DWORD equals a role address.
	for _, ef := range g.Fields {
		d := g.Definitions[ef.Field]
		regionFound := false
		var region ddiRegion4E0B30
		for _, r := range g.Case.Input.Regions {
			if r.Name == ef.Role {
				region = r
				regionFound = true
				break
			}
		}
		if !regionFound {
			continue
		}
		start := int(g.Case.Input.Roles[ef.Role]-region.Address) + d.Offset
		originalCount := len(ef.Hex) / 2
		if originalCount == 0 {
			continue
		}
		if start < 0 || start+originalCount > 4096 {
			return fmt.Errorf("projection canonical extent")
		}
		want := p.Canonical[ef.Role][start : start+originalCount]
		index := ddiRoleIndex4E0B30(ef.Role)
		actual := ddiSnapshot4E0B30(unsafe.Add(g.base(index), g.Offsets[d.ID]), originalCount)
		if d.pointer() && originalCount == 4 {
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
				return fmt.Errorf("reverse projection unknown full native pointer")
			}
		}
		if !bytes.Equal(actual, want) {
			return fmt.Errorf("actual C-owned mutable projection %s/field%d", ef.Role, d.ID)
		}
	}
	if !bytes.Equal(ddiSnapshot4E0B30(g.base(p.ClockIndex), 4), p.Clock[:]) {
		return fmt.Errorf("actual frame transport bytes")
	}
	return nil
}
