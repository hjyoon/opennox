.syntax unified
.arm
.text

.macro nox_damage_call entry, count
	.globl \entry
	.p2align 2
	.type \entry,%function
\entry:
	.fnstart
	.save {r4, lr}
	push {r4, lr}
	mov r12, r0
	mov r0, r1
	mov r1, r2
.if \count == 5
	mov r2, r3
	// Bring damage into r3 and pass the type in the outgoing stack slot.
	ldr r3, [sp, #8]
	ldr r4, [sp, #12]
	.pad #8
	sub sp, sp, #8
	str r4, [sp]
.endif
	blx r12
.if \count == 5
	add sp, sp, #8
.endif
	pop {r4, pc}
	.fnend
	.size \entry, .-\entry
.endm

#include "raw_default_damage_calls_4e0b30.h"

.section .note.GNU-stack,"",%progbits
