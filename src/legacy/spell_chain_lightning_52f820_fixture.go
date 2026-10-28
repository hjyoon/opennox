package legacy

/*
#include "spell_chain_lightning_52f820.h"

typedef int (*nox_chain_lifecycle_int_fn)(void*);
typedef char (*nox_chain_lifecycle_char_fn)(void*);
_Static_assert(_Generic(&nox_xxx_onStartLightning_52F820, nox_chain_lifecycle_int_fn: 1, default: 0), "native create ABI");
_Static_assert(_Generic(&nox_xxx_onFrameLightning_52F8A0, nox_chain_lifecycle_int_fn: 1, default: 0), "native update ABI");
_Static_assert(_Generic(&sub_530100, nox_chain_lifecycle_char_fn: 1, default: 0), "native destroy ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func chainLightningLifecycleCreateC52F820(record *server.DurSpell) uint32 {
	return uint32(C.nox_xxx_onStartLightning_52F820(record.C()))
}

func chainLightningLifecycleDestroyC530100(record *server.DurSpell) uint32 {
	return uint32(uint8(C.sub_530100(record.C())))
}
