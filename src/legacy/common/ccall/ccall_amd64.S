.text

.macro nox_ccall_entry entry, count
	.globl \entry
	.p2align 4, 0x90
#ifdef __ELF__
	.type \entry,@function
#endif
\entry:
#ifdef _WIN32
	.seh_proc \entry
	// Reserve home slots, two outgoing stack arguments, and alignment.
	subq $56, %rsp
	.seh_stackalloc 56
	.seh_endprologue
	movq %rcx, %r11
.if \count >= 1
	movq %rdx, %rcx
.endif
.if \count >= 2
	movq %r8, %rdx
.endif
.if \count >= 3
	movq %r9, %r8
.endif
.if \count >= 4
	movq 96(%rsp), %r9
.endif
.if \count >= 5
	movq 104(%rsp), %rax
	movq %rax, 32(%rsp)
.endif
.if \count >= 6
	movq 112(%rsp), %rax
	movq %rax, 40(%rsp)
.endif
	call *%r11
	nop
	addq $56, %rsp
	ret
	.seh_endproc
#else
	.cfi_startproc
	subq $8, %rsp
	.cfi_def_cfa_offset 16
	movq %rdi, %r11
.if \count >= 1
	movq %rsi, %rdi
.endif
.if \count >= 2
	movq %rdx, %rsi
.endif
.if \count >= 3
	movq %rcx, %rdx
.endif
.if \count >= 4
	movq %r8, %rcx
.endif
.if \count >= 5
	movq %r9, %r8
.endif
.if \count >= 6
	// The sixth callback argument follows the incoming return address.
	movq 16(%rsp), %r9
.endif
	call *%r11
	addq $8, %rsp
	.cfi_def_cfa_offset 8
	ret
	.cfi_endproc
#endif
#ifdef __ELF__
	.size \entry, .-\entry
#endif
.endm

#include "ccall_abi.h"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
