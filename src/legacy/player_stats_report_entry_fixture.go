package legacy

/*
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME5_2.h"

typedef short (*health_delta_entry)(int, short, short);
typedef intptr_t (*item_health_entry)(int, nox_object_t*);
typedef char (*player_stats_entry)(int, nox_object_t*, char);
typedef int (*armor_value_entry)(int, int);
typedef intptr_t (*durability_report_entry)(int, nox_object_t*, unsigned short, unsigned short);
typedef int (*health_bucket_entry)(unsigned short, unsigned short);
typedef int (*bulk_stats_flush_entry)(void);
_Static_assert(_Generic(&nox_xxx_netReportHealthDelta_4D8760, health_delta_entry: 1, default: 0), "health delta ABI");
_Static_assert(_Generic(&nox_xxx_itemReportHealth_4D87A0, item_health_entry: 1, default: 0), "item health ABI");
_Static_assert(_Generic(&nox_xxx_netReportPickup_4D8A60, item_health_entry: 1, default: 0), "pickup result ABI");
_Static_assert(_Generic(&nox_xxx_netReportModifiablePickup_4D8AD0, item_health_entry: 1, default: 0), "modified pickup result ABI");
_Static_assert(_Generic(&nox_xxx_servSendStats_4D8990, player_stats_entry: 1, default: 0), "player stats ABI");
_Static_assert(_Generic(&nox_xxx_netReportArmorVal_4D8A30, armor_value_entry: 1, default: 0), "armor value ABI");
_Static_assert(_Generic(&nox_xxx_itemDestroyed_4E1650, durability_report_entry: 1, default: 0), "durability report ABI");
_Static_assert(_Generic(&sub_57B190, health_bucket_entry: 1, default: 0), "health bucket ABI");
_Static_assert(_Generic(&nox_xxx_net_4263C0, bulk_stats_flush_entry: 1, default: 0), "bulk stats public ABI");

static int call_bulk_stats_flush(void) {
	bulk_stats_flush_entry volatile flush = nox_xxx_net_4263C0;
	return flush();
}

static uintptr_t call_stats_report_entry(int entry, int recipient, nox_object_t* obj, uint32_t a, uint32_t b) {
	switch (entry) {
	case 0: {
		health_delta_entry volatile send = nox_xxx_netReportHealthDelta_4D8760;
		return (uint16_t)send(recipient, (short)a, (short)b);
	}
	case 1: {
		item_health_entry volatile send = nox_xxx_itemReportHealth_4D87A0;
		return (uintptr_t)send(recipient, obj);
	}
	case 2: {
		player_stats_entry volatile send = nox_xxx_servSendStats_4D8990;
		return (uint8_t)send(recipient, obj, (char)a);
	}
	case 3: {
		armor_value_entry volatile send = nox_xxx_netReportArmorVal_4D8A30;
		return (uint32_t)send(recipient, (int)a);
	}
	case 4: {
		durability_report_entry volatile send = nox_xxx_itemDestroyed_4E1650;
		return (uintptr_t)send(recipient, obj, (unsigned short)a, (unsigned short)b);
	}
	case 5: {
		health_bucket_entry volatile bucket = sub_57B190;
		return (uintptr_t)bucket((unsigned short)a, (unsigned short)b);
	}
	default: return 0;
	}
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func callStatsReportEntry(entry int, recipient int32, unit *server.Object, a, b uint32) uintptr {
	return uintptr(C.call_stats_report_entry(C.int(entry), C.int(recipient), asObjectC(unit), C.uint32_t(a), C.uint32_t(b)))
}

func callStatsBulkFlushEntry() int32 {
	return int32(C.call_bulk_stats_flush())
}
