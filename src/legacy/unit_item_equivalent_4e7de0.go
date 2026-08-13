package legacy

const (
	itemEquivalentInfoBookClass4E7DE0      = uint32(0x00000100)
	itemEquivalentModifierClass4E7DE0      = uint32(0x13001000)
	itemEquivalentSpellBookSubclass4E7DE0  = uint32(0x00000001)
	itemEquivalentFieldGuideSubclass4E7DE0 = uint32(0x00000002)
)

type itemEquivalentHooks4E7DE0[O, A, U, M comparable] struct {
	loadType     func(O) uint16
	loadClass    func(O) uint32
	loadInitData func(O) A
	loadModifier func(A, int) M
	loadSubclass func(O) uint32
	loadUseData  func(O) U
	loadUseByte  func(U, int) byte
}

// itemEquivalent4E7DE0 preserves GAME.EXE 004E7DE0. The candidate alone
// selects the modifier and InfoBook comparison modes. Modifier-bearing items
// compare four native pointer identities. SpellBooks compare one use-data
// byte, FieldGuides compare a NUL-terminated byte string, and other InfoBooks
// compare one byte. Type, field, and byte loads retain the original
// short-circuit order.
func itemEquivalent4E7DE0[O, A, U, M comparable](
	candidate O,
	item O,
	hooks itemEquivalentHooks4E7DE0[O, A, U, M],
) bool {
	var zero O
	if candidate == zero || item == zero {
		return false
	}
	if hooks.loadType(candidate) != hooks.loadType(item) {
		return false
	}

	objectClass := hooks.loadClass(candidate)
	if objectClass&itemEquivalentModifierClass4E7DE0 != 0 {
		candidateAttrs := hooks.loadInitData(candidate)
		itemAttrs := hooks.loadInitData(item)
		for i := range 4 {
			candidateModifier := hooks.loadModifier(candidateAttrs, i)
			itemModifier := hooks.loadModifier(itemAttrs, i)
			if candidateModifier != itemModifier {
				return false
			}
		}
	}

	if objectClass&itemEquivalentInfoBookClass4E7DE0 == 0 {
		return true
	}

	objectSubclass := hooks.loadSubclass(candidate)
	if objectSubclass&itemEquivalentSpellBookSubclass4E7DE0 != 0 {
		candidateUse := hooks.loadUseData(candidate)
		itemUse := hooks.loadUseData(item)
		candidateByte := hooks.loadUseByte(candidateUse, 0)
		itemByte := hooks.loadUseByte(itemUse, 0)
		return candidateByte == itemByte
	}
	if objectSubclass&itemEquivalentFieldGuideSubclass4E7DE0 != 0 {
		itemUse := hooks.loadUseData(item)
		candidateUse := hooks.loadUseData(candidate)
		for i := 0; ; i++ {
			candidateByte := hooks.loadUseByte(candidateUse, i)
			itemByte := hooks.loadUseByte(itemUse, i)
			if candidateByte != itemByte {
				return false
			}
			if candidateByte == 0 {
				return true
			}
		}
	}

	candidateUse := hooks.loadUseData(candidate)
	itemUse := hooks.loadUseData(item)
	candidateByte := hooks.loadUseByte(candidateUse, 0)
	itemByte := hooks.loadUseByte(itemUse, 0)
	return candidateByte == itemByte
}
