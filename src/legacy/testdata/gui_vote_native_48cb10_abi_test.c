#include <stddef.h>
#include <stdint.h>

#include "../GAME2_3.h"
#include "../client__gui__guivote.h"

typedef uintptr_t (*vote_dialog_fn)(int);
typedef int (*vote_window_proc_fn)(nox_window*, int, nox_window*, uintptr_t);
typedef int (*vote_player_message_fn)(wchar2_t*);
typedef uintptr_t (*vote_topic_fn)(void);

_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"vote GUI pointers must remain native-width");
_Static_assert(_Generic(&sub_48CB10, vote_dialog_fn: 1, default: 0),
	"0048CB10 must preserve native-width dialog results");
_Static_assert(_Generic(&nox_xxx_guiKick_48D0A0, vote_window_proc_fn: 1, default: 0),
	"0048D0A0 must preserve native-width GUI event pointers");
_Static_assert(_Generic(&nox_xxx_voteSend_48D260, vote_player_message_fn: 1, default: 0),
	"0048D260 must accept player names without pointer narrowing");
_Static_assert(_Generic(&nox_xxx_netSendRenameMb_48D2D0, vote_player_message_fn: 1, default: 0),
	"0048D2D0 must accept player names without pointer narrowing");
_Static_assert(_Generic(&sub_48D410, vote_topic_fn: 1, default: 0),
	"0048D410 must preserve native-width listbox responses");
_Static_assert(offsetof(nox_playerInfo, name_final) == (sizeof(void*) == 4 ? 4704 : 6008),
	"vote player names must use the native player layout");
_Static_assert(offsetof(nox_playerInfo, field_4792) == (sizeof(void*) == 4 ? 4792 : 6096),
	"vote eligibility must use the native player layout");

int main(void) {
	return 0;
}
