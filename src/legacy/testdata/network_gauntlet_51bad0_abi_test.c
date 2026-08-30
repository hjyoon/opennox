// Keep the CGo bridge ABI independent from Win32-only aggregate headers.
#include "../network_gauntlet_51bad0.h"

#include <limits.h>
#include <stdint.h>

struct nox_object_t {
	uintptr_t marker;
};

typedef int (*network_gauntlet_fn)(unsigned char*, nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_server_netGauntlet_51BAD0,
		network_gauntlet_fn: 1, default: 0),
	"MSG_GAUNTLET must preserve all pointer-width arguments");

static unsigned char* expected_data;
static nox_object_t* expected_unit;
static void* expected_update;

int nox_server_netGauntlet_51BAD0(
		unsigned char* data,
		nox_object_t* unit,
		void* update) {
	if (data != expected_data || unit != expected_unit ||
		update != expected_update || data[0] != UINT8_C(0xF0))
		return -1;
	return 2;
}

int main(void) {
	unsigned char packet[2] = {UINT8_C(0xF0), UINT8_C(0x03)};
	nox_object_t unit = {.marker = UINTPTR_MAX};
	uintptr_t update = (uintptr_t)0x1234U;
	network_gauntlet_fn const gauntlet = nox_server_netGauntlet_51BAD0;
	expected_data = packet;
	expected_unit = &unit;
	expected_update = &update;
	if (gauntlet(packet, &unit, &update) != 2)
		return __LINE__;
	if (unit.marker != UINTPTR_MAX || update != (uintptr_t)0x1234U)
		return __LINE__;
	return 0;
}
