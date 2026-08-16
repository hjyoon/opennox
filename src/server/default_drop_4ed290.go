package server

const (
	defaultDropMonsterClass4ED290 = uint32(0x00000002)
	defaultDropPlayerClass4ED290  = uint32(0x00000004)
	defaultDropAudioClass4ED290   = uint32(0x00000040)
	defaultDropWeaponClass4ED290  = uint32(0x01000000)
	defaultDropFlagClass4ED290    = uint32(0x10000000)

	defaultDropRejectOwnerFlags4ED290 = uint32(0x00008020)
	defaultDropCandidateFlag4ED290    = uint32(0x00000100)
	defaultDropNoDecayFlag4ED290      = uint32(0x00080000)
	defaultDropMonsterStatus4ED290    = uint32(0x00000100)

	defaultDropCoopFlag4ED290  = uint32(0x00000800)
	defaultDropQuestFlag4ED290 = uint32(0x00001000)
)

// defaultDropHooks4ED290 keeps the original object, point, and update-data
// pointer domains distinct from GAME.EXE 004ED290's exact access order.
// Repeated loads are intentionally separate: legacy callbacks can mutate the
// owner, item, inventory, caches, or frame between any two of them.
type defaultDropHooks4ED290[O, P, U comparable] struct {
	loadOwnerArg       func() O
	loadItemArg        func() O
	loadInventoryOwner func(O) O

	loadObjectClassByte func(O) uint8
	loadObjectClass     func(O) uint32
	loadObjectFlags     func(O) uint32
	loadObjectNetCode   func(O) uint32
	loadObjectTeamID    func(O) uint8
	loadObjectType      func(O) uint16
	loadObjectUpdate    func(O) U

	itemIsDroppable func(O) int32
	itemDropMask    func(O, uint32) int32
	primaryMessage  func(O, string, uint8)
	audio           func(uint32, O, int32, uint32)

	detachInventory func(O, O)
	loadPointArg    func() P
	loadPointY      func(P) float32
	loadPointX      func(P) float32
	createAt        func(O, O, float32, float32, uint32)

	weaponEquipFlags  func(O) uint32
	loadInventoryHead func(O) O
	loadInventoryNext func(O) O
	loadUpdateByte2   func(U) uint8
	delayedDelete     func(O)

	materialIndex      func(O) uint32
	netInformFlagDrop  func(uint8, uint32, uint32)
	markMinimapForAll  func(O, uint32)
	loadFrame          func() uint32
	storeUpdateFrame   func(U, uint32)
	setTeamFlagStatus  func(uint8, uint8, uint8, uint16)
	loadMonsterStatus  func(U) uint32
	storeMonsterAction func(U, uint32)
	storeMonsterStatus func(U, uint32)

	loadGlyphCache    func() uint32
	storeGlyphCache   func(uint32)
	loadTorchCache    func() uint32
	storeTorchCache   func(uint32)
	loadLanternCache  func() uint32
	storeLanternCache func(uint32)
	lookupType        func(string) uint32
	gameFlag          func(uint32) uint32
	loadGameFPS       func() uint32
	setDecayTime      func(O, uint32)
	raise             func(O, float32)
	buffOff           func(O, uint32)
}

