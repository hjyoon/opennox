package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type DurSpell struct {
	ID       uint16         // 0, 0
	_        uint16         // 0, 2
	Spell    uint32         // 1, 4
	Level    uint32         // 2, 8
	Obj12    *Object        // 3, 12
	Caster16 *Object        // 4, 16
	Flag20   uint32         // 5, 20
	Obj24    *Object        // 6, 24
	Pos      types.Pointf   // 7, 28
	Field36  uint32         // 9, 36
	Field40  uint32         // 10, 40
	Field44  uint32         // 11, 44
	Target48 *Object        // 12, 48
	Pos2     types.Pointf   // 13, 52
	Frame60  uint32         // 15, 60
	Frame64  uint32         // 16, 64
	Frame68  uint32         // 17, 68
	Field72  int32          // 18, 72
	Field76  uintptr        // 19, 76
	Field80  uint32         // 20, 80
	Field84  uint32         // 21, 84
	Flags88  uint32         // 22, 88
	Create   unsafe.Pointer // 23, 92
	Update   unsafe.Pointer // 24, 96
	Destroy  unsafe.Pointer // 25, 100
	Sub104   *DurSpell      // 26, 104
	Sub108   *DurSpell      // 27, 108
	Prev     *DurSpell      // 28, 112
	Next     *DurSpell      // 29, 116
}

func (sp *DurSpell) C() unsafe.Pointer {
	return unsafe.Pointer(sp)
}

type SpellsDuration struct {
	s      *Server
	alloc  alloc.ClassT[DurSpell]
	List   *DurSpell
	lastID uint16
}

func (sp *SpellsDuration) init(s *Server) {
	sp.s = s
}

func (sp *SpellsDuration) Init() bool {
	return sp.SpellCreateDurations4FE850() != 0
}

func (sp *SpellsDuration) Free() {
	sp.SpellFreeDurations4FE880()
}

func (sp *SpellsDuration) NewRaw() *DurSpell {
	return sp.SpellDurationNew4FE950()
}

func (sp *SpellsDuration) NewLightningSub(src *DurSpell, from, to *Object) {
	p := sp.NewRaw()
	if p == nil {
		return
	}
	p.Target48 = to
	p.Caster16 = from
	p.Spell = uint32(spell.SPELL_CHAIN_LIGHTNING_BOLT)
	p.Sub108 = nil
	p.Sub104 = nil
	p.Prev = nil
	p.Next = src.Sub108
	if sub := src.Sub108; sub != nil {
		sub.Prev = p
	}
	src.Sub108 = p
}

func (sp *SpellsDuration) FreeRecursive(p *DurSpell) {
	sp.SpellDurationFreeRecursive4FE980(p)
}

func (sp *SpellsDuration) Add(p *DurSpell) {
	if sp.List != nil {
		sp.List.Prev = p
	}
	p.Prev = nil
	p.Next = sp.List
	sp.List = p
}

func (sp *SpellsDuration) Unlink(p *DurSpell) {
	sp.SpellDurationUnlink4FE900(p)
}

func (sp *SpellsDuration) Sub4FEE50(a1 spell.ID, a2 *Object) bool {
	for it := sp.List; it != nil; it = it.Next {
		if it.Flag20 == 0 && spell.ID(it.Spell) == a1 && it.Caster16 == a2 && it.Flags88&0x1 == 0 {
			return true
		}
	}
	return false
}

func (sp *SpellsDuration) CancelOffensiveFor(u *Object) {
	var next *DurSpell
	for it := sp.List; it != nil; it = next {
		next = it.Next
		if it.Caster16 == u && sp.s.Spells.Flags(spell.ID(it.Spell)).Has(things.SpellOffensive) {
			sp.CancelSpell(it)
		}
	}
}

func (sp *SpellsDuration) CancelFor(sid spell.ID, obj Obj) {
	var next *DurSpell
	for it := sp.List; it != nil; it = next {
		sid2 := spell.ID(it.Spell)
		next = it.Next
		if sid2 == sid && it.Caster16 == ToObject(obj) || SpellIsSummon(sid) && SpellIsSummon(sid2) && it.Caster16 == ToObject(obj) {
			sp.CancelSpell(it)
		}
	}
}

func (sp *SpellsDuration) CancelSpell(sd *DurSpell) {
	_ = sp.SpellDurationCancel4FE9D0(sd)
}
