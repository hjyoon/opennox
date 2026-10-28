#ifndef NOX_PORT_SPELL_BLINK_NATIVE_530310_H
#define NOX_PORT_SPELL_BLINK_NATIVE_530310_H

int nox_xxx_spellBlink2_530310(void* record);
int nox_xxx_spellBlink1_530380(void* record);
// Callers observe only the created wake. The old incidental pointer/integer
// return register is not a result contract.
void nox_xxx_spellTeleportCreateWake_530560(void* target, void* origin, void* destination);

#endif
