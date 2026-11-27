//go:build safe

package cfixture

/*
#include <stddef.h>

void* nox_memset(void*, int, size_t);
void* nox_memcpy(void*, const void*, size_t);
char* nox_strcpy(char*, const char*);
char* nox_strcat(char*, const char*);

static int safe_memory_results(char* out) {
	if (nox_memset(out + 1, 0x142, 14) != out + 1) return 1;
	if (out[1] != 'B' || out[14] != 'B') return 2;
	if (nox_memcpy(out + 2, "xy", 3) != out + 2) return 3;
	if (out[2] != 'x' || out[3] != 'y' || out[4] != 0) return 4;
	if (nox_strcpy(out + 5, "ab") != out + 5) return 5;
	if (out[5] != 'a' || out[6] != 'b' || out[7] != 0) return 6;
	if (nox_strcat(out + 5, "cd") != out + 5) return 7;
	if (out[7] != 'c' || out[8] != 'd' || out[9] != 0) return 8;
	if (out[0] != 0x5a || out[15] != 0x5a) return 9;
	return 0;
}
*/
import "C"

import "unsafe"

func SafeMemoryResults(dst unsafe.Pointer) int {
	return int(C.safe_memory_results((*C.char)(dst)))
}
