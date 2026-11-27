.syntax unified
.arm
.text

.macro nox_ccall_entry entry, count
	.globl \entry
	.p2align 2
	.type \entry,%function
\entry:
	.fnstart
	.save {r4, lr}
	push {r4, lr}
	.pad #8
	sub sp, sp, #8
	mov r12, r0
.if \count >= 1
	mov r0, r1
.endif
.if \count >= 2
	mov r1, r2
.endif
.if \count >= 3
	mov r2, r3
.endif
.if \count >= 4
	// The first three arguments were in registers after the callback pointer.
	ldr r3, [sp, #16]
.endif
.if \count >= 5
	ldr r4, [sp, #20]
	str r4, [sp]
.endif
.if \count >= 6
	ldr r4, [sp, #24]
	str r4, [sp, #4]
.endif
	blx r12
	add sp, sp, #8
	pop {r4, pc}
	.fnend
	.size \entry, .-\entry
.endm

#include "ccall_abi.h"

.section .note.GNU-stack,"",%progbits
