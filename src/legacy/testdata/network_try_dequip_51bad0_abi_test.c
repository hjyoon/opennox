// Keep the CGo bridge ABI independent from Win32-only aggregate headers.
#include "../network_try_dequip_51bad0.h"

#include <limits.h>
#include <stdint.h>

struct nox_object_t {
	uintptr_t marker;
};

typedef int (*network_try_dequip_fn)(unsigned char*, nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_server_netTryDequip_51BAD0,
		network_try_dequip_fn: 1, default: 0),
	"MSG_TRY_DEQUIP must preserve all pointer-width arguments");

static unsigned char* expected_data;
static nox_object_t* expected_unit;
static void* expected_update;

int nox_server_netTryDequip_51BAD0(
		unsigned char* data,
		nox_object_t* unit,
		void* update) {
	if (data != expected_data || unit != expected_unit ||
		update != expected_update || data[0] != UINT8_C(0x76))
		return -1;
	return 3;
}

int main(void) {
	unsigned char packet[3] = {UINT8_C(0x76), UINT8_C(0x34), UINT8_C(0x12)};
	nox_object_t unit = {.marker = UINTPTR_MAX};
	uintptr_t update = (uintptr_t)0x1234U;
	network_try_dequip_fn const try_dequip = nox_server_netTryDequip_51BAD0;
	expected_data = packet;
	expected_unit = &unit;
	expected_update = &update;
	if (try_dequip(packet, &unit, &update) != 3)
		return __LINE__;
	if (unit.marker != UINTPTR_MAX || update != (uintptr_t)0x1234U)
		return __LINE__;
	return 0;
}
