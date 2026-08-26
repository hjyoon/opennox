package legacy

/*
#include <stdint.h>

#include "reward_marker_activate_4f0720.h"

uint32_t* nox_xxx_rewardMakeArmor_4F0E80(int32_t marker, uint32_t stage);
int32_t nox_xxx_rewardMakeWeapon_4F14E0(int32_t marker, uint32_t stage);
uint32_t* nox_xxx_rewardMakePotion_4F1C40(int32_t marker, uint32_t stage);
uint32_t* nox_xxx_createGem_4F1D30(int32_t marker, uint32_t stage);
uint32_t* nox_xxx_createGem2_4F1F00(int32_t marker, uint32_t stage);

// The five creator bodies after the restored 004F09F0, 004F0C70, and 004F0D20
// paths remain separate ABI32 restoration units. Keep their sole pointer
// narrowing visible here while the public marker and restored creators use
// native pointers.
static inline int32_t nox_rewardMarkerCreatorArgABI32_4F0720(
		nox_object_t* marker) {
	return (int32_t)(uintptr_t)marker;
}

static inline nox_object_t* nox_rewardMarkerWeapon_4F0720(
		nox_object_t* marker, uint32_t stage) {
	return (nox_object_t*)(uintptr_t)(uint32_t)nox_xxx_rewardMakeWeapon_4F14E0(
		nox_rewardMarkerCreatorArgABI32_4F0720(marker), stage);
}

static inline nox_object_t* nox_rewardMarkerArmor_4F0720(
		nox_object_t* marker, uint32_t stage) {
	return (nox_object_t*)nox_xxx_rewardMakeArmor_4F0E80(
		nox_rewardMarkerCreatorArgABI32_4F0720(marker), stage);
}

static inline nox_object_t* nox_rewardMarkerGem_4F0720(
		nox_object_t* marker, uint32_t stage) {
	return (nox_object_t*)nox_xxx_createGem_4F1D30(
		nox_rewardMarkerCreatorArgABI32_4F0720(marker), stage);
}

static inline nox_object_t* nox_rewardMarkerPotion_4F0720(
		nox_object_t* marker, uint32_t stage) {
	return (nox_object_t*)nox_xxx_rewardMakePotion_4F1C40(
		nox_rewardMarkerCreatorArgABI32_4F0720(marker), stage);
}

static inline nox_object_t* nox_rewardMarkerGem2_4F0720(
		nox_object_t* marker, uint32_t stage) {
	return (nox_object_t*)nox_xxx_createGem2_4F1F00(
		nox_rewardMarkerCreatorArgABI32_4F0720(marker), stage);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func rewardMarkerObjectFromC4F0720(object *C.nox_object_t) *server.Object {
	return asObjectS((*nox_object_t)(object))
}

func rewardMarkerActivateRuntime4F0720(s *server.Server) server.RewardMarkerActivateRuntime4F0720 {
	return server.RewardMarkerActivateRuntime4F0720{
		SpellBook: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardSpellBook4F09F0(marker, stage)
		},
		AbilityBook: func(marker *server.Object, _ uint32) *server.Object {
			return s.RewardAbilityBook4F0C70(marker)
		},
		FieldGuide: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardFieldGuide4F0D20(marker, stage)
		},
		Weapon: func(marker *server.Object, stage uint32) *server.Object {
			return rewardMarkerObjectFromC4F0720(C.nox_rewardMarkerWeapon_4F0720(
				(*C.nox_object_t)(asObjectC(marker)), C.uint32_t(stage),
			))
		},
		Armor: func(marker *server.Object, stage uint32) *server.Object {
			return rewardMarkerObjectFromC4F0720(C.nox_rewardMarkerArmor_4F0720(
				(*C.nox_object_t)(asObjectC(marker)), C.uint32_t(stage),
			))
		},
		Gem: func(marker *server.Object, stage uint32) *server.Object {
			return rewardMarkerObjectFromC4F0720(C.nox_rewardMarkerGem_4F0720(
				(*C.nox_object_t)(asObjectC(marker)), C.uint32_t(stage),
			))
		},
		Potion: func(marker *server.Object, stage uint32) *server.Object {
			return rewardMarkerObjectFromC4F0720(C.nox_rewardMarkerPotion_4F0720(
				(*C.nox_object_t)(asObjectC(marker)), C.uint32_t(stage),
			))
		},
		Gem2: func(marker *server.Object, stage uint32) *server.Object {
			return rewardMarkerObjectFromC4F0720(C.nox_rewardMarkerGem2_4F0720(
				(*C.nox_object_t)(asObjectC(marker)), C.uint32_t(stage),
			))
		},
	}
}

//export nox_server_rewardgen_activateMarker_4F0720
func nox_server_rewardgen_activateMarker_4F0720(
	marker *C.nox_object_t,
	stage C.uint32_t,
) *C.nox_object_t {
	s := GetServer().S()
	result := rewardMarkerActivateCall4F0720(
		s,
		asObjectS((*nox_object_t)(marker)),
		uint32(stage),
		rewardMarkerActivateRuntime4F0720(s),
	)
	return (*C.nox_object_t)(asObjectC(result))
}
