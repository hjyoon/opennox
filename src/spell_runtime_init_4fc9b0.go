package opennox

const (
	spellRuntimeMagicClassName4FC9B0     = "magicEntityClass"
	spellRuntimeMagicClassPE32Size4FC9B0 = uintptr(60)
	spellRuntimeMagicClassCapacity4FC9B0 = 64
	spellRuntimeCasterType4FC9B0         = "ImaginaryCaster"
	spellRuntimeMapCenter4FC9B0          = float32(2944)
)

var spellRuntimeObjectTypeNames4FC9B0 = [...]string{
	"Pixie",
	"MagicMissile",
	"SmallFist",
	"MediumFist",
	"LargeFist",
	"DeathBall",
	"Meteor",
}

type spellRuntimeInitHooks4FC9B0[Allocator, Object comparable] struct {
	initDurations          func() int32
	newMagicClass          func(name string, recordSize uintptr, capacity int) Allocator
	storeMagicClass        func(Allocator)
	newObjectByTypeID      func(name string) Object
	storeImaginaryCaster   func(Object)
	createObjectAt         func(object, owner Object, x, y float32)
	objectTypeIDByName     func(name string) uint32
	storeSpellObjectTypeID func(index int, value uint32)
}

// spellRuntimeInit4FC9B0 restores GAME.EXE 004FC9B0's observable order while
// allowing pointer-bearing allocator records to widen on native 64-bit hosts.
// The PE32 caller supplies recordSize 60; the native adapter supplies the
// widened MagicEntityClass size. Both pointer results are stored before their
// zero tests, and an allocation failure intentionally leaves earlier state in
// place.
func spellRuntimeInit4FC9B0[Allocator, Object comparable](
	recordSize uintptr,
	hooks spellRuntimeInitHooks4FC9B0[Allocator, Object],
) int32 {
	if hooks.initDurations() == 0 {
		return 0
	}

	magicClass := hooks.newMagicClass(
		spellRuntimeMagicClassName4FC9B0,
		recordSize,
		spellRuntimeMagicClassCapacity4FC9B0,
	)
	hooks.storeMagicClass(magicClass)
	var nilAllocator Allocator
	if magicClass == nilAllocator {
		return 0
	}

	caster := hooks.newObjectByTypeID(spellRuntimeCasterType4FC9B0)
	hooks.storeImaginaryCaster(caster)
	var nilObject Object
	if caster == nilObject {
		return 0
	}

	hooks.createObjectAt(
		caster,
		nilObject,
		spellRuntimeMapCenter4FC9B0,
		spellRuntimeMapCenter4FC9B0,
	)
	for i, name := range spellRuntimeObjectTypeNames4FC9B0 {
		typeID := hooks.objectTypeIDByName(name)
		hooks.storeSpellObjectTypeID(i, typeID)
	}
	return 1
}
