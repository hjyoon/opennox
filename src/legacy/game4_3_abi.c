#include "GAME1.h"
#include "GAME4_3.h"
#include "noxstring_format_abi.h"

// The C entry points preserve their pointer-shaped return registers.
//----- (00531B40) --------------------------------------------------------
extern uintptr_t nox_monster_fight_scheduler_native_531B40(nox_object_t*, nox_object_t*, unsigned int);
int* nox_xxx_TODOsomeCallingMeleeAttack_531B40(nox_object_t* unit, nox_object_t* target) {
	return (int*)(uintptr_t)nox_monster_fight_scheduler_native_531B40(unit, target, 0x531B40);
}

//----- (00531C10) --------------------------------------------------------
int* sub_531C10(nox_object_t* unit, nox_object_t* target) {
	return (int*)(uintptr_t)nox_monster_fight_scheduler_native_531B40(unit, target, 0x531C10);
}

//----- (00531C60) --------------------------------------------------------
int* nox_xxx_monsterAction_531C60(nox_object_t* unit, nox_object_t* target) {
	return (int*)(uintptr_t)nox_monster_fight_scheduler_native_531B40(unit, target, 0x531C60);
}

//----- (00531D50) --------------------------------------------------------
int* sub_531D50(nox_object_t* unit, nox_object_t* target) {
	return (int*)(uintptr_t)nox_monster_fight_scheduler_native_531B40(unit, target, 0x531D50);
}

//----- (005341A0) --------------------------------------------------------
void nox_ai_debug_vprintf_5341A0(nox_format_byte* format, nox_format_args* args);
void nox_ai_debug_printf_5341A0(char* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	nox_ai_debug_vprintf_5341A0(format, &args);
	va_end(args.list);
}
