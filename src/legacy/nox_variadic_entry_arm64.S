.text

.macro nox_variadic_entry entry, handler, named
	.globl \entry
	.p2align 2
#ifdef __ELF__
	.type \entry,%function
#endif
\entry:
#ifdef _WIN32
	.seh_proc \entry
	// Windows passes anonymous floats in core register slots. Save x1-x7
	// directly below the incoming stack, with a separate cursor at +8.
	str x30, [sp, #-80]!
	.seh_save_reg_x x30, 80
	.seh_endprologue
	stp x1, x2, [sp, #24]
	stp x3, x4, [sp, #40]
	stp x5, x6, [sp, #56]
	str x7, [sp, #72]
	add x8, sp, #(16 + 8 * \named)
	str x8, [sp, #8]
.if \named == 1
	add x1, sp, #8
.else
	add x2, sp, #8
.endif
	bl \handler
	.seh_startepilogue
	ldr x30, [sp], #80
	.seh_save_reg_x x30, 80
	.seh_endepilogue
	ret
	.seh_endfunclet
	.seh_endproc
#elif defined(__APPLE__)
	.cfi_startproc
	// Darwin places anonymous arguments on the caller's stack.
	sub sp, sp, #32
	.cfi_def_cfa_offset 32
	stp x29, x30, [sp, #16]
	add x29, sp, #16
	.cfi_def_cfa w29, 16
	.cfi_offset w29, -16
	.cfi_offset w30, -8
	add x8, sp, #32
	str x8, [sp, #8]
.if \named == 1
	add x1, sp, #8
.else
	add x2, sp, #8
.endif
	bl \handler
	.cfi_def_cfa wsp, 32
	ldp x29, x30, [sp, #16]
	add sp, sp, #32
	.cfi_def_cfa_offset 0
	.cfi_restore w29
	.cfi_restore w30
	ret
	.cfi_endproc
#else
	.cfi_startproc
	// AAPCS64 saves eight vector slots and eight core slots below the
	// 32-byte cursor. Its two offsets are measured from each area's end.
	sub sp, sp, #240
	.cfi_def_cfa_offset 240
	stp x29, x30, [sp, #224]
	add x29, sp, #224
	.cfi_def_cfa w29, 16
	.cfi_offset w29, -16
	.cfi_offset w30, -8
	stp q0, q1, [sp]
	stp q2, q3, [sp, #32]
	stp q4, q5, [sp, #64]
	stp q6, q7, [sp, #96]
	stp x1, x2, [sp, #136]
	stp x3, x4, [sp, #152]
	stp x5, x6, [sp, #168]
	str x7, [sp, #184]
	add x8, sp, #240
	add x9, sp, #192
	stp x8, x9, [sp, #192]
	add x8, sp, #128
	str x8, [sp, #208]
	mov w8, #(-64 + 8 * \named)
	mov w9, #-128
	stp w8, w9, [sp, #216]
.if \named == 1
	add x1, sp, #192
.else
	add x2, sp, #192
.endif
	bl \handler
	.cfi_def_cfa wsp, 240
	ldp x29, x30, [sp, #224]
	add sp, sp, #240
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

#include "nox_variadic_entries.h"

#ifdef __ELF__
.section .note.GNU-stack,"",%progbits
#endif
