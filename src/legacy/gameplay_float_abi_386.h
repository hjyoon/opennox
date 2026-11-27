#ifndef NOX_GAMEPLAY_FLOAT_ABI_386_H
#define NOX_GAMEPLAY_FLOAT_ABI_386_H

#if defined(__i386__)
_Static_assert(sizeof(float) == 4 && sizeof(int) == 4 && sizeof(void*) == 4,
               "i386 gameplay arguments occupy DWORD stack slots");

// Both Go exports are defined in the same generated cgo translation unit.
// Alias the public cdecl entries to their word-based exports so the stack
// slots reach Go unchanged, including signaling NaNs.
#define NOX_FLOAT_STRINGIFY_386_(x) #x
#define NOX_FLOAT_STRINGIFY_386(x) NOX_FLOAT_STRINGIFY_386_(x)
#define NOX_FLOAT_SYMBOL_386(name) NOX_FLOAT_STRINGIFY_386(__USER_LABEL_PREFIX__) #name
#define NOX_FLOAT_WORD_ALIAS_386(public_name, word_name) \
	__asm__(".globl " NOX_FLOAT_SYMBOL_386(public_name) "\n" \
		".set " NOX_FLOAT_SYMBOL_386(public_name) ", " NOX_FLOAT_SYMBOL_386(word_name) "\n")

NOX_FLOAT_WORD_ALIAS_386(nox_xxx_mapDamageUnitsAround_4E25B0, nox_xxx_mapDamageUnitsAroundBits_4E25B0);
NOX_FLOAT_WORD_ALIAS_386(nox_xxx_updateObjectsVelocity_5118A0, nox_update_objects_velocity_bits_5118a0);

#undef NOX_FLOAT_WORD_ALIAS_386
#undef NOX_FLOAT_SYMBOL_386
#undef NOX_FLOAT_STRINGIFY_386
#undef NOX_FLOAT_STRINGIFY_386_
#endif

#endif
