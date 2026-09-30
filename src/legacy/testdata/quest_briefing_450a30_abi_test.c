// Public ABI fixture; the real body is exercised by the Go/CGo tests.
#include "../quest_briefing_450a30.h"

#include <limits.h>
#include <stdint.h>

typedef int (*quest_briefing_fn)(const unsigned char*, int);

_Static_assert(CHAR_BIT == 8, "packet bytes must remain eight bits");
_Static_assert(sizeof(int) == 4, "show and return values must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_client_showQuestBriefing_450A30, quest_briefing_fn: 1, default: 0),
	"Quest start briefing must retain the native packet pointer");

static const unsigned char* expected_packet;
static int expected_show;

int nox_client_showQuestBriefing_450A30(const unsigned char* packet, int a2) {
	if (packet != expected_packet || a2 != expected_show || packet[0] != UINT8_C(0xF0))
		return -1;
	return INT_MIN + 1;
}

int main(void) {
	unsigned char packet[69] = {UINT8_C(0xF0), UINT8_C(0x0E)};
	quest_briefing_fn const briefing = nox_client_showQuestBriefing_450A30;
	expected_packet = packet;
	expected_show = INT_MIN;
	if (briefing(packet, INT_MIN) != INT_MIN + 1)
		return __LINE__;
	if (packet[0] != UINT8_C(0xF0) || packet[1] != UINT8_C(0x0E))
		return __LINE__;
	return 0;
}
