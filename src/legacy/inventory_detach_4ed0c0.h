#ifndef NOX_INVENTORY_DETACH_4ED0C0_H
#define NOX_INVENTORY_DETACH_4ED0C0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

void sub_4ED0C0(nox_object_t* owner, nox_object_t* item);

// Narrow declarations for the legacy services used by the Go-owned body.
// The uint32_t* owner parameters still describe ABI32 service internals; the
// public sub_4ED0C0 object pointers above remain native-width.
int nox_xxx_netReportDequip_4D84C0(int index, const nox_object_t* item);
int sub_53E430(uint32_t* owner, nox_object_t* item, int mode, int report);
int nox_xxx_playerDequipWeapon_53A140(uint32_t* owner, nox_object_t* item, int mode, int report);
int nox_xxx_netReportDrop_4D8B50(int index, const nox_object_t* item);
int nox_xxx_protect_56FC50(int value, const nox_object_t* item);

#endif // NOX_INVENTORY_DETACH_4ED0C0_H
