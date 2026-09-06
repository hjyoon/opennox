package server

const (
	spellDurationCreateInactiveMask4FEBA0 = uint32(0x8020)
	spellDurationCreateTargeted4FEBA0     = uint32(4)
	spellDurationCreatePlasma4FEBA0       = int32(59)
	spellDurationCreateChain4FEBA0        = int32(43)
)

// SpellDurationCreateHooks4FEBA0 exposes every observable argument, cache,
// object, record, and callback access in GAME.EXE 004FEBA0. Comparable tokens
// model native null values without imposing the original PE32 pointer width.
type SpellDurationCreateHooks4FEBA0[Record, Object, AcceptArg, Callback comparable] struct {
	LoadGlyphCache  func() uint32
	LookupGlyph     func() uint32
	StoreGlyphCache func(uint32)

	LoadFourthArg   func() Object
	LoadThirdArg    func() Object
	LoadObjectFlags func(Object) uint32
	LoadObjectType  func(Object) uint16
	LoadSpellArg    func() int32

	FindDuplicate func(int32, Object) int32
	CancelFor     func(int32, Object)
	BeforeCreate  func()
	NewRecord     func() Record

	LoadLevelArg   func() int32
	LoadSecondArg  func() Object
	StoreSpell     func(Record, uint32)
	StoreLevel     func(Record, uint32)
	StoreCaster    func(Record, Object)
	StoreSource    func(Record, Object)
	StoreSub108    func(Record, Record)
	StoreSub104    func(Record, Record)
	StoreMode      func(Record, uint32)
	StoreAnchor    func(Record, Object)
	LoadPositionX  func(Object) float32
	StorePositionX func(Record, float32)
	LoadPositionY  func(Object) float32

	LoadAcceptArg  func() AcceptArg
	StoreField36   func(Record, uint32)
	StorePositionY func(Record, float32)
	LoadTarget     func(AcceptArg) Object
	LoadCreateArg  func() Callback
	StoreTarget    func(Record, Object)
	LoadAcceptX    func(AcceptArg) float32
	StoreAcceptX   func(Record, float32)
	LoadUpdateArg  func() Callback
	LoadAcceptY    func(AcceptArg) float32
	StoreCreate    func(Record, Callback)
	StoreAcceptY   func(Record, float32)
	LoadDestroyArg func() Callback
	StoreUpdate    func(Record, Callback)
	StoreDestroy   func(Record, Callback)

	LoadFrame         func() uint32
	LoadDurationArg   func() uint32
	StoreFrame60      func(Record, uint32)
	StoreFrame64      func(Record, uint32)
	StoreFlagsLowByte func(Record, byte)
	StoreFrame68      func(Record, uint32)
	AddRecord         func(Record)

	SpellHasFlags func(int32, uint32) int32
	SpellAudio    func(int32, int32) int32
	AudioEvent    func(int32, Object, int32, int32)
	CallCreate    func(Callback, Record) int32
	CancelSpell   func(Record)
}

// SpellDurationCreate4FEBA0 preserves GAME.EXE 004FEBA0's exact branch,
// access, write, and callback order. A missing or inactive caster is rejected
// only when a non-Glyph fourth object is present; a Glyph may therefore carry
// a nil caster. Plasma and Chain Lightning return early only when the duplicate
// probe returns exactly one. The record's flag field receives a low-byte-only
// clear, all three frame values are loaded independently, and any nonzero
// create result cancels the newly linked record.
func SpellDurationCreate4FEBA0[Record, Object, AcceptArg, Callback comparable](
	h SpellDurationCreateHooks4FEBA0[Record, Object, AcceptArg, Callback],
) int32 {
	glyph := h.LoadGlyphCache()
	if glyph == 0 {
		glyph = h.LookupGlyph()
		h.StoreGlyphCache(glyph)
	}

	fourth := h.LoadFourthArg()
	third := h.LoadThirdArg()
	var nilObject Object
	if third == nilObject || h.LoadObjectFlags(third)&spellDurationCreateInactiveMask4FEBA0 != 0 {
		if fourth != nilObject && uint32(h.LoadObjectType(fourth)) != glyph {
			return 0
		}
	}

	spellID := h.LoadSpellArg()
	if third != nilObject {
		if (spellID == spellDurationCreatePlasma4FEBA0 || spellID == spellDurationCreateChain4FEBA0) &&
			h.FindDuplicate(spellID, third) == 1 {
			return 1
		}
		h.CancelFor(spellID, third)
	}

	h.BeforeCreate()
	record := h.NewRecord()
	var nilRecord Record
	if record == nilRecord {
		return 0
	}

	level := h.LoadLevelArg()
	second := h.LoadSecondArg()
	h.StoreSpell(record, uint32(spellID))
	h.StoreLevel(record, uint32(level))
	h.StoreCaster(record, third)
	h.StoreSource(record, second)
	h.StoreSub108(record, nilRecord)
	h.StoreSub104(record, nilRecord)

	var positionY float32
	isGlyph := false
	if fourth != nilObject {
		liveGlyph := h.LoadGlyphCache()
		isGlyph = uint32(h.LoadObjectType(fourth)) == liveGlyph
	}
	if isGlyph {
		h.StoreMode(record, 1)
		h.StoreAnchor(record, fourth)
		h.StorePositionX(record, h.LoadPositionX(fourth))
		positionY = h.LoadPositionY(fourth)
	} else {
		h.StoreMode(record, 0)
		h.StoreAnchor(record, nilObject)
		h.StorePositionX(record, h.LoadPositionX(third))
		positionY = h.LoadPositionY(third)
	}

	arg := h.LoadAcceptArg()
	h.StoreField36(record, 0)
	h.StorePositionY(record, positionY)
	target := h.LoadTarget(arg)
	create := h.LoadCreateArg()
	h.StoreTarget(record, target)
	h.StoreAcceptX(record, h.LoadAcceptX(arg))
	update := h.LoadUpdateArg()
	acceptY := h.LoadAcceptY(arg)
	h.StoreCreate(record, create)
	h.StoreAcceptY(record, acceptY)
	destroy := h.LoadDestroyArg()
	h.StoreUpdate(record, update)
	h.StoreDestroy(record, destroy)

	frame60 := h.LoadFrame()
	duration := h.LoadDurationArg()
	h.StoreFrame60(record, frame60)
	h.StoreFrame64(record, h.LoadFrame())
	frame68 := h.LoadFrame() + duration
	h.StoreFlagsLowByte(record, 0)
	h.StoreFrame68(record, frame68)
	h.AddRecord(record)

	targeted := int32(0)
	if h.SpellHasFlags(spellID, spellDurationCreateTargeted4FEBA0) != 0 {
		targeted = 1
	}
	audio := h.SpellAudio(spellID, targeted)
	h.AudioEvent(audio, third, 0, 0)
	var nilCallback Callback
	if create == nilCallback || h.CallCreate(create, record) == 0 {
		return 1
	}
	h.CancelSpell(record)
	return 0
}
