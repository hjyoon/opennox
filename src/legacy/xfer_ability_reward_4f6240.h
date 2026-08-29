#ifndef NOX_XFER_ABILITY_REWARD_4F6240_H
#define NOX_XFER_ABILITY_REWARD_4F6240_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"AbilityRewardXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerAbilityReward_4F6240(nox_object_t* object);

#endif // NOX_XFER_ABILITY_REWARD_4F6240_H
