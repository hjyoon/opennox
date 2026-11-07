package legacy

/*
#include "GAME1.h"
#include <stddef.h>
#include <stdint.h>
extern uint32_t dword_5d4594_526276;
typedef int (*setting_eligibility_fixture_fn_40aa70)(nox_playerInfo*);
_Static_assert(_Generic(&sub_40AA70, setting_eligibility_fixture_fn_40aa70:1, default:0), "40AA70 int(player*) ABI");
_Static_assert(sizeof(int)==4, "40AA70 int32 return");
_Static_assert(sizeof(((nox_playerInfo*)0)->field_3680)==4 && sizeof(((nox_playerInfo*)0)->field_2068)==4 && sizeof(((nox_playerInfo*)0)->field_2140)==4, "40AA70 three player DWORDs");
static int setting_eligibility_fixture_call_40aa70(void* p) {
    setting_eligibility_fixture_fn_40aa70 fn=&sub_40AA70;
    return fn((nox_playerInfo*)p);
}
static size_t setting_eligibility_fixture_offset_40aa70(int field) {
    switch(field) {
    case 0: return offsetof(nox_playerInfo, field_3680);
    case 1: return offsetof(nox_playerInfo, field_2068);
    case 2: return offsetof(nox_playerInfo, field_2140);
    default: return (size_t)-1;
    }
}
static size_t setting_eligibility_fixture_pointer_width_40aa70(void) { return sizeof(void*); }
static void* setting_eligibility_fixture_limit_cell_40aa70(void) { return &dword_5d4594_526276; }
*/
import "C"
import "unsafe"

type settingEligibilityFixtureField40AA70 struct{ Offset, Width uintptr }

func settingEligibilityFixtureCall40AA70(player unsafe.Pointer) int32 {
	return int32(C.setting_eligibility_fixture_call_40aa70(player))
}
func settingEligibilityFixtureCLayout40AA70() [3]settingEligibilityFixtureField40AA70 {
	var out [3]settingEligibilityFixtureField40AA70
	for i := range out {
		out[i] = settingEligibilityFixtureField40AA70{uintptr(C.setting_eligibility_fixture_offset_40aa70(C.int(i))), 4}
	}
	return out
}
func settingEligibilityFixtureCPointerWidth40AA70() uintptr {
	return uintptr(C.setting_eligibility_fixture_pointer_width_40aa70())
}

// Existing C-owned limit is exposed only so controls can save and restore it.
func settingEligibilityFixtureLimitCell40AA70() unsafe.Pointer {
	return C.setting_eligibility_fixture_limit_cell_40aa70()
}
