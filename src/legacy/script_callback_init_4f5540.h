#ifndef NOX_SCRIPT_CALLBACK_INIT_4F5540_H
#define NOX_SCRIPT_CALLBACK_INIT_4F5540_H

#include <stdint.h>

typedef struct nox_script_callback_t nox_script_callback_t;

_Static_assert(sizeof(int32_t) == 4,
	"script callback initializer result must remain an exact 32-bit value");

int32_t sub_4F5540(nox_script_callback_t* handler);

#endif // NOX_SCRIPT_CALLBACK_INIT_4F5540_H
