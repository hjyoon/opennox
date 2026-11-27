#ifndef OPENNOX_CCALL_ABI_H
#define OPENNOX_CCALL_ABI_H

// Each entry gives the cgo declaration and the number of callback arguments.
// The native entries forward register and stack slots without changing C types.
#ifdef __ASSEMBLER__
#define NOX_CCALL_SYMBOL_(prefix, name) prefix##name
#define NOX_CCALL_SYMBOL(prefix, name) NOX_CCALL_SYMBOL_(prefix, name)
#define NOX_CCALL_ENTRY(name, count, result, params, body) \
    nox_ccall_entry NOX_CCALL_SYMBOL(__USER_LABEL_PREFIX__, name), count
#else
#include <stdint.h>

typedef struct nox_window nox_window;
typedef struct nox_window_data nox_window_data;

#ifdef NOX_CCALL_FALLBACK
#define NOX_CCALL_ENTRY(name, count, result, params, body) result name params { body }
#else
#define NOX_CCALL_ENTRY(name, count, result, params, body) result name params;
#endif
#endif

#include "ccall_entries.h"

NOX_CCALL_ENTRY(nox_ccall_window_proc, 4, int,
    (int (*fnc)(nox_window*, int, uintptr_t, uintptr_t),
     nox_window* win, int event, uintptr_t a, uintptr_t b),
    return fnc(win, event, a, b);)
NOX_CCALL_ENTRY(nox_ccall_uptr_ptr_int_uptr2, 4, uintptr_t,
    (uintptr_t (*fnc)(nox_window*, int, uintptr_t, uintptr_t),
     nox_window* win, int event, uintptr_t a, uintptr_t b),
    return fnc(win, event, a, b);)
NOX_CCALL_ENTRY(nox_ccall_window_draw, 2, int,
    (int (*fnc)(nox_window*, nox_window_data*), nox_window* win, nox_window_data* draw),
    return fnc(win, draw);)
NOX_CCALL_ENTRY(nox_ccall_window_tooltip, 3, void,
    (int (*fnc)(nox_window*, nox_window_data*, int), nox_window* win, nox_window_data* draw, int pos),
    fnc(win, draw, pos);)

#undef NOX_CCALL_ENTRY
#ifdef __ASSEMBLER__
#undef NOX_CCALL_SYMBOL
#undef NOX_CCALL_SYMBOL_
#endif

#endif
