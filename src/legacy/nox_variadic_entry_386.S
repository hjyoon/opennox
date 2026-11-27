.text

// i386 cdecl stores all arguments in four-byte stack slots. Keep the list
// pointer in the last slot of a 16-byte aligned outgoing frame.
.macro nox_variadic_entry entry, handler, named
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
	subl $16, %esp

	leal (8 + 4 * \named)(%ebp), %eax
	movl %eax, 12(%esp)
	leal 12(%esp), %eax
	movl %eax, (4 * \named)(%esp)
	movl 8(%ebp), %eax
	movl %eax, (%esp)
.if \named == 2
	movl 12(%ebp), %eax
	movl %eax, 4(%esp)
.endif
	call \handler

	leave
	.cfi_def_cfa %esp, 4
	.cfi_restore %ebp
	ret
	.cfi_endproc
#ifdef __ELF__
	.size \entry, .-\entry
#endif
.endm

#include "nox_variadic_entries.h"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
