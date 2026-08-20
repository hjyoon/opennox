#include "../unit_adjust_hp_4ee460.h"

typedef void (*unit_adjust_hp_callback_t)(nox_object_t*, int);
typedef void (*mob_inform_owner_hp_callback_t)(nox_object_t*);
typedef int (*current_hp_report_callback_t)(int, nox_object_t*);

_Static_assert(sizeof(int) == 4, "HP adjustment integers must remain 32-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitAdjustHP_4EE460, unit_adjust_hp_callback_t: 1, default: 0),
	"UnitAdjustHP must use one native object pointer and one 32-bit delta");
_Static_assert(
	_Generic(&nox_xxx_mobInformOwnerHP_4EE4C0, mob_inform_owner_hp_callback_t: 1, default: 0),
	"MobInformOwnerHP must use one native object pointer");
_Static_assert(
	_Generic(&nox_xxx_netReportUnitCurrentHP_4D8620, current_hp_report_callback_t: 1, default: 0),
	"NetReportUnitCurrentHP must use a 32-bit recipient and one native object pointer");

int main(void) {
	return 0;
}
