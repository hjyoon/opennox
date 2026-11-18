package legacy

/*
#include <stdlib.h>
#include "GAME3_2.h"

typedef void (*self_report_fn)(nox_object_t*);
typedef int (*key_report_fn)(int, nox_object_t*, char);
typedef int (*creature_report_fn)(int, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_playerReportAnything_4D9900, self_report_fn:1, default:0), "self report ABI");
_Static_assert(_Generic(&sub_4D9DF0, key_report_fn:1, default:0), "silver key report ABI");
_Static_assert(_Generic(&sub_4D9E30, key_report_fn:1, default:0), "gold key report ABI");
_Static_assert(_Generic(&nox_xxx_netFxShield_0_4D9200, creature_report_fn:1, default:0), "creature report ABI");

static self_report_fn volatile self_report_indirect = nox_xxx_playerReportAnything_4D9900;
static key_report_fn volatile silver_report_indirect = sub_4D9DF0;
static key_report_fn volatile gold_report_indirect = sub_4D9E30;
static creature_report_fn volatile creature_report_indirect = nox_xxx_netFxShield_0_4D9200;

static void call_self_report_native(int indirect, nox_object_t* unit) {
	if (indirect) {
		self_report_indirect(unit);
	} else {
		nox_xxx_playerReportAnything_4D9900(unit);
	}
}

static int call_report_packet_native(int indirect, int entry, int recipient, nox_object_t* unit, char value) {
	switch (entry) {
	case 0:
		return indirect ? silver_report_indirect(recipient, unit, value) : sub_4D9DF0(recipient, unit, value);
	case 1:
		return indirect ? gold_report_indirect(recipient, unit, value) : sub_4D9E30(recipient, unit, value);
	case 2:
		return indirect ? creature_report_indirect(recipient, unit) : nox_xxx_netFxShield_0_4D9200(recipient, unit);
	default:
		abort();
	}
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func callPlayerSelfReport4D9900(route int, unit *server.Object) {
	C.call_self_report_native(C.int(route), asObjectC(unit))
}

func callPlayerReportPacketNative(route, entry int, recipient uint32, unit *server.Object, value uint8) uint32 {
	return uint32(C.call_report_packet_native(C.int(route), C.int(entry), C.int(int32(recipient)), asObjectC(unit), C.char(value)))
}
