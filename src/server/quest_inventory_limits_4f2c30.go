package server

var questInventoryTypeNames4F2C30 = [...]string{
	"RedPotion",
	"BluePotion",
	"CurePoisonPotion",
	"HastePotion",
	"InvisibilityPotion",
	"ShieldPotion",
	"VampirismPotion",
	"FireProtectPotion",
	"ShockProtectPotion",
	"PoisonProtectPotion",
	"InvulnerabilityPotion",
	"InfravisionPotion",
	"InfinitePainWand",
}

const (
	questInventoryPotionTypes4F2C30 = len(questInventoryTypeNames4F2C30) - 1
	questInventoryBalanceKey4F2C30  = "ForceOfNatureStaffLimit"
	questInventoryPlayerClass4F2C30 = uint32(4)
)

// questInventoryLimitsCache4F2C30 is the native-width equivalent of the
// thirteen lazy object-type globals at 007533F8..00753428. As in GAME.EXE,
// only RedPotion is the initialization sentinel.
type questInventoryLimitsCache4F2C30 struct {
	typeIDs [len(questInventoryTypeNames4F2C30)]uint32
}

type questInventoryLimitsHooks4F2C30[O any] struct {
	objectTypeID   func(string) uint32
	isNil          func(O) bool
	loadClass      func(O) uint32
	countInventory func(O, int32) int32
	loadBalance    func(string) float32
	floatToInt     func(float32) int32
}

func questInventoryBool4F2C30(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func questInventoryLoadTypeCache4F2C30[O any](
	cache *questInventoryLimitsCache4F2C30,
	hooks questInventoryLimitsHooks4F2C30[O],
) {
	if cache.typeIDs[0] != 0 {
		return
	}
	for index, name := range questInventoryTypeNames4F2C30 {
		cache.typeIDs[index] = hooks.objectTypeID(name)
	}
}

// questInventoryLimits4F2C30 reconstructs GAME.EXE 004F2C30. It preserves
// the original cache-before-owner order, signed int32 count comparisons, and
// short-circuit order of all twelve potion limits before the staff limit.
func questInventoryLimits4F2C30[O any](
	owner O,
	cache *questInventoryLimitsCache4F2C30,
	hooks questInventoryLimitsHooks4F2C30[O],
) int32 {
	questInventoryLoadTypeCache4F2C30(cache, hooks)
	if hooks.isNil(owner) || hooks.loadClass(owner)&questInventoryPlayerClass4F2C30 == 0 {
		return 1
	}
	for index := 0; index < questInventoryPotionTypes4F2C30; index++ {
		if hooks.countInventory(owner, int32(cache.typeIDs[index])) > 9 {
			return 0
		}
	}
	limit := hooks.floatToInt(hooks.loadBalance(questInventoryBalanceKey4F2C30))
	staffCount := hooks.countInventory(owner, int32(cache.typeIDs[questInventoryPotionTypes4F2C30]))
	return questInventoryBool4F2C30(staffCount <= limit)
}
