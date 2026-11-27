#define NOX_DAMAGE_SYMBOL_(prefix, name) prefix##name
#define NOX_DAMAGE_SYMBOL(prefix, name) NOX_DAMAGE_SYMBOL_(prefix, name)
#define NOX_DAMAGE_CALL(entry, count) \
	nox_damage_call NOX_DAMAGE_SYMBOL(__USER_LABEL_PREFIX__, entry), count

NOX_DAMAGE_CALL(rdd_native_damage_callback_4e0b30, 5)
NOX_DAMAGE_CALL(rdd_native_sound_callback_4e0b30, 2)

#undef NOX_DAMAGE_CALL
#undef NOX_DAMAGE_SYMBOL
#undef NOX_DAMAGE_SYMBOL_
