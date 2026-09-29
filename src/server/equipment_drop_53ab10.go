package server

const (
	equipmentDropCoopFlag53AB10    = uint32(0x00000800)
	equipmentDropQuestFlag53AB10   = uint32(0x00001000)
	equipmentDropDecayFlag53AB10   = uint32(0x00000002)
	equipmentDropSeconds53AB10     = uint32(25)
	weaponDropWandClass53AAB0      = uint32(0x00001000)
	weaponDropMetalMaterial53AAB0  = uint16(0x0010)
	weaponDropWoodMaterial53AAB0   = uint16(0x0008)
	weaponDropWandAudio53AAB0      = uint32(831)
	weaponDropMetalAudio53AAB0     = uint32(843)
	weaponDropWoodAudio53AAB0      = uint32(845)
	armorDropMetalMaterial53EAE0   = uint16(0x0010)
	armorDropWoodMaterial53EAE0    = uint16(0x0008)
	armorDropLeatherMaterial53EAE0 = uint16(0x0004)
	armorDropClothMaterial53EAE0   = uint16(0x0002)
	armorDropShoesSubclass53EAE0   = uint32(0x00000020)
	armorDropMetalAudio53EAE0      = uint32(805)
	armorDropWoodAudio53EAE0       = uint32(811)
	armorDropLeatherAudio53EAE0    = uint32(808)
	armorDropClothAudio53EAE0      = uint32(814)
	armorDropShoesAudio53EAE0      = uint32(817)
)

// equipmentDropHooks53AB10 exposes the common GAME.EXE 0053AB10 and
// 0053EB70 argument-load and callback order. The point, owner, and item are
// cached before DefaultDrop. All mode and decay-state reads remain live after
// the type-specific drop sound callback.
type equipmentDropHooks53AB10[O, P comparable] struct {
	loadPointArg func() P
	loadOwnerArg func() O
	loadItemArg  func() O

	defaultDrop   func(O, O, P) int32
	dropSound     func(O)
	gameFlag      func(uint32) int32
	serverSubFlag func(uint32) int32
	loadGameFPS   func() uint32
	setDecay      func(O, uint32)
}

// equipmentDrop53AB10 preserves the body shared by GAME.EXE 0053AB10 and
// 0053EB70. DefaultDrop must return exactly one; every other value is a
// failure. The x86 LEA sequence computes 25*FPS modulo 2^32.
func equipmentDrop53AB10[O, P comparable](hooks equipmentDropHooks53AB10[O, P]) int32 {
	point := hooks.loadPointArg()
	owner := hooks.loadOwnerArg()
	item := hooks.loadItemArg()
	if hooks.defaultDrop(owner, item, point) != 1 {
		return 0
	}

	hooks.dropSound(item)
	if hooks.gameFlag(equipmentDropCoopFlag53AB10) != 0 {
		return 1
	}
	if hooks.gameFlag(equipmentDropQuestFlag53AB10) != 0 {
		return 1
	}
	if hooks.serverSubFlag(equipmentDropDecayFlag53AB10) == 0 {
		return 1
	}
	hooks.setDecay(item, hooks.loadGameFPS()*equipmentDropSeconds53AB10)
	return 1
}

func weaponDrop53AB10[O, P comparable](hooks equipmentDropHooks53AB10[O, P]) int32 {
	return equipmentDrop53AB10(hooks)
}

func armorDrop53EB70[O, P comparable](hooks equipmentDropHooks53AB10[O, P]) int32 {
	return equipmentDrop53AB10(hooks)
}

type weaponDropSoundHooks53AAB0[O comparable] struct {
	loadItemArg  func() O
	loadClass    func(O) uint32
	loadMaterial func(O) uint16
	audio        func(uint32, O, int32, uint32)
}

// weaponDropSound53AAB0 preserves GAME.EXE 0053AAB0. Wands take precedence
// over material, and metal takes precedence over wood when bits overlap.
func weaponDropSound53AAB0[O comparable](hooks weaponDropSoundHooks53AAB0[O]) {
	item := hooks.loadItemArg()
	var nilObject O
	if item == nilObject {
		return
	}
	if hooks.loadClass(item)&weaponDropWandClass53AAB0 != 0 {
		hooks.audio(weaponDropWandAudio53AAB0, item, 0, 0)
		return
	}
	material := hooks.loadMaterial(item)
	if material&weaponDropMetalMaterial53AAB0 != 0 {
		hooks.audio(weaponDropMetalAudio53AAB0, item, 0, 0)
		return
	}
	if material&weaponDropWoodMaterial53AAB0 != 0 {
		hooks.audio(weaponDropWoodAudio53AAB0, item, 0, 0)
	}
}

type armorDropSoundHooks53EAE0[O comparable] struct {
	loadItemArg  func() O
	loadMaterial func(O) uint16
	loadSubClass func(O) uint32
	audio        func(uint32, O, int32, uint32)
}

// armorDropSound53EAE0 preserves GAME.EXE 0053EAE0. Its material priority is
// metal, wood, leather, then cloth; the shoes subclass is read only for cloth.
func armorDropSound53EAE0[O comparable](hooks armorDropSoundHooks53EAE0[O]) {
	item := hooks.loadItemArg()
	var nilObject O
	if item == nilObject {
		return
	}
	material := hooks.loadMaterial(item)
	switch {
	case material&armorDropMetalMaterial53EAE0 != 0:
		hooks.audio(armorDropMetalAudio53EAE0, item, 0, 0)
	case material&armorDropWoodMaterial53EAE0 != 0:
		hooks.audio(armorDropWoodAudio53EAE0, item, 0, 0)
	case material&armorDropLeatherMaterial53EAE0 != 0:
		hooks.audio(armorDropLeatherAudio53EAE0, item, 0, 0)
	case material&armorDropClothMaterial53EAE0 != 0:
		if hooks.loadSubClass(item)&armorDropShoesSubclass53EAE0 != 0 {
			hooks.audio(armorDropShoesAudio53EAE0, item, 0, 0)
		} else {
			hooks.audio(armorDropClothAudio53EAE0, item, 0, 0)
		}
	}
}
