package legacy

/*
#include "GAME3_2.h"

typedef intptr_t (*vital_report_entry)(int, nox_object_t*);
typedef int (*enchantment_report_entry)(int, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_netReportStamina_4D8800, vital_report_entry: 1, default: 0), "stamina ABI");
_Static_assert(_Generic(&sub_4D8840, enchantment_report_entry: 1, default: 0), "enchantment ABI");
_Static_assert(_Generic(&sub_4D8870, vital_report_entry: 1, default: 0), "gold ABI");
_Static_assert(_Generic(&nox_xxx_netReportTotalMana_4D88C0, vital_report_entry: 1, default: 0), "total mana ABI");
_Static_assert(_Generic(&nox_xxx_netReportMana_4D8930, vital_report_entry: 1, default: 0), "mana ABI");

static intptr_t call_vital_report_entry(int entry, int recipient, nox_object_t* unit) {
	if (entry == 1) {
		enchantment_report_entry volatile send = sub_4D8840;
		return send(recipient, unit);
	}
	vital_report_entry volatile entries[] = {
		nox_xxx_netReportStamina_4D8800,
		NULL,
		sub_4D8870,
		nox_xxx_netReportTotalMana_4D88C0,
		nox_xxx_netReportMana_4D8930,
	};
	return entries[entry](recipient, unit);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func callVitalReportEntry(entry int, recipient int32, unit *server.Object) uintptr {
	return uintptr(C.call_vital_report_entry(C.int(entry), C.int(recipient), asObjectC(unit)))
}
