.text

.macro nox_variadic_entry entry, handler, named
	.globl \entry
	.p2align 4, 0x90
#ifdef __ELF__
	.type \entry,@function
#endif
\entry:
#ifdef _WIN32
	.seh_proc \entry
	// The four caller-allocated home slots precede the stack arguments.
	// Reserve separate shadow space and one slot for the cursor.
	subq $40, %rsp
	.seh_stackalloc 40
	.seh_endprologue
.if \named == 1
	movq %rdx, 56(%rsp)
.endif
	movq %r8, 64(%rsp)
	movq %r9, 72(%rsp)
	leaq (48 + 8 * \named)(%rsp), %rax
	movq %rax, 32(%rsp)
.if \named == 1
	leaq 32(%rsp), %rdx
.else
	leaq 32(%rsp), %r8
.endif
	call \handler
	nop
	addq $40, %rsp
	ret
	.seh_endproc
#else
	.cfi_startproc
	// SysV uses a 24-byte cursor at +8 and register save slots at +32.
	// The original stack arguments start one word above the entry %rsp.
	subq $216, %rsp
	.cfi_def_cfa_offset 224
	movq %rsi, 40(%rsp)
	movq %rdx, 48(%rsp)
	movq %rcx, 56(%rsp)
	movq %r8, 64(%rsp)
	movq %r9, 72(%rsp)
	testb %al, %al
	je 1f
	movaps %xmm0, 80(%rsp)
	movaps %xmm1, 96(%rsp)
	movaps %xmm2, 112(%rsp)
	movaps %xmm3, 128(%rsp)
	movaps %xmm4, 144(%rsp)
	movaps %xmm5, 160(%rsp)
	movaps %xmm6, 176(%rsp)
	movaps %xmm7, 192(%rsp)
1:
	movl $(8 * \named), 8(%rsp)
	movl $48, 12(%rsp)
	leaq 224(%rsp), %rax
	movq %rax, 16(%rsp)
	leaq 32(%rsp), %rax
	movq %rax, 24(%rsp)
.if \named == 1
	leaq 8(%rsp), %rsi
.else
	leaq 8(%rsp), %rdx
.endif
	call \handler
	addq $216, %rsp
	.cfi_def_cfa_offset 8
	ret
	.cfi_endproc
#endif
#ifdef __ELF__
	.size \entry, .-\entry
#endif
.endm

#include "nox_variadic_entries.h"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
