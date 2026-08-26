package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

type rewardSpellBookNativeDeps4F09F0 struct {
	pickSlots          func(uint32) uint32
	randomInt          func(int32, int32) int32
	checkSpellClass    func(uint32, uint32) int32
	createObjectByType func(string) *Object
}

func rewardSpellBookNative4F09F0(
	marker *Object,
	stage uint32,
	deps rewardSpellBookNativeDeps4F09F0,
) *Object {
	return rewardSpellBook4F09F0(marker, stage, rewardSpellBookHooks4F09F0[
		*Object,
		*RewardMarkerInitData,
		*Object,
	]{
		loadInitData: func(marker *Object) *RewardMarkerInitData {
			return (*RewardMarkerInitData)(marker.InitData)
		},
		loadFlags: func(data *RewardMarkerInitData) uint8 {
			return data.RewardFlags
		},
		loadExplicitSpell: func(data *RewardMarkerInitData, index int) uint8 {
			return data.Spells[index]
		},
		pickSlots:       deps.pickSlots,
		rows:            rewardSpellDefinitions4F09F0[:],
		randomInt:       deps.randomInt,
		checkSpellClass: deps.checkSpellClass,
		createObjectByType: func(typeName string) *Object {
			return deps.createObjectByType(typeName)
		},
		isNilObject: func(object *Object) bool {
			return object == nil
		},
		storeSpell: func(object *Object, spell uint8) {
			object.UseDataSpellReward().Spell = spell
		},
	})
}

func rewardSpellClassCheck4F09F0(flags things.SpellFlags, class uint32) int32 {
	var required things.SpellFlags
	switch class {
	case 1:
		required = things.SpellClassWizard
	case 2:
		required = things.SpellClassConjurer
	default:
		return 9
	}
	if flags.Has(things.SpellClassAny) || flags.Has(required) {
		return 0
	}
	return 9
}

// RewardSpellBook4F09F0 binds GAME.EXE 004F09F0 to native Object pointers,
// fixed-width RewardMarkerInitData, the native spell registry, and native
// SpellRewardUseData. The shared slot helper and both weighted draws consume
// the same server logic RNG. There are deliberately no nil guards.
//
//go:noinline
func (s *Server) RewardSpellBook4F09F0(marker *Object, stage uint32) *Object {
	randomInt := func(minimum, maximum int32) int32 {
		return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
	}
	return rewardSpellBookNative4F09F0(marker, stage, rewardSpellBookNativeDeps4F09F0{
		pickSlots: func(stage uint32) uint32 {
			return rewardRandomSlots4F0B60(stage, randomInt)
		},
		randomInt: randomInt,
		checkSpellClass: func(class, spellID uint32) int32 {
			flags := s.Spells.Flags(spell.ID(spellID))
			return rewardSpellClassCheck4F09F0(flags, class)
		},
		createObjectByType: s.NewObjectByTypeID,
	})
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(RewardMarkerInitData{})]
	_ = [1]struct{}{}[4-unsafe.Offsetof(RewardMarkerInitData{}.RewardFlags)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(RewardMarkerInitData{}.Spells)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(SpellRewardUseData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SpellRewardUseData{}.Spell)]
)
