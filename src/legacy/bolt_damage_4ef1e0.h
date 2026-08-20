#ifndef NOX_PORT_BOLT_DAMAGE_4EF1E0
#define NOX_PORT_BOLT_DAMAGE_4EF1E0

#include <stdint.h>

typedef struct nox_modifier_t nox_modifier_t;

double nox_xxx_calcBoltDamage_4EF1E0(int32_t strength, nox_modifier_t* modifier);
double nox_xxx_calcBoltDamageValues_4EF1E0(
	int32_t strength,
	uint32_t type_index,
	uint16_t required_strength,
	float coefficient,
	uint16_t minimum);

uint32_t nox_xxx_boltDamageModifierType_4EF1E0(nox_modifier_t* modifier);
uint16_t nox_xxx_boltDamageModifierMinimum_4EF1E0(nox_modifier_t* modifier);
float nox_xxx_boltDamageModifierRange_4EF1E0(nox_modifier_t* modifier);

#endif // NOX_PORT_BOLT_DAMAGE_4EF1E0
