package server

import "github.com/opennox/libs/types"

type greaterHealCreateHooks52F220[Record, Object comparable] struct {
	loadMode     func(Record) uint32
	loadCaster   func(Record) Object
	loadSpell    func(Record) uint32
	loadAim      func(Record) *types.Pointf
	spellFlags   func(uint32) uint32
	searchTarget func(*types.Pointf, Object, uint32, float32, int, Object) Object
	adjustHP     func(Object, int32)
	storeTarget  func(Record, Object)
	isEnemy      func(Object, Object) bool
	startRay     func(Record)
	noTarget     func(Object)
}

// spellGreaterHealCreate52F220 is GAME.EXE 0052F220's create callback. Glyph
// casts heal once and cancel; ordinary casts find a friendly target and start
// the duration ray. Record fields are never decoded as PE32 dwords.
func spellGreaterHealCreate52F220[Record, Object comparable](record Record, h greaterHealCreateHooks52F220[Record, Object]) int32 {
	var nilObject Object
	mode := h.loadMode(record)
	caster := h.loadCaster(record)
	if caster == nilObject && mode == 0 {
		return 1
	}
	if caster == nilObject || mode != 0 {
		flags := h.spellFlags(h.loadSpell(record))
		target := h.searchTarget(h.loadAim(record), nilObject, flags, 400, 1, nilObject)
		if target != nilObject {
			h.adjustHP(target, 20)
		}
		return 1
	}

	caster = h.loadCaster(record)
	flags := h.spellFlags(h.loadSpell(record))
	target := h.searchTarget(h.loadAim(record), caster, flags, 400, 1, caster)
	h.storeTarget(record, target)
	if target == nilObject {
		h.noTarget(h.loadCaster(record))
		return 1
	}
	if h.isEnemy(h.loadCaster(record), target) {
		return 1
	}
	h.startRay(record)
	return 0
}

type greaterHealUpdateHooks52F2E0[Record, Object comparable] struct {
	loadTarget    func(Record) Object
	loadFlags     func(Object) uint32
	loadCaster    func(Record) Object
	testBuff      func(Object, int32) int32
	canInteract   func(Object, Object) bool
	oldMana       func(Object) uint16
	loadClass     func(Object) uint8
	positionDelta func(Object, Record) int32
	wasDamaged    func(Object) bool
	maxHP         func(Object) uint16
	getHP         func(Object) uint16
	loadFraction  func(Record) float32
	loadLevel     func(Record) uint32
	coefficient   func(uint32) float32
	playerClass   func(Object) uint8
	classHealth   func(uint8) float32
	storeFraction func(Record, float32)
	adjustHP      func(Object, int32)
	manaSub       func(Object, int32)
}

// spellGreaterHealUpdate52F2E0 follows GAME.EXE 0052F2E0's live caster and
// target reads. The health fraction is a float32 bit pattern in Field72.
func spellGreaterHealUpdate52F2E0[Record, Object comparable](record Record, h greaterHealUpdateHooks52F2E0[Record, Object]) int32 {
	var nilObject Object
	target := h.loadTarget(record)
	if target == nilObject || h.loadFlags(target)&0x8020 != 0 {
		return 1
	}
	caster := h.loadCaster(record)
	if caster != nilObject && h.testBuff(caster, 8) != 0 {
		return 1
	}
	if !h.canInteract(h.loadCaster(record), h.loadTarget(record)) || h.oldMana(h.loadCaster(record)) == 0 {
		return 1
	}
	caster = h.loadCaster(record)
	if h.loadClass(caster)&2 != 0 && h.positionDelta(caster, record) != 0 {
		return 1
	}
	if h.wasDamaged(h.loadCaster(record)) {
		return 1
	}
	if h.maxHP(h.loadTarget(record)) == h.getHP(h.loadTarget(record)) {
		return 1
	}
	caster = h.loadCaster(record)
	value := h.loadFraction(record) + h.coefficient(h.loadLevel(record))
	if caster != nilObject && h.loadClass(caster)&4 != 0 {
		class := h.playerClass(caster)
		if class <= 2 {
			value = float32(float64(value) * float64(h.classHealth(class)))
		}
	}
	amount := spellDurationRoundNearestEven(value)
	h.storeFraction(record, float32(float64(value)-float64(amount)))
	h.adjustHP(h.loadTarget(record), amount)
	h.manaSub(h.loadCaster(record), 1)
	return 0
}
