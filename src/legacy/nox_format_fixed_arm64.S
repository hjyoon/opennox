#ifdef __APPLE__
#define NOX_FIXED _nox_format_fixed_bits
#else
#define NOX_FIXED nox_format_fixed_bits
#endif

.text
.globl NOX_FIXED
.p2align 2
#ifdef __ELF__
.type NOX_FIXED,%function
#endif
NOX_FIXED:
	.cfi_startproc
#ifdef __APPLE__
	// Darwin places the precision and double in eight-byte stack slots.
	sub sp, sp, #32
	.cfi_def_cfa_offset 32
	stp x29, x30, [sp, #16]
	add x29, sp, #16
	.cfi_def_cfa w29, 16
	.cfi_offset w29, -16
	.cfi_offset w30, -8
	stp x3, x0, [sp]
	mov x0, x1
	mov x1, x2
	adrp x2, L_nox_format_fixed@PAGE
	add x2, x2, L_nox_format_fixed@PAGEOFF
	bl _snprintf
	.cfi_def_cfa wsp, 32
	ldp x29, x30, [sp, #16]
	add sp, sp, #32
	.cfi_def_cfa_offset 0
	.cfi_restore w29
	.cfi_restore w30
	ret
#else
#ifdef _WIN32
	// Windows ARM64 passes anonymous double arguments in core registers.
	mov x4, x0
#else
	fmov d0, x0
#endif
	mov x0, x1
	mov x1, x2
	adrp x2, .Lnox_format_fixed
	add x2, x2, :lo12:.Lnox_format_fixed
	b snprintf
#endif
	.cfi_endproc

#ifdef __ELF__
.size NOX_FIXED, .-NOX_FIXED
.section .rodata
#elif defined(__APPLE__)
.section __TEXT,__cstring,cstring_literals
#else
.section .rdata,"dr"
#endif
#ifdef __APPLE__
L_nox_format_fixed:
#else
.Lnox_format_fixed:
#endif
	.asciz "%.*f"

#ifdef __ELF__
.section .note.GNU-stack,"",%progbits
#endif
