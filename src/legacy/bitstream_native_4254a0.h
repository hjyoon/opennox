#ifndef NOX_PORT_BITSTREAM_NATIVE_4254A0
#define NOX_PORT_BITSTREAM_NATIVE_4254A0

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

// GAME.EXE placed the bit index immediately after a PE32 buffer pointer.
// The index follows the complete native pointer in every current caller.
typedef struct nox_bit_cursor {
	uint8_t* data;
	uint8_t bit;
} nox_bit_cursor;

_Static_assert(offsetof(nox_bit_cursor, bit) == sizeof(void*), "native bit cursor index offset");
_Static_assert(sizeof(nox_bit_cursor) == 2 * sizeof(void*), "native bit cursor size");

int sub_4254A0(nox_bit_cursor* cursor, uint8_t* data);
bool sub_4254C0(nox_bit_cursor* cursor);
uint8_t* sub_425500(nox_bit_cursor* cursor, uint8_t* data, char value);
char sub_425520(nox_bit_cursor* cursor, char value);
int sub_425550(uint8_t* source, uint8_t* target, int bits);

#endif
