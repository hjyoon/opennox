#ifndef NOX_PORT_SPELL_TAG_NATIVE_530160_H
#define NOX_PORT_SPELL_TAG_NATIVE_530160_H

int nox_xxx_spellTagCreature_530160(void* record);
unsigned int sub_530250(void* record);
// Duration destruction has no return-value contract; no native pointer is
// exposed through the original incidental PE32 return register.
void sub_530270(void* record);

#endif
