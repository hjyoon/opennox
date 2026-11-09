#ifndef NOX_GAME1_MEM_DUMP_NATIVE_413E30_H
#define NOX_GAME1_MEM_DUMP_NATIVE_413E30_H
#include <stdint.h>

/* Native debug-list representation. The packed image's DWORD head is never
 * decoded as a host address on a 64-bit build. Scalar counters stay DWORDs. */
typedef struct nox_mem_dump_node_413e30 {
    void* reserved_0;
    struct nox_mem_dump_node_413e30* next;
    uint32_t reserved_8;
    uint32_t reserved_12;
    uint32_t count;
    char name[80];
} nox_mem_dump_node_413e30;
#endif
