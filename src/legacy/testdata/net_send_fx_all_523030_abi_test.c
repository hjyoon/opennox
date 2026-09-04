#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../net_send_fx_all_523030.h"

struct float2 {
	float field_0;
	float field_4;
};

typedef int (*net_send_fx_all_fn)(float2*, const void*, int);
typedef int (*net_send_fx_all_native_fn)(float2*, void*, int);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int) == 4, "packet size and result must remain 32-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(float2) == 8, "float2 must remain two binary32 values");
_Static_assert(
	_Generic(&nox_xxx_netSendFxAllCli_523030,
		net_send_fx_all_fn: 1, default: 0),
	"00523030 must preserve native position and const packet pointers");
_Static_assert(
	_Generic(&nox_xxx_netSendFxAllCliNative_523030,
		net_send_fx_all_native_fn: 1, default: 0),
	"00523030 native bridge must preserve both pointer-width arguments");

static float2* observed_pos;
static void* observed_data;
static int observed_size;

int nox_xxx_netSendFxAllCliNative_523030(float2* pos, void* data, int size) {
	observed_pos = pos;
	observed_data = data;
	observed_size = size;
	return INT_MIN;
}

int nox_xxx_netSendFxAllCli_523030(float2* pos, const void* data, int size) {
	return nox_xxx_netSendFxAllCliNative_523030(pos, (void*)data, size);
}

int main(void) {
	float2 pos = {.field_0 = 123.5F, .field_4 = -45.25F};
	uint8_t packet[] = {UINT8_C(0x99), UINT8_C(0x30), UINT8_C(0x23)};
	net_send_fx_all_fn const send_fx = nox_xxx_netSendFxAllCli_523030;

	assert(send_fx(&pos, packet, INT_MIN) == INT_MIN);
	assert(observed_pos == &pos);
	assert(observed_data == packet);
	assert(observed_size == INT_MIN);
	return 0;
}
