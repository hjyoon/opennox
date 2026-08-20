#ifndef NOX_PORT_POISON_STATE_4EE7E0
#define NOX_PORT_POISON_STATE_4EE7E0

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

int32_t nox_xxx_activatePoison_4EE7E0(nox_object_t* unit, int32_t increment, int32_t maximum);
void nox_xxx_updatePoison_4EE8F0(nox_object_t* unit, int32_t amount);
void nox_xxx_removePoison_4EE9D0(nox_object_t* unit);
void nox_xxx_setSomePoisonData_4EEA90(nox_object_t* unit, int32_t value);

#endif // NOX_PORT_POISON_STATE_4EE7E0
