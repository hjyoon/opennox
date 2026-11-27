.text

// All arguments fit in the general-purpose registers on AArch64.
.macro nox_ccall_entry entry, count
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
.if \count >= 1
	mov x0, x1
.endif
.if \count >= 2
	mov x1, x2
.endif
.if \count >= 3
	mov x2, x3
.endif
.if \count >= 4
	mov x3, x4
.endif
.if \count >= 5
	mov x4, x5
.endif
.if \count >= 6
	mov x5, x6
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

#include "ccall_abi.h"

#ifdef __ELF__
.section .note.GNU-stack,"",%progbits
#endif
