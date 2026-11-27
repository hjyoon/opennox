.text

// Preserve a named, non-tail indirect call frame for a native target fault.
.macro nox_damage_call entry, count
	.globl \entry
	.p2align 2
#ifdef __ELF__
	.type \entry,%function
#endif
\entry:
#ifdef _WIN32
	.seh_proc \entry
	str x30, [sp, #-16]!
	.seh_save_reg_x x30, 16
	.seh_endprologue
#else
	.cfi_startproc
	stp x29, x30, [sp, #-16]!
	.cfi_def_cfa_offset 16
	.cfi_offset w29, -16
	.cfi_offset w30, -8
	mov x29, sp
#endif
	mov x16, x0
	mov x0, x1
	mov x1, x2
.if \count == 5
	mov x2, x3
	mov w3, w4
	mov w4, w5
.endif
	blr x16
#ifdef _WIN32
	.seh_startepilogue
	ldr x30, [sp], #16
	.seh_save_reg_x x30, 16
	.seh_endepilogue
	ret
	.seh_endfunclet
	.seh_endproc
#else
	ldp x29, x30, [sp], #16
	.cfi_def_cfa_offset 0
	.cfi_restore w29
	.cfi_restore w30
	ret
	.cfi_endproc
#endif
#ifdef __ELF__
	.size \entry, .-\entry
#endif
.endm

#include "raw_default_damage_calls_4e0b30.h"

#ifdef __ELF__
.section .note.GNU-stack,"",%progbits
#endif
