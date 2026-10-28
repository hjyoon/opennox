#ifndef NOX_PORT_SPELL_CHAIN_LIGHTNING_52F820_H
#define NOX_PORT_SPELL_CHAIN_LIGHTNING_52F820_H

typedef struct nox_object_t nox_object_t;

// Duration records are opaque native pointers, not PE32 words or float bits.
int nox_xxx_onStartLightning_52F820(void* record);
int nox_xxx_onFrameLightning_52F8A0(void* record);
char sub_530100(void* record);
void nox_xxx_lightningSpellDuration_52FFD0(void* record, nox_object_t* from, nox_object_t* to);

#endif
