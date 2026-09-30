// Public ABI only; the old body is retained for the signature-only step.
#include "../quest_win_screen_450770.h"
#include <limits.h>

typedef int (*quest_win_fn)(const unsigned char*);
_Static_assert(CHAR_BIT == 8, "packet bytes must remain eight bits");
_Static_assert(sizeof(int) == 4, "return values must remain signed dwords");
_Static_assert(_Generic(&nox_xxx_clientQuestWinScreen_450770, quest_win_fn: 1, default: 0),
	"Quest win screen must accept a native packet pointer");

static const unsigned char* expected_packet;
int nox_xxx_clientQuestWinScreen_450770(const unsigned char* packet) {
	return packet == expected_packet && packet[0] == 0xf0 ? INT_MIN + 1 : 0;
}
int main(void) {
	unsigned char packet[90] = {0xf0, 0x0c};
	expected_packet = packet;
	quest_win_fn fn = nox_xxx_clientQuestWinScreen_450770;
	return fn(packet) == INT_MIN + 1 && packet[1] == 0x0c ? 0 : 1;
}
