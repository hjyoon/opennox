#ifndef NOX_UNIT_ADJUST_HP_4EE460_H
#define NOX_UNIT_ADJUST_HP_4EE460_H

typedef struct nox_object_t nox_object_t;

void nox_xxx_unitAdjustHP_4EE460(nox_object_t* unit, int delta);
void nox_xxx_mobInformOwnerHP_4EE4C0(nox_object_t* object);
int nox_xxx_netReportUnitCurrentHP_4D8620(int recipient, nox_object_t* object);

#endif // NOX_UNIT_ADJUST_HP_4EE460_H
