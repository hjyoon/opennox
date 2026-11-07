package legacy

/*
#include "GAME1.h"
#include <stddef.h>
#include <stdint.h>
typedef void (*sudden_death_fixture_fn_40a970)(void);
_Static_assert(_Generic(&sub_40A970, sudden_death_fixture_fn_40a970: 1, default: 0), "40A970 void(void) ABI");
_Static_assert(sizeof(((nox_playerInfo*)0)->field_3680) == 4, "40A970 player status DWORD");
static void sudden_death_fixture_call_40a970(void) {
    sudden_death_fixture_fn_40a970 fn = &sub_40A970;
    fn();
}
static size_t sudden_death_fixture_player_offset_40a970(void) {
    return offsetof(nox_playerInfo, field_3680);
}
static size_t sudden_death_fixture_player_width_40a970(void) {
    return sizeof(((nox_playerInfo*)0)->field_3680);
}
*/
import "C"

// The public C entry has no parameters and no integer return transport.
func suddenDeathFixtureCall40A970() {
	C.sudden_death_fixture_call_40a970()
}

// Both expressions concern the compiled declaration only. No player instance
// or complete struct allocation is required by this helper.
func suddenDeathFixtureCPlayerLayout40A970() (offset, width uintptr) {
	return uintptr(C.sudden_death_fixture_player_offset_40a970()), uintptr(C.sudden_death_fixture_player_width_40a970())
}
