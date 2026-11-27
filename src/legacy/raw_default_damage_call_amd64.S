.text

.macro nox_damage_call entry, count
	.globl \entry
	.p2align 4, 0x90
#ifdef __ELF__
	.type \entry,@function
#endif
\entry:
#ifdef _WIN32
	.seh_proc \entry
	subq $40, %rsp
	.seh_stackalloc 40
	.seh_endprologue
	movq %rcx, %rax
	movq %rdx, %rcx
	movq %r8, %rdx
.if \count == 5
	// The original damage and type follow the incoming home slots.
	movq %r9, %r8
	movl 80(%rsp), %r9d
	movl 88(%rsp), %r10d
	movl %r10d, 32(%rsp)
.endif
	call *%rax
	nop
	addq $40, %rsp
	ret
	.seh_endproc
#else
	.cfi_startproc
	subq $8, %rsp
	.cfi_def_cfa_offset 16
	movq %rdi, %rax
	movq %rsi, %rdi
	movq %rdx, %rsi
.if \count == 5
	movq %rcx, %rdx
	movl %r8d, %ecx
	movl %r9d, %r8d
.endif
	call *%rax
	addq $8, %rsp
	.cfi_def_cfa_offset 8
	ret
	.cfi_endproc
#endif
#ifdef __ELF__
	.size \entry, .-\entry
#endif
.endm

#include "raw_default_damage_calls_4e0b30.h"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
