#include <stdint.h>
#include "GAME1.h"

extern uint32_t sub_409B50_go(void* source);
_Static_assert(sizeof(unsigned int) == 4, "setter public unsigned ABI is32bits");
_Static_assert(_Generic(&sub_409B50, unsigned int (*)(const char*): 1, default: 0),
               "setter public const source pointer ABI must remain unchanged");

unsigned int sub_409B50(const char* source) {
    return (unsigned int)sub_409B50_go((void*)source);
}
