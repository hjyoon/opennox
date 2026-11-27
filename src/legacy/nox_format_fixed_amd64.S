#ifdef __APPLE__
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
#ifdef _WIN32
	.seh_proc NOX_FIXED
	// The double is the fifth printf argument, above the shadow space.
	subq $40, %rsp
	.seh_stackalloc 40
	.seh_endprologue
	movq %rcx, 32(%rsp)
	movq %rdx, %rcx
	movq %r8, %rdx
	leaq 1f(%rip), %r8
	call snprintf
	nop
	addq $40, %rsp
	ret
	.seh_endproc
#else
	.cfi_startproc
	movq %rdi, %xmm0
	movq %rsi, %rdi
	movq %rdx, %rsi
	leaq 1f(%rip), %rdx
	// The SysV variadic entry receives the vector register count in %al.
	movl $1, %eax
#ifdef __APPLE__
	jmp _snprintf
#else
	jmp snprintf@PLT
#endif
	.cfi_endproc
#endif

#ifdef __ELF__
.size NOX_FIXED, .-NOX_FIXED
.section .rodata
#elif defined(__APPLE__)
.section __TEXT,__cstring,cstring_literals
#else
.section .rdata,"dr"
#endif
1:
	.asciz "%.*f"

#ifdef __ELF__
.section .note.GNU-stack,"",@progbits
#endif
