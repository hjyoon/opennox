package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"
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
	sp.SpellDurationInsert4FED40(p)
}

func (sp *SpellsDuration) Unlink(p *DurSpell) {
	sp.SpellDurationUnlink4FE900(p)
}

// Sub4FEE50 is retained as a source-compatible bool view of the restored
// canonical dword predicate. Production callers use the descriptive method.
func (sp *SpellsDuration) Sub4FEE50(a1 spell.ID, a2 *Object) bool {
	return sp.SpellDurationDuplicate4FEE50(int32(a1), a2) != 0
}

func (sp *SpellsDuration) CancelOffensiveFor(u *Object) {
	sp.SpellDurationCancelOffensive4FF310(u)
}

func (sp *SpellsDuration) CancelFor(sid spell.ID, obj Obj) {
	sp.SpellCancelDurSpell4FEB10(int32(sid), ToObject(obj))
}

func (sp *SpellsDuration) CancelSpell(sd *DurSpell) {
	_ = sp.SpellDurationCancel4FE9D0(sd)
}
