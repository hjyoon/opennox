#include "GAME1.h"
#include "GAME4_3.h"
#include "noxstring.h"

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

// The variadic entry formats the legacy log before the Go logger.
//----- (005341A0) --------------------------------------------------------
int nox_ai_debug_enabled(void);
void nox_ai_debug_print(char* str);
void nox_ai_debug_printf_5341A0(char* a1, ...) {
	va_list va; // [esp+8h] [ebp+8h]

	va_start(va, a1);
	if (nox_ai_debug_enabled()) {
		nox_vsprintf((char*)getMemAt(0x5D4594, 2487996), a1, va);
		nox_ai_debug_print((char*)getMemAt(0x5D4594, 2487996));
	}
	va_end(va);
}
