.text

// Cdecl places the callback first, followed by the callback's own arguments.
.macro nox_damage_call entry, count
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
	movl 12(%ebp), %eax
	movl %eax, (%esp)
	movl 16(%ebp), %eax
	movl %eax, 4(%esp)
.if \count == 5
	movl 20(%ebp), %eax
	movl %eax, 8(%esp)
	movl 24(%ebp), %eax
	movl %eax, 12(%esp)
	movl 28(%ebp), %eax
	movl %eax, 16(%esp)
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

#include "raw_default_damage_calls_4e0b30.h"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
