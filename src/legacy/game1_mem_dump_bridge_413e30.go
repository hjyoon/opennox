package legacy

/*
#include "GAME1.h"
#include <stdlib.h>
static void nox_mem_dump_sort_413e30(void* base, unsigned count) {
    qsort(base, count, 84, sub_413F60);
}
*/
import "C"
import "unsafe"

func memDumpSort413E30(base unsafe.Pointer, count int) {
	C.nox_mem_dump_sort_413e30(base, C.uint(count))
}
