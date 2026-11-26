package cfixture

/*
#include "world_native_storage.h"
#define dword_5d4594_526276 (nox_world_word_storage()->team_entries_limit)
#include "GAME1.h"
#include <stddef.h>
#include <stdint.h>
typedef int (*setting_eligibility_fixture_fn_40aa70)(nox_playerInfo*);
typedef void* (*setting_eligibility_info_fn_40aa70)(void);
typedef char (*setting_eligibility_char_fn_40aa70)(void);
typedef int (*setting_eligibility_limit_fn_40aa70)(void);
_Static_assert(_Generic(&sub_40AA70, setting_eligibility_fixture_fn_40aa70:1, default:0), "40AA70 int(player*) ABI");
_Static_assert(sizeof(int)==4, "40AA70 int32 return");
_Static_assert(_Generic(&sub_416640,setting_eligibility_info_fn_40aa70:1,default:0), "Info416640 native void pointer ABI");
_Static_assert(_Generic(&sub_417DE0,setting_eligibility_char_fn_40aa70:1,default:0), "count417DE0 char ABI");
_Static_assert(_Generic(&sub_417DC0,setting_eligibility_limit_fn_40aa70:1,default:0), "limit417DC0 int ABI");
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

type SettingEligibilityFixtureField40AA70 struct{ Offset, Width uintptr }

func SettingEligibilityFixtureCall40AA70(player unsafe.Pointer) int32 {
	return int32(C.setting_eligibility_fixture_call_40aa70(player))
}
func SettingEligibilityFixtureCLayout40AA70() [3]SettingEligibilityFixtureField40AA70 {
	var out [3]SettingEligibilityFixtureField40AA70
	for i := range out {
		out[i] = SettingEligibilityFixtureField40AA70{uintptr(C.setting_eligibility_fixture_offset_40aa70(C.int(i))), 4}
	}
	return out
}
func SettingEligibilityFixtureCPointerWidth40AA70() uintptr {
	return uintptr(C.setting_eligibility_fixture_pointer_width_40aa70())
}

// The live team-entry limit is exposed so controls can save and restore it.
func SettingEligibilityFixtureLimitCell40AA70() unsafe.Pointer {
	return C.setting_eligibility_fixture_limit_cell_40aa70()
}
