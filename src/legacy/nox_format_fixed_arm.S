.syntax unified
.arm
.text
.globl nox_format_fixed_bits
.p2align 2
.type nox_format_fixed_bits,%function
nox_format_fixed_bits:
	.fnstart
	.save {r4, lr}
	push {r4, lr}
	.pad #8
	sub sp, sp, #8
	// The named arguments and precision fill the core registers.
	strd r0, r1, [sp]
	mov r0, r2
	mov r1, r3
	adr r2, 1f
	ldr r3, [sp, #16]
	bl snprintf
	add sp, sp, #8
	pop {r4, pc}
	.fnend
.size nox_format_fixed_bits, .-nox_format_fixed_bits
1:
	.asciz "%.*f"

.section .note.GNU-stack,"",%progbits
