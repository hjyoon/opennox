#ifndef OPENNOX_NATIVE_TEST_STDIO_H
#define OPENNOX_NATIVE_TEST_STDIO_H

/* The native fixtures exchange exact binary inputs and LF receipts with Go.
 * Configure the Windows CRT before any fixture input/output, keeping the
 * production body and the captured oracle bytes unchanged. */
#ifdef _WIN32
#include <fcntl.h>
#include <io.h>
#include <stdio.h>
#include <stdlib.h>

__attribute__((constructor)) static void opennox_native_test_stdio(void) {
	if (_setmode(_fileno(stdin), _O_BINARY) == -1 ||
		_setmode(_fileno(stdout), _O_BINARY) == -1) {
		perror("native fixture binary standard streams");
		exit(125);
	}
}
#endif

#endif
