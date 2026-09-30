// Public ABI fixture; this stub is not the runtime briefing body.
#include "../quest_briefing_450980.h"

#include <limits.h>
#include <stdint.h>

typedef int (*quest_briefing_fn)(const unsigned char*, int);

_Static_assert(CHAR_BIT == 8, "packet bytes must remain eight bits");
_Static_assert(sizeof(int) == 4, "show and return values must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_client_showQuestBriefing2_450980, quest_briefing_fn: 1, default: 0),
	"Quest stage briefing must retain the native packet pointer");

static const unsigned char* expected_packet;

int nox_client_showQuestBriefing2_450980(const unsigned char* packet, int a2) {
	if (packet != expected_packet || a2 != INT_MIN || packet[0] != UINT8_C(0xF0))
		return -1;
	return INT_MIN + 1;
}

int main(void) {
	unsigned char packet[69] = {UINT8_C(0xF0), UINT8_C(0x0D)};
	quest_briefing_fn const briefing = nox_client_showQuestBriefing2_450980;
	expected_packet = packet;
	if (briefing(packet, INT_MIN) != INT_MIN + 1)
		return __LINE__;
	if (packet[0] != UINT8_C(0xF0) || packet[1] != UINT8_C(0x0D))
		return __LINE__;
	return 0;
}
