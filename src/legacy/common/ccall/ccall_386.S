.text

// The cdecl callback gets its own aligned argument area.
.macro nox_ccall_entry entry, count
	.globl \entry
	.p2align 4, 0x90
#ifdef __ELF__
	.type \entry,@function
#endif
\entry:
	.cfi_startproc
	pushl %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset %ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register %ebp
	andl $-16, %esp
	subl $32, %esp
.if \count >= 1
	movl 12(%ebp), %eax
	movl %eax, (%esp)
.endif
.if \count >= 2
	movl 16(%ebp), %eax
	movl %eax, 4(%esp)
.endif
.if \count >= 3
	movl 20(%ebp), %eax
	movl %eax, 8(%esp)
.endif
.if \count >= 4
	movl 24(%ebp), %eax
	movl %eax, 12(%esp)
.endif
.if \count >= 5
	movl 28(%ebp), %eax
	movl %eax, 16(%esp)
.endif
.if \count >= 6
	movl 32(%ebp), %eax
	movl %eax, 20(%esp)
.endif
	call *8(%ebp)
	leave
	.cfi_def_cfa %esp, 4
	.cfi_restore %ebp
	ret
	.cfi_endproc
#ifdef __ELF__
	.size \entry, .-\entry
#endif
.endm

#include "ccall_abi.h"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
