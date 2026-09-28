package server

import (
	"math"

	"github.com/opennox/libs/object"
)

const (
	questShopAnkhType50E970   = "AnkhTradable"
	questShopMarkerType50E970 = "RewardMarker"
)

// questShopBonusItemTypes50E970 is the exact 46-entry PE32 pointer table at
// 005C0540 (00587000:234816). Keeping its semantic contents in native Go
// avoids walking packed four-byte pointers as host-width pointers on LP64.
var questShopBonusItemTypes50E970 = [...]string{
	"Diamond", "Diamond", "Diamond", "Diamond", "Diamond",
	"Emerald", "Emerald", "Emerald", "Emerald", "Emerald",
	"Ruby", "Ruby", "Ruby", "Ruby", "Ruby",
	"RedPotion", "RedPotion", "RedPotion",
	"BluePotion", "BluePotion", "BluePotion",
	"CurePoisonPotion", "CurePoisonPotion", "CurePoisonPotion",
	"HastePotion", "InvisibilityPotion", "ShieldPotion", "VampirismPotion",
	"FireProtectPotion", "ShockProtectPotion", "PoisonProtectPotion",
	"InvulnerabilityPotion", "InfravisionPotion",
	"Sword", "StaffWooden", "LesserFireballWand",
	"Quiver", "Quiver", "Quiver",
	"FanChakram", "FanChakram",
	"LeatherBoots", "LeatherArmor", "LeatherHelm", "WizardRobe", "WizardHelm",
}

// questShopRewardCategories50E970 preserves the exact 17 activation calls in
// GAME.EXE's Quest branch: four weapons, four armor pieces, six spell books,
// and three field guides. A single ability-book activation may follow after
// the original inclusive 0..100 roll.
var questShopRewardCategories50E970 = [...]uint32{
	8, 8, 8, 8,
	16, 16, 16, 16,
	1, 1, 1, 1, 1, 1,
	4, 4, 4,
}

// QuestShopItemLoadRuntime50E970 supplies the state that lives in the outer
// game server. Generated item and marker pointers stay native-width.
type QuestShopItemLoadRuntime50E970 struct {
	QuestStage      int32
	AnkhCutoffStage float32
	DelayedDelete   func(*Object)
}

type questShopItemLoadHooks50E970 struct {
	newObject      func(string) *Object
	addItem        func(*Object) bool
	markerData     func(*Object) *RewardMarkerInitData
	activateReward func(*Object, uint32) *Object
	randomInt      func(int32, int32) int32
	delayedDelete  func(*Object)
}

// questShopRoundFloat32ToInt32_50E970 models nox_float2int at 00419A70:
// x87 round-to-nearest-even and INT32_MIN for invalid or out-of-range values.
func questShopRoundFloat32ToInt32_50E970(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

// loadQuestShopItems50E970 is the pointer-width-independent control flow of
// the Quest half of GAME.EXE 0050E970. A nil reward is a valid generator
// result and is skipped exactly like the original; failures to create a fixed
// item, own a created item, or construct the temporary marker make complete
// false while allowing the remaining original sequence to run.
func loadQuestShopItems50E970(
	stage, ankhCutoff int32,
	hooks questShopItemLoadHooks50E970,
) (loaded int, complete bool) {
	if hooks.newObject == nil || hooks.addItem == nil || hooks.markerData == nil ||
		hooks.activateReward == nil || hooks.randomInt == nil || hooks.delayedDelete == nil {
		return 0, false
	}
	complete = true
	addType := func(typeID string) {
		item := hooks.newObject(typeID)
		if item == nil || !hooks.addItem(item) {
			complete = false
			return
		}
		loaded++
	}
	if uint32(stage) < uint32(ankhCutoff) {
		addType(questShopAnkhType50E970)
	}
	for _, typeID := range questShopBonusItemTypes50E970 {
		addType(typeID)
	}

	marker := hooks.newObject(questShopMarkerType50E970)
	if marker == nil {
		return loaded, false
	}
	defer hooks.delayedDelete(marker)
	data := hooks.markerData(marker)
	if data == nil {
		return loaded, false
	}
	rewardStage := uint32(stage) + 2
	addReward := func(category uint32) {
		data.CategoryMask = category
		item := hooks.activateReward(marker, rewardStage)
		if item == nil {
			return
		}
		if !hooks.addItem(item) {
			complete = false
			return
		}
		loaded++
	}
	for _, category := range questShopRewardCategories50E970 {
		addReward(category)
	}
	if hooks.randomInt(0, 100) > 90 {
		addReward(2)
	}
	return loaded, complete
}

// LoadQuestShopItemsNative50E970 restores the generated Quest inventory half
// of 0050E970. Every generated Object, RewardMarker, and TradeItem link stays
// native-width; only serialized type indices and prices remain fixed-width.
func (s *Server) LoadQuestShopItemsNative50E970(
	session *TradeSession,
	runtime QuestShopItemLoadRuntime50E970,
) (loaded int, complete bool) {
	if session == nil || !s.IsTradeSessionNative(session) || runtime.DelayedDelete == nil {
		return 0, false
	}
	merchant := session.Field8
	if merchant != nil && merchant.Class().Has(object.ClassPlayer) {
		merchant = session.Field12
	}
	if merchant == nil || merchant.InitData == nil {
		return 0, false
	}
	addItem := func(item *Object) bool {
		if item == nil {
			return false
		}
		cost, ok := s.shopItemBuyCost50E3D0(session, item)
		if !ok || s.addShopItemNative50EE00(session, item, cost) == nil {
			s.Objs.FreeObject(item)
			return false
		}
		return true
	}
	return loadQuestShopItems50E970(
		runtime.QuestStage,
		questShopRoundFloat32ToInt32_50E970(runtime.AnkhCutoffStage),
		questShopItemLoadHooks50E970{
			newObject: s.NewObjectByTypeID,
			addItem:   addItem,
			markerData: func(marker *Object) *RewardMarkerInitData {
				return (*RewardMarkerInitData)(marker.InitData)
			},
			activateReward: s.RewardMarkerActivateDefault4F0720,
			randomInt: func(minimum, maximum int32) int32 {
				return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
			},
			delayedDelete: runtime.DelayedDelete,
		},
	)
}
