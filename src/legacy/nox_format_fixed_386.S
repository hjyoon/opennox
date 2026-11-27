#ifdef _WIN32
#define NOX_FIXED _nox_format_fixed_bits
#else
#define NOX_FIXED nox_format_fixed_bits
#endif

.text
.globl NOX_FIXED
.p2align 4, 0x90
#ifdef __ELF__
.type NOX_FIXED,@function
#endif
NOX_FIXED:
	.cfi_startproc
	pushl %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset %ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register %ebp
#ifdef __ELF__
	pushl %ebx
	.cfi_offset %ebx, -12
#endif
	andl $-16, %esp
	subl $32, %esp
	movl 16(%ebp), %eax
	movl %eax, (%esp)
	movl 20(%ebp), %eax
	movl %eax, 4(%esp)
#ifdef __ELF__
	call 1f
1:
	popl %ebx
	addl $_GLOBAL_OFFSET_TABLE_+(.-1b), %ebx
	leal .Lnox_format_fixed@GOTOFF(%ebx), %eax
#else
	leal .Lnox_format_fixed, %eax
#endif
	movl %eax, 8(%esp)
	movl 24(%ebp), %eax
	movl %eax, 12(%esp)
	// Copy the double to the variadic argument area as two DWORDs.
	movl 8(%ebp), %eax
	movl %eax, 16(%esp)
	movl 12(%ebp), %eax
	movl %eax, 20(%esp)
#ifdef __ELF__
	call snprintf@PLT
	movl -4(%ebp), %ebx
	.cfi_restore %ebx
#else
	call _snprintf
#endif
	leave
	.cfi_def_cfa %esp, 4
	.cfi_restore %ebp
	ret
	.cfi_endproc

#ifdef __ELF__
.size NOX_FIXED, .-NOX_FIXED
.section .rodata
#else
.section .rdata,"dr"
#endif
.Lnox_format_fixed:
	.asciz "%.*f"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
