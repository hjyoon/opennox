.syntax unified
.arm
.text

// AAPCS uses core registers for variadic calls. Spill them immediately below
// the caller's stack arguments, with r0 and r2 at eight-byte boundaries.
.macro nox_variadic_entry entry, handler, named
	.globl \entry
	.p2align 2
	.type \entry,%function
\entry:
	.fnstart
	.pad #16
	push {r0-r3}
	.save {r11, lr}
	push {r11, lr}
	.pad #8
	sub sp, sp, #8
	add r12, sp, #(16 + 4 * \named)
	str r12, [sp]
.if \named == 1
	mov r1, sp
.else
	mov r2, sp
.endif
	bl \handler
	add sp, sp, #8
	pop {r11, lr}
	add sp, sp, #16
	bx lr
	.fnend
	.size \entry, .-\entry
.endm

#include "nox_variadic_entries.h"

.section .note.GNU-stack,"",%progbits