// defaultDrop4ED290 preserves GAME.EXE 004ED290.
//
// The inventory-owner comparison is the only entry gate. In particular, the
// original does not nil-check either object or the point. Its success path
// reloads live object state after callbacks, except for the Flag update-data,
// team ID, and material values explicitly cached by the x86 registers/stack.
func defaultDrop4ED290[O, P, U comparable](hooks defaultDropHooks4ED290[O, P, U]) int32 {
	var zeroObject O

	owner := hooks.loadOwnerArg()
	item := hooks.loadItemArg()
	if hooks.loadInventoryOwner(item) != owner {
		return 0
	}

	if hooks.loadObjectClassByte(owner)&uint8(defaultDropPlayerClass4ED290) != 0 &&
		hooks.itemIsDroppable(item) != 0 &&
		hooks.itemDropMask(item, 1) != 0 {
		if hooks.loadObjectFlags(owner)&defaultDropRejectOwnerFlags4ED290 == 0 {
			hooks.primaryMessage(owner, "drop.c:CantDropThat", 0)
			netCode := hooks.loadObjectNetCode(owner)
			hooks.audio(925, owner, 2, netCode)
		}
		return 0
	}

	hooks.detachInventory(owner, item)
	point := hooks.loadPointArg()
	y := hooks.loadPointY(point)
	x := hooks.loadPointX(point)
	// The fifth zero is the unused stack slot present at the original call site.
	hooks.createAt(item, zeroObject, x, y, 0)

	if hooks.loadObjectClass(item)&defaultDropWeaponClass4ED290 != 0 &&
		hooks.weaponEquipFlags(item) == 4 {
		candidate := hooks.loadInventoryHead(owner)
		for candidate != zeroObject {
			if hooks.loadObjectClass(candidate)&defaultDropWeaponClass4ED290 != 0 &&
				hooks.weaponEquipFlags(candidate) == 2 &&
				hooks.loadObjectFlags(candidate)&defaultDropCandidateFlag4ED290 == 0 {
				update := hooks.loadObjectUpdate(candidate)
				if hooks.loadUpdateByte2(update) != 0 {
					hooks.detachInventory(owner, candidate)
					hooks.delayedDelete(candidate)
					break
				}
			}
			candidate = hooks.loadInventoryNext(candidate)
		}
	}

	if hooks.loadObjectClass(item)&defaultDropFlagClass4ED290 != 0 {
		update := hooks.loadObjectUpdate(item)
		material := hooks.materialIndex(item)
		teamID := hooks.loadObjectTeamID(item)
		netCode := hooks.loadObjectNetCode(owner)
		hooks.netInformFlagDrop(7, netCode, material)
		hooks.markMinimapForAll(item, 1)
		frame := hooks.loadFrame()
		hooks.storeUpdateFrame(update, frame)
		hooks.setTeamFlagStatus(teamID, 2, uint8(material), 0)
	}

	if hooks.loadGlyphCache() == 0 {
		hooks.storeGlyphCache(hooks.lookupType("Glyph"))
	}
	if hooks.gameFlag(defaultDropCoopFlag4ED290) == 0 &&
		hooks.gameFlag(defaultDropQuestFlag4ED290) == 0 &&
		hooks.loadObjectFlags(item)&defaultDropNoDecayFlag4ED290 == 0 &&
		hooks.loadObjectClass(item)&defaultDropFlagClass4ED290 == 0 {
		glyph := hooks.loadGlyphCache()
		itemType := uint32(hooks.loadObjectType(item))
		if itemType != glyph {
			// x86 LEA+SHL performs the multiplication modulo 2^32.
			hooks.setDecayTime(item, hooks.loadGameFPS()*10)
		}
	}

	hooks.raise(item, 0)
	if hooks.loadObjectClassByte(item)&uint8(defaultDropAudioClass4ED290) != 0 {
		hooks.audio(821, item, 0, 0)
	}
	if hooks.loadObjectClassByte(item)&uint8(defaultDropMonsterClass4ED290) != 0 {
		update := hooks.loadObjectUpdate(item)
		status := hooks.loadMonsterStatus(update)
		hooks.storeMonsterAction(update, 15)
		hooks.storeMonsterStatus(update, status|defaultDropMonsterStatus4ED290)
	}

	if hooks.loadTorchCache() == 0 {
		hooks.storeTorchCache(hooks.lookupType("Torch"))
		hooks.storeLanternCache(hooks.lookupType("Lantern"))
	}
	torch := hooks.loadTorchCache()
	itemType := uint32(hooks.loadObjectType(item))
	if itemType == torch || itemType == hooks.loadLanternCache() {
		hooks.buffOff(owner, 15)
	}
	return 1
}
